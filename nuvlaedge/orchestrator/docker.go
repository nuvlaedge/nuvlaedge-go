package orchestrator

import (
	"context"
	"encoding/json"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/system"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/nuvlaedge/common"
	nuvlaTypes "nuvlaedge-go/nuvlaedge/types"
	"reflect"
	"sync"
	"time"
)

type ClusterData struct {
	NodeId             string
	NodeRole           string
	ClusterId          string
	ClusterManagers    []string
	ClusterWorkers     []string
	ClusterNodes       []string
	ClusterNodeLabels  []string
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
		log.Debugf("Updating swarm sw field: %s", updaterName)
		updater := reflect.ValueOf(sw).MethodByName(updaterName)
		log.Infof("Updater %s", updater.String())
		if updater.IsValid() {
			log.Debugf("Calling updater: %s", updaterName)
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

	log.Infof("Is docker Swarm available?: %s", ping.SwarmStatus.ControlAvailable)
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
	}
	dc.clusterData.ClusterNodes = nodeIds
	dc.clusterData.ClusterWorkers = workerIds

	// Gather cluster node labels
	dc.clusterData.ClusterNodeLabels = info.Labels
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

func (dc *DockerCoe) GetOrchestratorCredentials(attrs *nuvlaTypes.CommissioningAttributes) error {
	log.Debugf("Retrieving orchestrator credentials...")
	dc.swarmData.UpdateSwarmData()

	b, _ := json.MarshalIndent(dc.swarmData, "", "  ")
	log.Infof("Swarm data: %s", string(b))
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
