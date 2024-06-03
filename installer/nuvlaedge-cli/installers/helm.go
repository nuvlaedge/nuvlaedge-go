package installers

import (
	"context"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/helmpath"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/repo"
	certsv1 "k8s.io/api/certificates/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"nuvlaedge-cli/types"
	"os"
	"path/filepath"
	"sigs.k8s.io/yaml"
	"strings"
	"time"
)

const (
	HelmRepoName = "nuvlaedge"
	HelmRepoUrl  = "https://nuvlaedge.github.io/deployment"
)

type HelmInstaller struct {
	Uuid              string
	Version           string
	kubernetesVersion string
	NameSpace         string
	Flags             types.InstallFlags

	helmCli     *cli.EnvSettings
	action      *action.List
	repoManager *HelmRepoManager
}

func NewHelmInstaller(uuid, version string, flags *types.InstallFlags) *HelmInstaller {
	h := cli.New()
	h.KubeConfig = "/etc/rancher/k3s/k3s.yaml"
	return &HelmInstaller{
		Uuid:              uuid,
		Flags:             *flags,
		Version:           version,
		kubernetesVersion: version,
		NameSpace:         getNameSpaceFromUUID(uuid),
		helmCli:           h,
		repoManager:       NewHelmRepoManager(h),
	}
}

func (hi *HelmInstaller) Install() error {
	v, err := hi.isNuvlaEdgeInstalled()
	if err != nil {
		log.Infof("Cannot check if NuvlaEdge is already installed." +
			" To prevent errors and lose of data, nothing will be done")
		return err
	}

	if v != "" {
		log.Infof("NuvlaEdge is already installed on version %s", v)
		return nil
	}

	// Add NuvlaEdge repo if needed, and always update it
	if err = hi.addNuvlaEdgeRepo(); err != nil {
		log.Errorf("Error adding/updating NuvlaEdge repository: %v", err)
		return err
	}

	// Get Requested release
	releaseInstall := hi.GetRelease(hi.Version)
	log.Infof("Requested release: %s", releaseInstall)
	if releaseInstall == "" {
		return errors.New(fmt.Sprintf("could not find a suitable release for requested %s", hi.Version))
	}

	// Create action
	actionConf, _ := hi.createActionConf()
	installAction := action.NewInstall(actionConf)

	// Add configuration to the action
	installAction.ReleaseName = hi.NameSpace // TODO: Not sure what ReleaseName should be
	hi.helmCli.SetNamespace(hi.NameSpace)
	installAction.Namespace = hi.NameSpace
	installAction.ChartPathOptions.Version = releaseInstall

	// Get the chart
	chartRequested, err := hi.getInstallChart(installAction)
	if err != nil {
		return err
	}

	// Prepare install configuration
	conf, err := hi.getInstallConfiguration()
	if err != nil {
		return err
	}

	// Install the chart
	_, err = installAction.Run(chartRequested, conf)
	if err != nil {
		return err
	}

	// Wait and sign the certificate
	ctx := context.Background()
	err = hi.waitAndApproveCertificates(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (hi *HelmInstaller) getInstallChart(action *action.Install) (*chart.Chart, error) {
	cp, err := action.ChartPathOptions.LocateChart("nuvlaedge/nuvlaedge", hi.helmCli)
	if err != nil {
		log.Errorf("Error locating chart: %s", err)
		return nil, err
	}
	log.Infof("Chart path: %s", cp)

	// Load charts
	chartRequested, err := loader.Load(cp)
	if err != nil {
		log.Errorf("Error loading chart: %s", err)
		return nil, err
	}
	log.Infof("Chart requested: %s", chartRequested.Metadata.Name)

	return chartRequested, nil
}

func (hi *HelmInstaller) getInstallConfiguration() (map[string]interface{}, error) {
	// Extract values from ENV configuration
	p := getter.All(hi.helmCli)
	valueOpts := &values.Options{}

	// For the moment, assume the kubernetes node corresponds with the node name
	nodeName, _ := getNodeName()
	valueOpts.Values = []string{
		"NUVLAEDGE_UUID=" + hi.Uuid,
		"vpnClient=true",
		fmt.Sprintf("kubernetesNode=%s", nodeName)}

	vals, err := valueOpts.MergeValues(p)
	if err != nil {
		log.Errorf("Error merging values: %s", err)
		return nil, err
	}

	return vals, nil
}

// getNodeName returns the hostname of the machine and assumes is the name of the local K8s instance
// TODO: This should be extracted from Kubectl
func getNodeName() (string, error) {
	return os.Hostname()
}

// isCertificateSelfRequested checks if the certificate was requested by the helm service account
// Expects the next format: system:serviceaccount:<nuvlaedge_UUID>:nuvlaedge-service-account
func (hi *HelmInstaller) isCertificateSelfRequested(requester string) bool {
	fields := strings.Split(requester, ":")
	if len(fields) != 4 {
		log.Warnf("Requester format not recognized: %s", requester)
		return false
	}
	log.Infof("Requester: %s", fields[2])
	log.Infof("UUID: %s", hi.NameSpace)
	log.Infof("Expected result %t", hi.NameSpace == fields[2])
	return hi.NameSpace == fields[2]
}

func (hi *HelmInstaller) waitAndApproveCertificates(ctx context.Context) error {
	kubeconfig := "/etc/rancher/k3s/k3s.yaml" // replace with your kubeconfig file
	config, _ := clientcmd.BuildConfigFromFlags("", kubeconfig)
	clientset, _ := kubernetes.NewForConfig(config)

	timeoutCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			log.Errorf("Timeout waiting for certificate request")
			return errors.New("timeout waiting for certificate request")
		case <-ticker.C:
			csrs, _ := clientset.CertificatesV1().CertificateSigningRequests().List(context.Background(), metav1.ListOptions{})
			for _, csr := range csrs.Items {
				if hi.isCertificateSelfRequested(csr.Spec.Username) {
					log.Infof("Approving certificate request: %s", csr.Name)
					if err := hi.approveCertificateRequest(ctx, clientset, csr); err != nil {
						log.Errorf("Error approving certificate request: %s", err)
						return err
					}
					log.Infof("Certificate request approved")
					return nil
				}
			}

		}

	}
}

func (hi *HelmInstaller) approveCertificateRequest(ctx context.Context, clientset *kubernetes.Clientset, csrv certsv1.CertificateSigningRequest) error {
	csr, _ := clientset.CertificatesV1().CertificateSigningRequests().Get(ctx, csrv.Name, metav1.GetOptions{})
	csr.Status.Conditions = append(csr.Status.Conditions, certsv1.CertificateSigningRequestCondition{
		Type:    certsv1.CertificateApproved,
		Reason:  "KubectlApprove",
		Message: "This CSR was approved by kubectl certificate approve",
		Status:  corev1.ConditionTrue,
	})
	_, err := clientset.CertificatesV1().CertificateSigningRequests().UpdateApproval(ctx, csrv.Name, csr, metav1.UpdateOptions{})
	if err != nil {
		log.Errorf("Error approving certificate request: %s", err)
		return err
	}
	return nil
}
func (hi *HelmInstaller) Start() error {
	return nil
}

func (hi *HelmInstaller) Stop() error {
	return nil
}

func (hi *HelmInstaller) Remove() error {
	return nil
}

func (hi *HelmInstaller) Status() string {
	return ""
}

func (hi *HelmInstaller) String() string {
	return ""
}

func (hi *HelmInstaller) createActionConf() (*action.Configuration, error) {
	actionConf := new(action.Configuration)

	if err := actionConf.Init(hi.helmCli.RESTClientGetter(), hi.helmCli.Namespace(), os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
		log.Errorf("Error initializing helm action configuration: %s", err)
		os.Exit(1)
	}
	return actionConf, nil
}

// GetRelease checks if the requested release exists, else returns the latest. It also accepts "latest" or empty string
func (hi *HelmInstaller) GetRelease(requested string) string {
	path := filepath.Join(hi.helmCli.RepositoryCache, helmpath.CacheIndexFile("nuvlaedge"))
	log.Infof("Getting releases from %s", path)
	indexFile, err := repo.LoadIndexFile(path)
	if err != nil {
		log.Errorf("Error loading index file: %s", err)
		return ""
	}

	if requested == "" || requested == "latest" {
		latest := indexFile.Entries["nuvlaedge"][0].Metadata.AppVersion
		log.Infof("Getting latest version of NuvlaEdge: %s", latest)
		return latest
	}

	version := ""
	for name, versions := range indexFile.Entries["nuvlaedge"] {
		log.Infof("Name: %d - Versions: %v", name, versions.Metadata.AppVersion)
		if versions.Metadata.AppVersion == requested {
			version = requested
			log.Infof("Found version: %s", version)
			break
		}
	}
	return version
}

// isNuvlaEdgeInstalled checks if a nuvlaedge with the same UUID is already installed. Returns
func (hi *HelmInstaller) isNuvlaEdgeInstalled() (string, error) {
	releases, err := hi.listReleases()
	if err != nil {
		log.Errorf("Error listing releases: %s", err)
		return "", err
	}

	for _, r := range releases {
		if r.Name == hi.NameSpace || r.Namespace == hi.NameSpace {
			log.Infof("NuvlaEdge with UUID %s is already installed on version %s", hi.Uuid, r.Chart.Metadata.Version)
			return r.Chart.Metadata.Version, nil
		}

	}
	return "", nil
}

// listReleases lists the helm releases present in the local machine
func (hi *HelmInstaller) listReleases() ([]*release.Release, error) {

	actionConf, _ := hi.createActionConf()
	listAction := action.NewList(actionConf)
	listAction.Deployed = true

	results, err := listAction.Run()
	if err != nil {
		log.Errorf("Error listing releases: %s", err)
		return nil, err
	}

	return results, nil
}

func (hi *HelmInstaller) addNuvlaEdgeRepo() error {
	return hi.repoManager.AddRepo(HelmRepoName, HelmRepoUrl)
}

type HelmRepoManager struct {
	repoFile    string
	repoCache   string
	file        *repo.File
	envSettings *cli.EnvSettings
}

// NewHelmRepoManager creates a new HelmRepoManager. Accepts empty string paths
func NewHelmRepoManager(helmSettings *cli.EnvSettings) *HelmRepoManager {
	return &HelmRepoManager{
		repoFile:    helmSettings.RepositoryConfig,
		repoCache:   helmSettings.RepositoryCache,
		file:        repo.NewFile(),
		envSettings: helmSettings,
	}
}

func (hrm *HelmRepoManager) LoadFile() error {
	if _, err := os.Stat(hrm.repoFile); os.IsNotExist(err) {
		return fmt.Errorf("file %s does not exist", hrm.repoFile)
	}

	b, err := os.ReadFile(hrm.repoFile)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(b, hrm.file)
	if err != nil {
		return err
	}
	return nil
}

func (hrm *HelmRepoManager) SaveFile() error {
	return hrm.file.WriteFile(hrm.repoFile, 0600)
}

// updateRepo assumes file is preloaded and receives the Repo name
func (hrm *HelmRepoManager) updateRepo(name string) error {
	e := hrm.file.Get(name)
	if e == nil {
		return fmt.Errorf("repository %s not found", name)
	}

	r, err := repo.NewChartRepository(e, getter.All(hrm.envSettings))
	if err != nil {
		return err
	}

	if _, err = r.DownloadIndexFile(); err != nil {
		return err
	}
	return nil
}

func (hrm *HelmRepoManager) AddRepo(name, url string) error {
	if err := hrm.LoadFile(); err != nil {
		return err
	}

	c := repo.Entry{
		Name: name,
		URL:  url,
	}

	if hrm.RepoExistsAndIsEqual(&c) {
		log.Infof("Repository %s already exists and is up to date", name)
		return hrm.updateRepo(name)
	}

	r, err := repo.NewChartRepository(&c, getter.All(hrm.envSettings))
	if err != nil {
		return err
	}

	if _, err = r.DownloadIndexFile(); err != nil {
		return err
	}

	hrm.file.Update(&c)
	if err = hrm.SaveFile(); err != nil {
		return err
	}
	return nil
}

func (hrm *HelmRepoManager) RepoExistsAndIsEqual(r *repo.Entry) bool {
	if hrm.RepoExists(r.Name) {
		present := hrm.file.Get(r.Name)
		return *present == *r
	}
	return false
}

func (hrm *HelmRepoManager) RepoExists(name string) bool {
	return hrm.file.Has(name)
}

func getNameSpaceFromUUID(uuid string) string {
	return strings.Replace(uuid, "/", "-", -1)
}
