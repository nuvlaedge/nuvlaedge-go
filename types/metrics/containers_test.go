package metrics

import (
	"github.com/stretchr/testify/assert"
	"nuvlaedge-go/types"
	"testing"
)

func Test_ClusterData_WriteToStatus(t *testing.T) {
	c := ClusterData{
		NodeId:                  "node-id",
		NodeRole:                "node-role",
		ClusterId:               "cluster-id",
		ClusterManagers:         []string{"manager1", "manager2"},
		ClusterWorkers:          []string{"worker1", "worker2"},
		ClusterNodes:            []string{"node1", "node2"},
		ClusterOrchestrator:     "orchestrator",
		ClusterNodeLabels:       []map[string]string{{"key1": "value1"}, {"key2": "value2"}},
		ClusterJoinAddress:      "join-address",
		DockerServerVersion:     "server-version",
		SwarmNodeCertExpiryDate: "expiry-date",
		ContainerPlugins:        []string{"plugin1", "plugin2"},
	}

	status := &NuvlaEdgeStatus{}
	err := c.WriteToStatus(status)
	assert.NoErrorf(t, err, "error writing cluster data to status")
	assert.Equal(t, c.NodeId, status.NodeId, "node id not set correctly")
	assert.Equal(t, c.NodeRole, status.ClusterNodeRole, "node role not set correctly")
	assert.Equal(t, c.ClusterId, status.ClusterId, "cluster id not set correctly")
	assert.Equal(t, c.ClusterManagers, status.ClusterManagers, "cluster managers not set correctly")
	assert.Equal(t, c.ClusterOrchestrator, status.Orchestrator, "orchestrator not set correctly")
	assert.Equal(t, c.ClusterNodeLabels, status.ClusterNodeLabels, "node labels not set correctly")
	assert.Equal(t, c.ClusterJoinAddress, status.ClusterJoinAddress, "join address not set correctly")
	assert.Equal(t, c.DockerServerVersion, status.DockerServerVersion, "server version not set correctly")
	assert.Equal(t, c.SwarmNodeCertExpiryDate, status.SwarmNodeCertExpiryDate, "cert expiry date not set correctly")
	assert.Equal(t, c.ContainerPlugins, status.ContainerPlugins, "container plugins not set correctly")
}

func Test_ClusterData_WriteToAttrs(t *testing.T) {
	c := ClusterData{
		ClusterId:           "cluster-id",
		ClusterManagers:     []string{"manager1", "manager2"},
		ClusterWorkers:      []string{"worker1", "worker2"},
		ClusterOrchestrator: "orchestrator",
	}
	data := &types.CommissionAttributes{}
	err := c.WriteToAttrs(data)
	assert.NoErrorf(t, err, "error writing cluster data to commission attributes")
	assert.Equal(t, c.ClusterId, data.ClusterID, "cluster id not set correctly")
	assert.Equal(t, c.ClusterManagers, data.ClusterManagers, "cluster managers not set correctly")
	assert.Equal(t, c.ClusterWorkers, data.ClusterWorkers, "cluster workers not set correctly")
	assert.Equal(t, c.ClusterOrchestrator, data.ClusterOrchestrator, "cluster orchestrator not set correctly")
}

func Test_SwarmData_WriteToAttrs(t *testing.T) {
	s := SwarmData{
		SwarmEndPoint:     "endpoint",
		SwarmTokenManager: "token-manager",
		SwarmTokenWorker:  "token-worker",
		SwarmClientKey:    "client-key",
		SwarmClientCert:   "client-cert",
		SwarmClientCa:     "client-ca",
	}
	data := &types.CommissionAttributes{}
	err := s.WriteToAttrs(data)
	assert.NoErrorf(t, err, "error writing swarm data to commission attributes")
	assert.Equal(t, s.SwarmEndPoint, data.SwarmEndPoint, "swarm endpoint not set correctly")
	assert.Equal(t, s.SwarmTokenManager, data.SwarmTokenManager, "swarm token manager not set correctly")
	assert.Equal(t, s.SwarmTokenWorker, data.SwarmTokenWorker, "swarm token worker not set correctly")
	assert.Equal(t, s.SwarmClientKey, data.SwarmClientKey, "swarm client key not set correctly")
	assert.Equal(t, s.SwarmClientCert, data.SwarmClientCert, "swarm client cert not set correctly")
	assert.Equal(t, s.SwarmClientCa, data.SwarmClientCa, "swarm client ca not set correctly")
}

func Test_ContainerStats_WriteToStatus(t *testing.T) {
	c := ContainerData{
		ContainerId: "container-id",
		Name:        "container-name",
		Image:       "container-image",
		CpuUsage:    0.1,
		MemUsage:    139,
	}
	cs := ContainerStats{c}

	status := &NuvlaEdgeStatus{}
	err := cs.WriteToStatus(status)
	assert.NoErrorf(t, err, "error writing container stats to status")
	assert.Equal(t, c.ContainerId, status.Resources.ContainerStats[0].ContainerId, "container id not set correctly")
	assert.Equal(t, c.Name, status.Resources.ContainerStats[0].Name, "container name not set correctly")
	assert.Equal(t, c.Image, status.Resources.ContainerStats[0].Image, "container image not set correctly")
	assert.Equal(t, c.CpuUsage, status.Resources.ContainerStats[0].CpuUsage, "container cpu usage not set correctly")
	assert.Equal(t, c.MemUsage, status.Resources.ContainerStats[0].MemUsage, "container memory usage not set correctly")
}
