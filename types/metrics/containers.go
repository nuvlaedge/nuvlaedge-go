package metrics

import (
	"cmp"
	"nuvlaedge-go/types"
	"slices"
	"strings"
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
}

func (c ClusterData) WriteToStatus(status *NuvlaEdgeStatus) error {
	status.NodeId = c.NodeId
	status.ClusterNodeRole = c.NodeRole
	status.ClusterId = c.ClusterId
	status.ClusterManagers = c.ClusterManagers

	status.Orchestrator = c.ClusterOrchestrator
	status.ClusterNodeLabels = c.ClusterNodeLabels
	status.ClusterJoinAddress = c.ClusterJoinAddress
	status.DockerServerVersion = c.DockerServerVersion
	status.SwarmNodeCertExpiryDate = c.SwarmNodeCertExpiryDate
	status.ContainerPlugins = c.ContainerPlugins
	return nil
}

func (c ClusterData) WriteToAttrs(attrs *types.CommissionAttributes) error {
	attrs.ClusterID = c.ClusterId
	attrs.ClusterManagers = c.ClusterManagers
	attrs.ClusterWorkers = c.ClusterWorkers
	attrs.ClusterOrchestrator = c.ClusterOrchestrator
	return nil
}

type SwarmData struct {
	SwarmEndPoint     string `json:"swarm-endpoint"`
	SwarmTokenManager string `json:"swarm-token-manager"`
	SwarmTokenWorker  string `json:"swarm-token-worker"`
	SwarmClientKey    string `json:"swarm-client-key"`
	SwarmClientCert   string `json:"swarm-client-cert"`
	SwarmClientCa     string `json:"swarm-client-ca"`
}

func (s SwarmData) WriteToAttrs(attrs *types.CommissionAttributes) error {
	attrs.SwarmEndPoint = s.SwarmEndPoint
	attrs.SwarmTokenManager = s.SwarmTokenManager
	attrs.SwarmTokenWorker = s.SwarmTokenWorker
	attrs.SwarmClientKey = s.SwarmClientKey
	attrs.SwarmClientCert = s.SwarmClientCert
	attrs.SwarmClientCa = s.SwarmClientCa
	return nil
}

type ContainerStats []ContainerData

func (cs ContainerStats) WriteToStatus(status *NuvlaEdgeStatus) error {
	slices.SortFunc(cs, func(a, b ContainerData) int {
		return cmp.Compare(strings.ToLower(a.CreatedAt), strings.ToLower(b.CreatedAt))
	})
	status.Resources.ContainerStats = cs
	return nil
}

type ContainerData struct {
	ContainerId     string  `json:"id,omitempty"`
	Name            string  `json:"name,omitempty"`
	ContainerStatus string  `json:"status,omitempty"`
	CpuUsage        float64 `json:"cpu-usage"`
	CpuLimit        float64 `json:"cpu-limit"`
	MemUsage        uint64  `json:"mem-usage"`
	MemLimit        uint64  `json:"mem-limit"`
	NetIn           uint64  `json:"net-in"`
	NetOut          uint64  `json:"net-out"`
	DiskIn          uint64  `json:"disk-in"`
	DiskOut         uint64  `json:"disk-out"`
	RestartCount    int     `json:"restart-count"`
	State           string  `json:"state,omitempty"`
	CreatedAt       string  `json:"created-at,omitempty"`
	Image           string  `json:"image,omitempty"`
}
