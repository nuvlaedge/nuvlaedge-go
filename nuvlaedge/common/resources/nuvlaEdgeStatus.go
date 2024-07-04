package resources

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
	StatusNotes []string `json:"status-notes,omitempty"`

	// Telemetry
	Network   map[string]any `json:"network,omitempty"`
	Resources map[string]any `json:"resources,omitempty"`
}

type Vulnerabilities struct {
	Summary map[string]any   `json:"summary,omitempty"`
	Items   []map[string]any `json:"items,omitempty"`
}

type InstallationParameters struct {
	ProjectName string   `json:"project-name,omitempty"`
	Environment []string `json:"environment,omitempty"`
	WorkingDir  string   `json:"working-dir,omitempty"`
	ConfigFiles []string `json:"config-files,omitempty"`
}
