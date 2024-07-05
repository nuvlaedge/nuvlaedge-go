package orchestrator

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/system"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	"io"
	"nuvlaedge-go/nuvlaedge/common"
	"nuvlaedge-go/nuvlaedge/common/resources"
	"nuvlaedge-go/nuvlaedge/jobs/executors"
	neTypes "nuvlaedge-go/nuvlaedge/types"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"
)

const (
	ImagePullTimeout = 120 * time.Second
)

type ClusterData struct {
	NodeId             string
	NodeRole           string
	ClusterId          string
	ClusterManagers    []string
	ClusterWorkers     []string
	ClusterNodes       []string
	ClusterNodeLabels  []map[string]string
	ClusterJoinAddress string

	DockerServerVersion     string
	SwarmNodeCertExpiryDate string
	ContainerPlugins        []string
	updated                 time.Time
}

type SwarmData struct {
	SwarmEndPoint     string `json:"swarm-endpoint"`
	SwarmTokenManager string `json:"swarm-token-manager"`
	SwarmTokenWorker  string `json:"swarm-token-worker"`
	SwarmClientKey    string `json:"swarm-client-key"`
	SwarmClientCert   string `json:"swarm-client-cert"`
	SwarmClientCa     string `json:"swarm-client-ca"`

	updaters map[string]func() error
	client   *client.Client
}

func NewSwarmData(client *client.Client) *SwarmData {
	sw := &SwarmData{
		client:   client,
		updaters: make(map[string]func() error),
	}

	// Scan all fields of the struct and register the updaters
	swarmFields := reflect.ValueOf(*sw)

	for i := 0; i < swarmFields.NumField(); i++ {
		fieldName := swarmFields.Type().Field(i).Name
		updaterName := "Update" + fieldName
		updater := reflect.ValueOf(sw).MethodByName(updaterName)
		if updater.IsValid() {
			sw.updaters[fieldName] = updater.Interface().(func() error)
		}
	}
	return sw
}

func (sw *SwarmData) UpdateSwarmData() {
	var wg sync.WaitGroup

	wg.Add(len(sw.updaters))

	for _, updater := range sw.updaters {
		go func(updater func() error) {
			defer wg.Done()
			if err := updater(); err != nil {
				log.Errorf("Error updating swarm data: %s", err)
			}
		}(updater)
	}

	wg.Wait()
}

func (sw *SwarmData) UpdateSwarmDataIfNeeded() error {
	log.Infof("Updating swarm data")
	return nil
}

func (sw *SwarmData) UpdateSwarmEndPoint() error {
	log.Infof("Updating swarm endpoint")
	sw.SwarmEndPoint = "local"
	return nil
}

func (sw *SwarmData) UpdateSwarmTokenManager() error {
	log.Infof("Updating swarm token manager")
	ctx := context.Background()
	swarm, err := sw.client.SwarmInspect(ctx)
	if err != nil {
		return err
	}
	sw.SwarmTokenManager = swarm.JoinTokens.Manager
	return nil
}

func (sw *SwarmData) UpdateSwarmTokenWorker() error {
	log.Infof("Updating swarm token worker")
	ctx := context.Background()
	swarm, err := sw.client.SwarmInspect(ctx)
	if err != nil {
		return err
	}
	sw.SwarmTokenWorker = swarm.JoinTokens.Worker
	return nil
}

func (sw *SwarmData) UpdateSwarmClientKey() error {
	sw.SwarmClientKey = "null"
	return nil
}

func (sw *SwarmData) UpdateSwarmClientCert() error {
	sw.SwarmClientCert = "null"
	return nil
}

func (sw *SwarmData) UpdateSwarmClientCa() error {
	sw.SwarmClientCa = "null"
	return nil
}

type DockerCoe struct {
	coeType CoeType
	client  *client.Client

	clusterData     *ClusterData
	clusterDataLock *sync.Mutex
	swarmData       *SwarmData
}

func NewDockerCoe() *DockerCoe {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	common.GenericErrorHandler("Error instantiating client", err)
	ping, _ := cli.Ping(context.Background())

	if ping.SwarmStatus == nil {
		log.Infof("Swarm status is nil")
	} else {
		log.Infof("Swarm status: %s", ping.SwarmStatus.ControlAvailable)
	}

	return &DockerCoe{
		coeType: DockerType,
		client:  cli,
		clusterData: &ClusterData{
			updated: time.Now().Add(-10 * time.Second),
		},
		clusterDataLock: &sync.Mutex{},
		swarmData:       NewSwarmData(cli),
	}
}

/**************************************** NuvlaEdge Utils *****************************************/

func (dc *DockerCoe) GetCoeType() CoeType {
	return dc.coeType
}

// GetCoeVersion returns the version of the Docker client.
// It is a method of the DockerCoe struct.
func (dc *DockerCoe) GetCoeVersion() (string, error) {
	ctx := context.Background()
	version, err := dc.client.ServerVersion(ctx)
	if err != nil {
		return "", err
	}
	return version.Version, nil
}

func extractDockerPlugins(plugIns *system.PluginsInfo) []string {
	var plugins []string
	for _, plugin := range plugIns.Network {
		plugins = append(plugins, plugin)
	}
	for _, plugin := range plugIns.Volume {
		plugins = append(plugins, plugin)
	}
	for _, plugin := range plugIns.Authorization {
		plugins = append(plugins, plugin)
	}
	for _, plugin := range plugIns.Log {
		plugins = append(plugins, plugin)
	}
	return plugins
}

func (dc *DockerCoe) updateClusterData() error {
	log.Debug("Updating cluster data")

	ctx := context.Background()
	info, err := dc.client.Info(ctx)
	if err != nil {
		return err
	}
	// TODO: Check if Swarm is active
	if info.Swarm.LocalNodeState != "active" {
		log.Debugf("Swarm is not active: %s", info.Swarm.LocalNodeState)
		dc.clusterData.updated = time.Now()
		return nil
	}

	// Gather node ID
	dc.clusterData.NodeId = info.Swarm.NodeID
	// Gather cluster ID
	dc.clusterData.ClusterId = info.Swarm.Cluster.ID

	// Gather manager List
	var managerIds []string

	for _, manager := range info.Swarm.RemoteManagers {
		managerIds = append(managerIds, manager.NodeID)
		// Gather cluster join address if available
		if manager.NodeID == info.Swarm.NodeID {
			dc.clusterData.ClusterJoinAddress = manager.Addr
		}
	}
	dc.clusterData.ClusterManagers = managerIds

	// Gather cluster node List
	nodes, err := dc.client.NodeList(ctx, types.NodeListOptions{})
	if err != nil {
		return err
	}
	var nodeIds []string
	var workerIds []string
	for _, node := range nodes {
		nodeIds = append(nodeIds, node.ID)

		if node.Spec.Role == "worker" {
			workerIds = append(workerIds, node.ID)
		}
		if node.ID == info.Swarm.NodeID {
			// TODO: Probably best to assign node role like this
			dc.clusterData.NodeRole = string(node.Spec.Role)

			dc.clusterData.ClusterNodeLabels = make([]map[string]string, 0)
			for key, label := range node.Spec.Labels {
				dc.clusterData.ClusterNodeLabels =
					append(dc.clusterData.ClusterNodeLabels,
						map[string]string{"name": key, "value": label})
			}
		}
	}
	dc.clusterData.ClusterNodes = nodeIds
	dc.clusterData.ClusterWorkers = workerIds

	// Gather node role
	if info.Swarm.ControlAvailable {
		dc.clusterData.NodeRole = "manager"
	} else {
		dc.clusterData.NodeRole = "worker"
	}
	// Expiration date certificates
	// TODO: Extract expiration date calling openssl

	// Gather container plugins
	plugins, err := dc.client.PluginList(ctx, filters.Args{})
	var pluginsStr []string
	for _, plugin := range plugins {
		pluginsStr = append(pluginsStr, plugin.Name)
	}
	dc.clusterData.ContainerPlugins = pluginsStr

	dc.clusterData.updated = time.Now()
	return nil
}

func extractDockerSwarmCertExpiryDate() string {
	return ""
}

func (dc *DockerCoe) updateClusterDataIfNeeded() error {
	dc.clusterDataLock.Lock()
	defer dc.clusterDataLock.Unlock()
	if time.Since(dc.clusterData.updated) > 10*time.Second {
		log.Infof("Updating cluster data")
		return dc.updateClusterData()
	}
	return nil
}

func (dc *DockerCoe) GetClusterData() (*ClusterData, error) {
	err := dc.updateClusterDataIfNeeded()
	if err != nil {
		return nil, err
	}
	return dc.clusterData, nil
}

/**************************************** Swarm Data *****************************************/

func (dc *DockerCoe) GetOrchestratorCredentials(attrs *neTypes.CommissioningAttributes) error {
	log.Debugf("Retrieving orchestrator credentials...")
	dc.swarmData.UpdateSwarmData()

	b, _ := json.MarshalIndent(dc.swarmData, "", "  ")
	log.Debugf("Swarm data: %s", string(b))
	attrs.SwarmEndPoint = "local"
	attrs.SwarmTokenManager = dc.swarmData.SwarmTokenManager
	attrs.SwarmTokenWorker = dc.swarmData.SwarmTokenWorker
	attrs.SwarmClientKey = dc.swarmData.SwarmClientKey
	attrs.SwarmClientCert = dc.swarmData.SwarmClientCert
	attrs.SwarmClientCa = dc.swarmData.SwarmClientCa
	log.Debugf("Retrieving orchestrator credentials... Success.")
	return nil
}

func (dc *DockerCoe) GetSwarmData() (*SwarmData, error) {
	return nil, nil
}

/**************************************** Struct Utils *****************************************/

// String
func (dc *DockerCoe) String() string {
	return string(dc.GetCoeType())
}

/********************************* Docker container management functions *************************************/

func (dc *DockerCoe) RunContainer(image string, configuration map[string]string) (string, error) {
	//ctx := context.Background()

	return "", nil
}

type ImagePullResponse struct {
	Status         string `json:"status"`
	ProgressDetail struct {
		Current int `json:"current"`
		Total   int `json:"total"`
	} `json:"progressDetail"`
	Progress string `json:"progress"`
	ID       string `json:"id"`
}

func (dc *DockerCoe) pullAndWaitImage(ctx context.Context, imageName string) error {
	ctxTimed, cancel := context.WithTimeout(ctx, ImagePullTimeout)
	defer cancel()

	// Pull image
	r, err := dc.client.ImagePull(ctxTimed, imageName, image.PullOptions{})
	defer r.Close()
	if err != nil {
		return err
	}

	// Wait for image pull to complete
	_, err = io.Copy(io.Discard, r)
	if err != nil {
		return err
	}

	log.Infof("Successfully pulled image %s", imageName)
	return nil

}

func (dc *DockerCoe) GetInstallationParameters(parameters *resources.InstallationParameters) error {
	log.Infof("Reading installation parameters...")
	if executors.IsRunningOnHost() {
		log.Info("Reading parameters as Host")
		parameters.ConfigFiles = []string{"/bin/nuvlaedge"}
		parameters.ProjectName = "nuvlaedge"
		parameters.Environment = os.Environ()
		dir, err := os.Getwd()
		if err == nil {
			parameters.WorkingDir = dir
		}
	}

	if executors.IsRunningInDocker() {
		log.Debug("Reading installation parameters as Docker")

		parameters.Environment = os.Environ()
		pName := os.Getenv("COMPOSE_PROJECT_NAME")
		if pName == "" {
			return errors.New("COMPOSE_PROJECT_NAME not set")
		}
		containerName := fmt.Sprintf("%s-agent-go", pName)
		inspect, err := dc.client.ContainerInspect(context.Background(), containerName)
		if err != nil {
			return err
		}
		log.Infof("Inspecting container: %v", inspect.Config.Labels)
		parameters.ConfigFiles = strings.Split(inspect.Config.Labels["com.docker.compose.project.config_files"], ",")
		parameters.ProjectName = inspect.Config.Labels["com.docker.compose.project"]
		parameters.WorkingDir = inspect.Config.Labels["com.docker.compose.project.working_dir"]
	}

	return nil
}

func (dc *DockerCoe) RunJobEngineContainer(conf *neTypes.LegacyJobConf) (string, error) {
	if conf.Image == "" {
		conf.Image = common.JobEngineContainerImage
	}
	ctx := context.Background()
	// Pull image
	if err := dc.pullAndWaitImage(ctx, conf.Image); err != nil {
		return "", err
	}

	command := []string{"--", "/app/job_executor.py",
		"--api-url", conf.Endpoint,
		"--api-key", conf.ApiKey,
		"--api-secret", conf.ApiSecret,
		"--nuvlaedge-fs", "/tmp/nuvlaedge-fs",
		"--job-id", conf.JobId}
	if conf.EndpointInsecure {
		command = append(command, "--api-insecure")
	}

	envs := common.GetEnvironWithPrefix("NE_IMAGE_", "JOB_")
	log.Debugf("Passing envs: %v", envs)
	// Create container config
	config := &container.Config{
		Image:        conf.Image,
		Cmd:          command,
		AttachStderr: false,
		AttachStdout: false,
		AttachStdin:  false,
		Hostname:     conf.JobId,
		Env:          envs,
	}

	hostConf := &container.HostConfig{
		AutoRemove: true,
		Binds: []string{
			"/var/run/docker.sock:/var/run/docker.sock:rw", // Bind mount Docker socket
		},
	}

	resp, err := dc.client.ContainerCreate(
		ctx,
		config,
		hostConf,
		nil,
		nil,
		strings.Replace(conf.JobId, "/", "-", -1))
	if err != nil {
		log.Infof("Error creating container: %s", err)
		return "", err
	}
	log.Infof("Created container: %s, %v", resp.ID, resp.Warnings)

	err = dc.client.ContainerStart(ctx, resp.ID, container.StartOptions{})
	if err != nil {
		log.Infof("Error starting container: %s", err)
		return "", err
	}

	return resp.ID, nil

}

func (dc *DockerCoe) GetContainerLogs(containerId, since string) (io.ReadCloser, error) {
	logs, err := dc.client.ContainerLogs(
		context.Background(),
		containerId,
		container.LogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Timestamps: true,
			Since:      since,
		})

	if err != nil {
		return nil, err
	}
	return logs, nil
}

// printLogLines reads the log lines from the reader and prints them to the log.
// It returns the timestamp of the last log line read.
func printLogLines(reader io.ReadCloser) string {
	scanner := bufio.NewScanner(reader)
	var sinceTime string
	for scanner.Scan() {
		logLine := scanner.Text()
		log.Infof("Container log: %s", logLine)

		// Update the sinceTime to the timestamp of the current log line
		logParts := strings.SplitN(logLine, " ", 2)
		if len(logParts) > 0 {
			// Remove any non-timestamp characters from the start of the timestamp
			timestamp := strings.TrimLeft(logParts[0], "\x02\x00\x00\x00\x00\x00\x01\x1b")
			// Update the sinceTime to the timestamp of the current log line
			sinceTime = timestamp
		}
	}
	return sinceTime
}

func (dc *DockerCoe) printLogsUntilFinished(containerId string, exitFlag chan interface{}) {
	var sinceTime string
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-exitFlag:
			return
		case <-ticker.C:
			logs, err := dc.GetContainerLogs(containerId, sinceTime)
			if err != nil {
				log.Infof("Error getting logs: %s", err)
				return
			}
			sinceTime = printLogLines(logs)
			log.Infof("Container logs: %s", logs)
		}
	}
}

func (dc *DockerCoe) GetContainerStatus(containerId string) (string, error) {
	info, err := dc.client.ContainerInspect(context.Background(), containerId)
	if err != nil {
		return "", err
	}
	return info.State.Status, nil
}

func (dc *DockerCoe) WaitContainerFinish(containerId string, timeout time.Duration, printLogs bool) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if printLogs {
		exitFlag := make(chan interface{})
		go dc.printLogsUntilFinished(containerId, exitFlag)
		// TODO: Not sure if this is needed, or it is enough to close the channel
		defer func() {
			exitFlag <- struct{}{}
			close(exitFlag)
		}()
	}

	statusCh, errCh := dc.client.ContainerWait(ctx, containerId, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			log.Infof("Error waiting for container %s to finish: %s", containerId, err)
			return -1, err
		}
	case status := <-statusCh:
		log.Infof("Container %s finished with status: %d", containerId, status.StatusCode)
		return status.StatusCode, nil
	}
	return -1, nil
}

func (dc *DockerCoe) StopContainer(containerId string, force bool) (bool, error) {
	return false, nil
}

func (dc *DockerCoe) RemoveContainer(containerId string, containerName string) (bool, error) {
	return false, nil
}

/**************************************** NuvlaEdge Utils *****************************************/

// TelemetryStart Runs one iteration of the telemetry process related to the COE
func (dc *DockerCoe) TelemetryStart() error {
	return nil
}

func (dc *DockerCoe) TelemetryStatus() (int, error) {
	return 404, nil
}

func (dc *DockerCoe) TelemetryStop() (bool, error) {
	return false, nil
}

/**************************************** Docker Compose Management *****************************************/

func (dc *DockerCoe) RunCompose(composeFile string) error {
	return nil
}

func (dc *DockerCoe) StopCompose() error {
	return nil
}

func (dc *DockerCoe) RemoveCompose() error {
	return nil
}

func (dc *DockerCoe) GetComposeStatus() (string, error) {
	return "", nil
}
