package metrics

type NuvlaEdgeStatus struct {
	// Clustering configuration
	// Either Swarm or Kubernetes
	Orchestrator            string              `json:"orchestrator,omitempty"`
	NodeId                  string              `json:"node-id,omitempty"`
	ClusterId               string              `json:"cluster-id,omitempty"`
	ClusterManagers         []string            `json:"cluster-managers,omitempty"`
	ClusterNodes            []string            `json:"cluster-nodes,omitempty"`
	ClusterNodeLabels       []map[string]string `json:"cluster-node-labels,omitempty"`
	ClusterNodeRole         string              `json:"cluster-node-role,omitempty"`
	ClusterJoinAddress      string              `json:"cluster-join-address,omitempty"`
	SwarmNodeCertExpiryDate string              `json:"swarm-node-cert-expiry-date,omitempty"`
	CoeResources            *CoeResources       `json:"coe-resources,omitempty"`

	// Host System Settings
	Architecture        string `json:"architecture,omitempty"`
	OperatingSystem     string `json:"operating-system,omitempty"`
	IpV4Address         string `json:"ip,omitempty"`
	LastBoot            string `json:"last-boot,omitempty"`
	HostName            string `json:"hostname,omitempty"`
	DockerServerVersion string `json:"docker-server-version,omitempty"`

	// NuvlaEdge Configuration
	ContainerPlugins       []string                `json:"container-plugins,omitempty"`
	CurrentTime            string                  `json:"current-time,omitempty"`
	NuvlaEdgeEngineVersion string                  `json:"nuvlabox-engine-version,omitempty"`
	Version                int                     `json:"version,omitempty"`
	HostUserHome           string                  `json:"host-user-home,omitempty"`
	InstallationParameters *InstallationParameters `json:"installation-parameters,omitempty"` // Parameters that allow to relaunch the Nuvlaedge
	Components             []string                `json:"components,omitempty"`

	// NuvlaEdge report
	Status      string   `json:"status,omitempty"`
	StatusNotes []string `json:"status-notes"`

	// Telemetry
	Network   NetworkMetrics `json:"network,omitempty"`
	Resources Resources      `json:"resources,omitempty"`
}
