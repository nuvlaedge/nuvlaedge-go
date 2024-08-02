package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	nuvla "github.com/nuvla/api-client-go"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	neTypes "nuvlaedge-go/types"
	"nuvlaedge-go/types/metrics"
	"runtime"
	"strings"
	"time"
)

type DockerMonitor struct {
	BaseMonitor

	client           neTypes.DockerMetricsClient
	commissionerChan chan neTypes.CommissionData

	// metrics
	clusterData             metrics.ClusterData
	swarmData               metrics.SwarmData
	containersData          metrics.ContainerStats
	containerStatsSupported bool
}

// NewDockerMonitor creates a new DockerMonitor
func NewDockerMonitor(
	c neTypes.DockerMetricsClient,
	period int,
	repChan chan metrics.Metric,
	endpoint string,
	commChan chan neTypes.CommissionData) *DockerMonitor {

	dockerMonitor := &DockerMonitor{
		BaseMonitor:      NewBaseMonitor(period, repChan),
		client:           c,
		commissionerChan: commChan,
	}
	dockerMonitor.containerStatsSupported = checkSupportForContainerStats(endpoint)
	dockerMonitor.setDefaultSwarmData()

	return dockerMonitor
}

func (dm *DockerMonitor) Run(ctx context.Context) error {
	dm.SetRunning()
	defer dm.Close()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-dm.Ticker.C:
			// Send metric to channel
			if err := dm.updateMetrics(); err != nil {
				log.Errorf("Error updating Docker metrics: %s", err)
			}
			dm.sendMetrics()
		}
	}

}

func (dm *DockerMonitor) sendMetrics() {
	dm.reportChan <- dm.clusterData
	if dm.containerStatsSupported {
		dm.reportChan <- dm.containersData
	}
	dm.commissionerChan <- dm.clusterData
	dm.commissionerChan <- dm.swarmData
}

// setDefaultClusterData resets the structure and sets default values for the cluster data
func (dm *DockerMonitor) setDefaultSwarmData() {
	// Swarm endpoint is a deprecated feature. Since no longer is Swarm endpoint used to push containers in NuvlaEdge,
	// we can safely set it to "local"
	dm.swarmData = metrics.SwarmData{
		SwarmEndPoint:   "local",
		SwarmClientKey:  "null",
		SwarmClientCert: "null",
		SwarmClientCa:   "null",
	}
}

func (dm *DockerMonitor) updateSwarmData() error {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	swarm, err := dm.client.SwarmInspect(ctx)
	// TODO: Test if error is returned when Swarm is not present
	if err != nil {
		// We need to reset Swarm Data to avoid sending old data
		dm.setDefaultSwarmData()
		if strings.Contains(err.Error(), "This node is not a swarm manager") {
			log.Info("Node is not a swarm manager")
			return nil
		}
		return err
	}

	dm.swarmData.SwarmTokenManager = swarm.JoinTokens.Manager
	dm.swarmData.SwarmTokenWorker = swarm.JoinTokens.Worker

	return nil
}

func (dm *DockerMonitor) updateClusterData() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	info, err := dm.client.Info(ctx)

	if err != nil {
		log.Errorf("Error getting Docker info: %s", err)
		return err
	}
	dm.clusterData = metrics.ClusterData{}
	data := &dm.clusterData
	data.DockerServerVersion = info.ServerVersion

	if info.Swarm.LocalNodeState != "active" {
		// Docker swarm is not enabled, just return
		log.Info("Docker swarm is not enabled")
		return nil
	}

	data.ClusterOrchestrator = "swarm"

	// Cluster IDs
	data.NodeId = info.Swarm.NodeID
	data.ClusterId = info.Swarm.Cluster.ID

	// Gather manager List
	var managerIds []string
	for _, manager := range info.Swarm.RemoteManagers {
		managerIds = append(managerIds, manager.NodeID)
		// Gather cluster join address if available
		if manager.NodeID == info.Swarm.NodeID {
			data.ClusterJoinAddress = manager.Addr
		}
	}
	data.ClusterManagers = managerIds

	nodes, err := dm.client.NodeList(ctx, types.NodeListOptions{})
	if err != nil {
		log.Errorf("Error getting Docker swarm nodes: %s", err)
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
			data.NodeRole = string(node.Spec.Role)

			data.ClusterNodeLabels = make([]map[string]string, 0)
			for key, label := range node.Spec.Labels {
				data.ClusterNodeLabels =
					append(data.ClusterNodeLabels,
						map[string]string{"name": key, "value": label})
			}
		}
	}
	data.ClusterNodes = nodeIds
	data.ClusterWorkers = workerIds

	// Gather node role
	if info.Swarm.ControlAvailable {
		data.NodeRole = "manager"
	} else {
		data.NodeRole = "worker"
	}

	// Gather container plugins
	plugins, err := dm.client.PluginList(ctx, filters.Args{})
	if err == nil {
		var pluginsStr []string
		for _, plugin := range plugins {
			pluginsStr = append(pluginsStr, plugin.Name)
		}
		data.ContainerPlugins = pluginsStr
	}

	return nil
}

func (dm *DockerMonitor) updateContainersData() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	containers, err := dm.client.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		log.Errorf("Error getting Docker containers: %s", err)
		return err
	}

	dm.containersData = make([]metrics.ContainerData, 0)
	for _, c := range containers {
		// Retrieve container stats
		stats, err := dm.client.ContainerStats(ctx, c.ID, false)
		if err != nil {
			log.Infof("Error getting container stats for %s: %s", c.ID, err)
		}

		// Retrieve container inspect info
		inspect, err := dm.client.ContainerInspect(ctx, c.ID)
		if err != nil {
			log.Infof("Error inspecting container %s: %s", c.ID, err)
		}

		var stat = &container.StatsResponse{}
		decErr := json.NewDecoder(stats.Body).Decode(stat)

		err = stats.Body.Close()
		if err != nil {
			log.Errorf("Error closing stats reader: %s", err)
		}

		if decErr != nil {
			log.Errorf("Error decoding container stats: %s", err)
			stat = nil
		}

		d := NewContainerDataFromContainer(&c, stat, inspect)
		if err != nil {
			log.Errorf("Error creating container data: %s", err)
			continue
		}
		dm.containersData = append(dm.containersData, d)
	}
	return nil
}

func (dm *DockerMonitor) updateMetrics() error {
	var errs []error
	if err := dm.updateClusterData(); err != nil {
		errs = append(errs, err)
	}

	if err := dm.updateSwarmData(); err != nil {
		errs = append(errs, err)
	}

	if dm.containerStatsSupported {
		if err := dm.updateContainersData(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("errors updating metrics: %v", errs)
	}
	return nil
}

func (dm *DockerMonitor) Close() error {
	dm.runningLock.Lock()
	dm.running = false
	dm.runningLock.Unlock()
	if dm.Ticker != nil {
		dm.Ticker.Stop()
		dm.Ticker = nil
	}
	if err := dm.client.Close(); err != nil {
		log.Errorf("Error closing Docker client: %s", err)
		return err
	}
	return nil
}

func NewContainerDataFromContainer(
	info *types.Container,
	stat *container.StatsResponse,
	inspect types.ContainerJSON) metrics.ContainerData {

	var data metrics.ContainerData
	// Container Information
	data.ContainerId = info.ID
	data.Name = strings.TrimPrefix(info.Names[0], "/")
	data.Image = info.Image
	data.State = info.State
	data.ContainerStatus = info.Status
	data.CreatedAt = time.Unix(info.Created, 0).Format(time.RFC3339)

	if stat == nil {
		return data
	}

	// ContainerStats
	cpuUsage := stat.CPUStats.CPUUsage.TotalUsage - stat.PreCPUStats.CPUUsage.TotalUsage
	systemUsage := stat.CPUStats.SystemUsage - stat.PreCPUStats.SystemUsage

	cpuPercent := 0.0
	cpuLimit := float64(inspect.HostConfig.NanoCPUs) / 1_000_000_000
	if systemUsage != 0 {
		cpuPercent = (float64(cpuUsage) / float64(systemUsage)) * 100
	}

	var diskIn uint64
	var diskOut uint64

	if runtime.GOOS == "windows" {
		diskIn = stat.StorageStats.ReadSizeBytes
		diskOut = stat.StorageStats.WriteSizeBytes
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

	data.RestartCount = inspect.RestartCount
	data.CpuUsage = cpuPercent
	data.CpuLimit = cpuLimit
	data.MemUsage = stat.MemoryStats.Usage
	data.MemLimit = stat.MemoryStats.Limit
	data.NetIn = rxBytes
	data.NetOut = txBytes
	data.DiskIn = diskIn
	data.DiskOut = diskOut
	return data
}

func checkSupportForContainerStats(endpoint string) bool {
	resp, err := http.Get(nuvla.SanitiseEndpoint(endpoint) + "/api/resource-metadata/nuvlabox-status-2")
	if err != nil {
		log.Errorf("Error getting NuvlaEdgeStatus metadata: %s", err)
		return true
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Errorf("Error closing response body: %s", err)
		}
	}(resp.Body)

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Error reading response body: %s", err)
		return true
	}
	bodyString := string(bodyBytes)

	return strings.Contains(bodyString, "cpu-usage")
}
