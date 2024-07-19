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
	StatusNotes []string `json:"status-notes"`

	// Telemetry
	Network   map[string]any `json:"network,omitempty"`
	Resources map[string]any `json:"resources,omitempty"`

	// Container Stats
	ContainerStats []any `json:"container-stats,omitempty"`
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

type ContainerStatsOld struct {
	ContainerId     string  `json:"id,omitempty"`
	Name            string  `json:"name,omitempty"`
	ContainerStatus string  `json:"container-status,omitempty"`
	CpuPercent      string  `json:"cpu-percent,omitempty"`
	MemUsageLimit   string  `json:"mem-usage-limit,omitempty"`
	MemPercent      string  `json:"mem-percent,omitempty"`
	NetInOut        string  `json:"net-in-out,omitempty"`
	BlkInOut        string  `json:"blk-in-out,omitempty"`
	RestartCount    int     `json:"restart-count"`
}

type ContainerStatsNew struct {
	ContainerId     string  `json:"id,omitempty"`
	Name            string  `json:"name,omitempty"`
	ContainerStatus string  `json:"status,omitempty"`
	CpuUsage        float64 `json:"cpu-usage"`
	CpuLimit        float64 `json:"cpu-limit,omitempty"`
	CpuCapacity     uint32  `json:"cpu-capacity,omitempty"`
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
