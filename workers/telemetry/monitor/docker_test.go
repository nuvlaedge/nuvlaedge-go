package monitor

import (
	"context"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/api/types/system"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io"
	"nuvlaedge-go/testutils"
	neTypes "nuvlaedge-go/types"
	"nuvlaedge-go/types/metrics"
	"strings"
	"sync"
	"testing"
	"time"
)

var mockChan chan metrics.Metric
var commChan chan neTypes.CommissionData

func init() {
	mockChan = make(chan metrics.Metric)
	commChan = make(chan neTypes.CommissionData)
	// Set log level to panic to avoid logs during tests
	log.SetLevel(log.PanicLevel)
}

func TestNewDockerMonitor(t *testing.T) {
	mockTestClient := testutils.TestDockerMetricsClient{}
	dockerMonitor := NewDockerMonitor(&mockTestClient, 10, mockChan, "https://nuvla.io", commChan)
	assert.NotNil(t, dockerMonitor, "DockerMonitor should not be nil")
	assert.Equal(t, 10, dockerMonitor.GetPeriod(), "DockerMonitor period should be 10")
}

func TestDockerMonitor_Run(t *testing.T) {
	mockTestClient := testutils.TestDockerMetricsClient{}
	dockerMonitor := NewDockerMonitor(&mockTestClient, 10, mockChan, "https://nuvla.io", commChan)
	ctx := context.Background()
	var err error
	go func() {
		err = dockerMonitor.Run(ctx)
	}()
	ctx.Done()

	assert.Nil(t, err, "DockerMonitor Run should not return error when stopped gracefully")

}

func TestDockerMonitor_sendMetrics(t *testing.T) {
	mockTestClient := testutils.TestDockerMetricsClient{}
	commChan = make(chan neTypes.CommissionData)
	mockChan = make(chan metrics.Metric)
	dockerMonitor := NewDockerMonitor(&mockTestClient, 10, mockChan, "https://nuvla.io", commChan)
	count := 0
	countComm := 0
	dockerMonitor.containerStatsSupported = true
	dockerMonitor.period = 1
	var wg sync.WaitGroup
	fmt.Printf("count %d\n", count)
	cxt, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	wg.Add(2)
	go func(ctx context.Context) {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case <-commChan:
				countComm++
				if countComm == 2 {
					return
				}

			}
		}
	}(cxt)

	go func(ctx context.Context) {
		defer wg.Done()
		for {
			select {
			case <-cxt.Done():
				return
			case <-mockChan:
				// A
				count++
				if count == 3 {
					return
				}
			}
		}
	}(cxt)
	time.Sleep(100 * time.Millisecond)
	dockerMonitor.sendMetrics()
	wg.Wait()
	assert.Equal(t, 2, countComm, "DockerMonitor sendMetrics should send 2 metrics to commChan")
	assert.Equal(t, 3, count, "DockerMonitor sendMetrics should send 3 metrics")
}

func TestDockerMonitor_GetChannel(t *testing.T) {
	mockTestClient := testutils.TestDockerMetricsClient{}
	dockerMonitor := NewDockerMonitor(&mockTestClient, 10, mockChan, "https://nuvla.io", commChan)
	assert.NotNil(t, dockerMonitor.GetChannel(), "DockerMonitor GetChannel should not be nil")
}

func Test_DockerMonitor_setDefaultSwarmData(t *testing.T) {
	mockTestClient := testutils.TestDockerMetricsClient{}
	dockerMonitor := NewDockerMonitor(&mockTestClient, 10, mockChan, "https://nuvla.io", commChan)
	dockerMonitor.setDefaultSwarmData()
	assert.NotNil(t, dockerMonitor.swarmData, "DockerMonitor swarmData should not be nil")
	assert.Equal(t, "local", dockerMonitor.swarmData.SwarmEndPoint, "DockerMonitor swarmData.Nodes should be 0")
	assert.Equal(t, "null", dockerMonitor.swarmData.SwarmClientKey, "DockerMonitor swarmData.Nodes should be 0")
	assert.Equal(t, "null", dockerMonitor.swarmData.SwarmClientCert, "DockerMonitor swarmData.Nodes should be 0")
}

func Test_DockerMonitor_UpdateSwarmData(t *testing.T) {
	mockTestClient := testutils.TestDockerMetricsClient{}
	mockTestClient.InspectReturn = swarm.Swarm{
		JoinTokens: swarm.JoinTokens{
			Worker:  "workerToken",
			Manager: "managerToken",
		},
	}
	dockerMonitor := NewDockerMonitor(&mockTestClient, 10, mockChan, "https://nuvla.io", commChan)
	err := dockerMonitor.updateSwarmData()
	assert.Nil(t, err, "updateSwarmData should not return an error for active swarm")
	assert.Equal(t, "workerToken", dockerMonitor.swarmData.SwarmTokenWorker, "SwarmTokenWorker should be set")
	assert.Equal(t, "managerToken", dockerMonitor.swarmData.SwarmTokenManager, "SwarmTokenManager should be set")

	mockTestClient = testutils.TestDockerMetricsClient{}
	mockTestClient.InspectErr = errors.New("swarm not active")
	dockerMonitor = NewDockerMonitor(&mockTestClient, 10, mockChan, "https://nuvla.io", commChan)
	err = dockerMonitor.updateSwarmData()
	assert.NotNil(t, err, "updateSwarmData should return an error for inactive swarm")
	assert.Equal(t, "local", dockerMonitor.swarmData.SwarmEndPoint, "SwarmEndPoint should be reset to 'local'")
	assert.Equal(t, "null", dockerMonitor.swarmData.SwarmClientKey, "SwarmClientKey should be reset to 'null'")
	assert.Equal(t, "null", dockerMonitor.swarmData.SwarmClientCert, "SwarmClientCert should be reset to 'null'")

	mockTestClient = testutils.TestDockerMetricsClient{}
	mockTestClient.InspectErr = errors.New("error inspecting swarm")
	dockerMonitor = NewDockerMonitor(&mockTestClient, 10, mockChan, "https://nuvla.io", commChan)
	err = dockerMonitor.updateSwarmData()
	assert.NotNil(t, err, "updateSwarmData should return an error when inspecting swarm fails")
}

func TestUpdateClusterData_SwarmDisabled_SetsEmptyClusterData(t *testing.T) {
	mockTestClient := testutils.TestDockerMetricsClient{}
	mockTestClient.InfoReturn = system.Info{
		Swarm: swarm.Info{
			LocalNodeState: "inactive",
		},
	}
	dockerMonitor := NewDockerMonitor(&mockTestClient, 10, mockChan, "https://nuvla.io", commChan)
	err := dockerMonitor.updateClusterData()
	assert.Nil(t, err, "updateClusterData should not return an error when swarm is disabled")
	assert.Empty(t, dockerMonitor.clusterData, "ClusterData should be empty when swarm is disabled")

	mockTestClient = testutils.TestDockerMetricsClient{}
	mockTestClient.InfoErr = errors.New("failed to get Docker info")
	dockerMonitor = NewDockerMonitor(&mockTestClient, 10, mockChan, "https://nuvla.io", commChan)
	err = dockerMonitor.updateClusterData()
	assert.NotNil(t, err, "updateClusterData should return an error when Docker info retrieval fails")
}

func TestUpdateClusterData_SwarmEnabledButNoNodes_ReturnsClusterDataWithoutNodes(t *testing.T) {
	mockTestClient := testutils.TestDockerMetricsClient{}
	mockTestClient.InfoReturn = system.Info{
		ServerVersion: "20.10",
		Swarm: swarm.Info{
			LocalNodeState: "active",
			NodeID:         "managerNodeID",
			Cluster: &swarm.ClusterInfo{
				ID: "clusterID",
			},
			ControlAvailable: true,
			RemoteManagers: []swarm.Peer{
				{
					NodeID: "managerNodeID",
					Addr:   "10.0.0.1:2377",
				},
			},
		},
	}
	mockTestClient.NodeListReturn = []swarm.Node{}
	dockerMonitor := NewDockerMonitor(&mockTestClient, 10, mockChan, "https://nuvla.io", commChan)
	err := dockerMonitor.updateClusterData()
	assert.Nil(t, err, "updateClusterData should not return an error when swarm is enabled but no nodes are present")
	assert.Empty(t, dockerMonitor.clusterData.ClusterNodes, "ClusterNodes should be empty when no nodes are present")
	assert.Empty(t, dockerMonitor.clusterData.ClusterWorkers, "ClusterWorkers should be empty when no nodes are present")
}

func TestUpdateClusterData_SwarmEnabledWithNodes_SetsClusterAndNodeData(t *testing.T) {
	mockTestClient := testutils.TestDockerMetricsClient{}
	mockTestClient.InfoReturn = system.Info{
		ServerVersion: "20.10",
		Swarm: swarm.Info{
			LocalNodeState: "active",
			NodeID:         "managerNodeID",
			Cluster: &swarm.ClusterInfo{
				ID: "clusterID",
			},
			ControlAvailable: true,
			RemoteManagers: []swarm.Peer{
				{
					NodeID: "managerNodeID",
					Addr:   "10.0.0.1:2377",
				},
			},
		},
	}
	nodeSpect := swarm.NodeSpec{}
	nodeSpect.Role = "manager"
	nodeSpect.Labels = map[string]string{
		"role": "manager",
	}
	mockTestClient.NodeListReturn = []swarm.Node{
		{
			ID:   "managerNodeID",
			Spec: nodeSpect,
		},
		{
			ID: "workerNodeID",
			Spec: swarm.NodeSpec{
				Role: "worker",
			},
		},
	}
	dockerMonitor := NewDockerMonitor(&mockTestClient, 10, mockChan, "https://nuvla.io", commChan)
	err := dockerMonitor.updateClusterData()
	assert.Nil(t, err, "updateClusterData should not return an error when swarm is enabled with nodes")
	assert.NotEmpty(t, dockerMonitor.clusterData.ClusterNodes, "ClusterNodes should not be empty when nodes are present")
	assert.NotEmpty(t, dockerMonitor.clusterData.ClusterWorkers, "ClusterWorkers should not be empty when worker nodes are present")
	assert.Equal(t, "manager", dockerMonitor.clusterData.NodeRole, "NodeRole should be set correctly for manager node")
}

func TestUpdateClusterData_PluginsListedSuccessfully_SetsContainerPluginsCorrectly(t *testing.T) {
	mockTestClient := testutils.TestDockerMetricsClient{}
	mockTestClient.InfoReturn = system.Info{
		Swarm: swarm.Info{
			LocalNodeState: "active",
			Cluster: &swarm.ClusterInfo{
				ID: "clusterID",
			},
		},
	}
	mockTestClient.PluginListReturn = types.PluginsListResponse{
		{Name: "plugin1"},
		{Name: "plugin2"},
	}
	dockerMonitor := NewDockerMonitor(&mockTestClient, 10, mockChan, "https://nuvla.io", commChan)
	err := dockerMonitor.updateClusterData()
	assert.Nil(t, err, "updateClusterData should not return an error when plugins are listed successfully")
	assert.ElementsMatch(t, []string{"plugin1", "plugin2"}, dockerMonitor.clusterData.ContainerPlugins, "ContainerPlugins should be set correctly with plugin names")
}

func TestUpdateClusterData_PluginListError_HandlesErrorGracefully(t *testing.T) {
	mockTestClient := testutils.TestDockerMetricsClient{}
	mockTestClient.InfoReturn = system.Info{
		Swarm: swarm.Info{
			LocalNodeState: "active",
			Cluster: &swarm.ClusterInfo{
				ID: "clusterID",
			},
		},
	}
	mockTestClient.PluginListErr = errors.New("error listing plugins")
	dockerMonitor := NewDockerMonitor(&mockTestClient, 10, mockChan, "https://nuvla.io", commChan)
	err := dockerMonitor.updateClusterData()
	assert.Nil(t, err, "updateClusterData should handle plugin list retrieval error gracefully")
	assert.Empty(t, dockerMonitor.clusterData.ContainerPlugins, "ContainerPlugins should be empty when there is an error listing plugins")
}

func TestUpdateContainersData_Integrated(t *testing.T) {
	mockTestClient := testutils.TestDockerMetricsClient{}

	// Mocking container list to return two containers
	mockTestClient.ContainerListReturn = []types.Container{
		{
			ID:      "container1",
			Names:   []string{"/container_one"},
			Image:   "image1",
			State:   "running",
			Status:  "Up 24 hours",
			Created: 1622540800,
		},
		{
			ID:      "container2",
			Names:   []string{"/container_two"},
			Image:   "image2",
			State:   "exited",
			Status:  "Exited (0) 2 hours ago",
			Created: 1622637200,
		},
	}

	// Mocking ContainerStats to return valid stats for the first container and an error for the second
	mockTestClient.ContainerStatsReturn = container.StatsResponseReader{
		Body: io.NopCloser(strings.NewReader(`{
            "cpu_stats": {"cpu_usage": {"total_usage": 100}},
            "memory_stats": {"usage": 200},
            "networks": {"eth0": {"rx_bytes": 100, "tx_bytes": 200}},
            "blkio_stats": {"io_service_bytes_recursive": [{"op": "Read", "value": 100}, {"op": "Write", "value": 200}]}
        }`)),
	}

	// Mocking ContainerInspect to return valid inspect data for the first container and nil for the second
	hostConf := &container.HostConfig{}
	hostConf.NanoCPUs = 1000000000
	mockTestClient.ContainerInspectReturn = types.ContainerJSON{
		ContainerJSONBase: &types.ContainerJSONBase{
			ID:         "container1",
			HostConfig: hostConf,
		},
		Config: &container.Config{
			Image: "image1",
		},
	}
	mockTestClient.ContainerInspectErr = nil

	//// Setting up a condition to simulate an error for the second container's inspect data
	//mockTestClient.ContainerInspectErrFunc = func(containerID string) error {
	//	if containerID == "container2" {
	//		return errors.New("inspect data not available")
	//	}
	//	return nil
	//}

	dockerMonitor := NewDockerMonitor(&mockTestClient, 10, mockChan, "https://nuvla.io", commChan)
	err := dockerMonitor.updateContainersData()
	assert.Nil(t, err, "updateContainersData should not return an error")

	// Asserting that the containersData has been populated correctly for the first container
	// and handled gracefully for the second container where inspect data was nil
	assert.Len(t, dockerMonitor.containersData, 2, "containersData should contain data for two containers")
	assert.Equal(t, "container_one", dockerMonitor.containersData[0].Name, "First container's name should be set correctly")
}

func TestUpdateMetrics_AllUpdatesSucceed_ReturnsNoError(t *testing.T) {
	mockTestClient := testutils.TestDockerMetricsClient{}
	mockTestClient.InfoReturn = system.Info{
		ServerVersion: "20.10",
		Swarm: swarm.Info{
			LocalNodeState: "active",
			NodeID:         "managerNodeID",
			Cluster: &swarm.ClusterInfo{
				ID: "clusterID",
			},
			ControlAvailable: true,
			RemoteManagers: []swarm.Peer{
				{
					NodeID: "managerNodeID",
					Addr:   "10.0.0.1:2377",
				},
			},
		},
	}
	mockTestClient.NodeListReturn = []swarm.Node{
		{
			ID: "managerNodeID",
			Spec: swarm.NodeSpec{
				Role: "manager",
			},
		},
	}
	mockTestClient.ContainerStatsReturn = container.StatsResponseReader{
		Body: io.NopCloser(strings.NewReader(`{
            "cpu_stats": {"cpu_usage": {"total_usage": 100}},
            "memory_stats": {"usage": 200},
            "networks": {"eth0": {"rx_bytes": 100, "tx_bytes": 200}},
            "blkio_stats": {"io_service_bytes_recursive": [{"op": "Read", "value": 100}, {"op": "Write", "value": 200}]}
        }`)),
	}
	mockTestClient.PluginListReturn = types.PluginsListResponse{
		{Name: "plugin1"},
		{Name: "plugin2"},
	}
	mockTestClient.ContainerListReturn = []types.Container{
		{
			ID:      "container1",
			Names:   []string{"/container_one"},
			Image:   "image1",
			State:   "running",
			Status:  "Up 24 hours",
			Created: 1622540800,
		},
	}
	hostConf := &container.HostConfig{}
	hostConf.NanoCPUs = 1000000000
	mockTestClient.ContainerInspectReturn = types.ContainerJSON{
		ContainerJSONBase: &types.ContainerJSONBase{
			ID:         "container1",
			HostConfig: hostConf,
		},
		Config: &container.Config{
			Image: "image1",
		},
	}
	dockerMonitor := NewDockerMonitor(&mockTestClient, 10, mockChan, "https://nuvla.io", commChan)
	err := dockerMonitor.updateMetrics()
	assert.Nil(t, err, "updateMetrics should not return an error when all updates succeed")
}

func TestUpdateMetrics_ClusterDataUpdateFails_ReturnsError(t *testing.T) {
	mockTestClient := testutils.TestDockerMetricsClient{}
	mockTestClient.InfoErr = errors.New("failed to get Docker info")

	dockerMonitor := NewDockerMonitor(&mockTestClient, 10, mockChan, "https://nuvla.io", commChan)
	err := dockerMonitor.updateMetrics()
	assert.NotNil(t, err, "updateMetrics should return an error when cluster data update fails")

	mockTestClient = testutils.TestDockerMetricsClient{}
	mockTestClient.InspectErr = errors.New("error inspecting swarm")

	dockerMonitor = NewDockerMonitor(&mockTestClient, 10, mockChan, "https://nuvla.io", commChan)
	err = dockerMonitor.updateMetrics()
	assert.NotNil(t, err, "updateMetrics should return an error when swarm data update fails")

	mockTestClient = testutils.TestDockerMetricsClient{}
	mockTestClient.InfoErr = errors.New("failed to get Docker info")
	mockTestClient.InspectErr = errors.New("error inspecting swarm")
	mockTestClient.ContainerListErr = errors.New("error listing containers")

	dockerMonitor = NewDockerMonitor(&mockTestClient, 10, mockChan, "https://nuvla.io", commChan)
	err = dockerMonitor.updateMetrics()
	assert.NotNil(t, err, "updateMetrics should return an aggregated error when multiple updates fail")
	assert.Contains(t, err.Error(), "failed to get Docker info", "Error message should include failure to get Docker info")
	assert.Contains(t, err.Error(), "error inspecting swarm", "Error message should include error inspecting swarm")
}

func TestCloseSucceeds_ReportsNoError(t *testing.T) {
	mockTestClient := testutils.TestDockerMetricsClient{}
	mockTestClient.CloseErr = nil
	dockerMonitor := NewDockerMonitor(&mockTestClient, 10, mockChan, "https://nuvla.io", commChan)
	dockerMonitor.Close()
	assert.Equal(t, 1, mockTestClient.CloseCount, "Close should be called once")
	assert.False(t, dockerMonitor.running, "dockerMonitor should be stopped", "DockerMonitor should be stopped")
	assert.Nil(t, dockerMonitor.Ticker, "Ticker should be nil")
}
