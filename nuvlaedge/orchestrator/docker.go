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
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	ImagePullTimeout = 120 * time.Second
)

type ClusterData struct {
	NodeId              string
	NodeRole            string
	ClusterId           string
	ClusterManagers     []string
	ClusterWorkers      []string
	ClusterNodes        []string
	ClusterOrchestrator string
	ClusterNodeLabels   []map[string]string
	ClusterJoinAddress  string

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

	for k, updater := range sw.updaters {
		go func(updater func() error) {
			defer wg.Done()
			if err := updater(); err != nil {
				log.Errorf("[%s] Error updating swarm data: %s", k, err)
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
		sw.SwarmTokenManager = ""
		if strings.Contains(err.Error(), "This node is not a swarm manager.") {
			log.Infof("This node is not a swarm manager or swarm is disabled")
			return nil
		}
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
		sw.SwarmTokenWorker = ""
		if strings.Contains(err.Error(), "This node is not a swarm manager.") {
			log.Infof("This node is not a swarm manager or swarm is disabled")
			return nil
		}
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
		log.Infof("Swarm is not active: %s", info.Swarm.LocalNodeState)
		dc.clusterData = &ClusterData{}
		dc.clusterData.updated = time.Now()
		return nil
	}
	dc.clusterData.ClusterOrchestrator = string(dc.GetCoeType())

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

func (dc *DockerCoe) GetContainers(oldVersion bool) ([]interface{}, error) {
	containers, err := dc.client.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		return nil, err
	}
	var containerInfos []interface{}
	for _, containerInfo := range containers {
		if oldVersion {
			var containerOldStat resources.ContainerStatsOld
			containerOldStat.ContainerId = containerInfo.ID
			containerOldStat.Name = strings.TrimPrefix(containerInfo.Names[0], "/")
			containerOldStat.ContainerStatus = containerInfo.Status
			containerInfos = append(containerInfos, containerOldStat)
		} else {
			var containerNewStat resources.ContainerStatsNew
			containerNewStat.ContainerId = containerInfo.ID
			containerNewStat.Name = strings.TrimPrefix(containerInfo.Names[0], "/")
			containerNewStat.Image = containerInfo.Image
			containerNewStat.State = containerInfo.State
			containerNewStat.ContainerStatus = containerInfo.Status
			containerNewStat.CreatedAt = time.Unix(containerInfo.Created, 0).Format(time.RFC3339)
			containerInfos = append(containerInfos, containerNewStat)
		}
	}
	log.Debugf("Got the Containers %v", containerInfos)
	return containerInfos, nil
}

func (dc *DockerCoe) GetContainerStats(containerId string, statMap *interface{}) error {
	stats, err := dc.client.ContainerStats(context.Background(), containerId, false)
	if err != nil {
		return fmt.Errorf("error getting container stats from docker: %s", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Errorf("Error closing stats body: %s", err)
		}
	}(stats.Body)

	var stat types.StatsJSON
	err = json.NewDecoder(stats.Body).Decode(&stat)
	if err != nil {
		return err
	}

	inspect, err := dc.client.ContainerInspect(context.Background(), containerId)
	if err != nil {
		return fmt.Errorf("error inspecting container: %s", err)
	}

	cpuUsage := stat.CPUStats.CPUUsage.TotalUsage - stat.PreCPUStats.CPUUsage.TotalUsage
	systemUsage := stat.CPUStats.SystemUsage - stat.PreCPUStats.SystemUsage

	cpuPercent := 0.0
	cpulimit := float64(inspect.HostConfig.NanoCPUs) / 1_000_000_000
	if systemUsage != 0 {
		cpuPercent = (float64(cpuUsage) / float64(systemUsage)) * 100
	}

	var diskIn uint64 = 0
	var diskOut uint64 = 0

	if runtime.GOOS == "windows" {
		diskIn = stat.StorageStats.ReadSizeBytes
		diskOut = stat.StorageStats.WriteSizeBytes
		cpuUsage *= 100
	} else {
		var blkIn, blkOut uint64 = 0, 0
		for _, blkStat := range stat.BlkioStats.IoServiceBytesRecursive {
			if blkStat.Op == "Read" {
				blkIn += blkStat.Value
			} else if blkStat.Op == "Write" {
				blkOut += blkStat.Value
			}
		}
		diskIn = blkIn
		diskOut = blkOut
	}

	var rxBytes uint64 = 0
	var txBytes uint64 = 0

	for _, netStat := range stat.Networks {
		rxBytes += netStat.RxBytes
		txBytes += netStat.TxBytes
	}

	var memPercent float64 = 0.0
	if stat.MemoryStats.Limit != 0 {
		memPercent = (float64(stat.MemoryStats.Usage) / float64(stat.MemoryStats.Limit)) * 100
	}

	const FormatUsageLimit = "%d / %d"
	switch containerStats := (*statMap).(type) {
	case resources.ContainerStatsOld:
		containerStats.RestartCount = inspect.RestartCount
		containerStats.CpuPercent = fmt.Sprintf("%.2f", cpuPercent)
		containerStats.MemUsageLimit = fmt.Sprintf(FormatUsageLimit, stat.MemoryStats.Usage, stat.MemoryStats.Limit)
		containerStats.MemPercent = fmt.Sprintf("%.2f", memPercent)
		containerStats.NetInOut = fmt.Sprintf(FormatUsageLimit, rxBytes, txBytes)
		containerStats.BlkInOut = fmt.Sprintf(FormatUsageLimit, diskIn, diskOut)
		*statMap = containerStats
	case resources.ContainerStatsNew:
		containerStats.RestartCount = inspect.RestartCount
		containerStats.CpuUsage = cpuPercent
		containerStats.CpuLimit = cpulimit
		containerStats.MemUsage = stat.MemoryStats.Usage
		containerStats.MemLimit = stat.MemoryStats.Limit
		containerStats.NetIn = rxBytes
		containerStats.NetOut = txBytes
		containerStats.DiskIn = diskIn
		containerStats.DiskOut = diskOut
		*statMap = containerStats
	default:
		return fmt.Errorf("unknown container stats type: %T", containerStats)
	}

	return nil
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
