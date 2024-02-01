package resources

type NuvlaEdgeStatus struct {
	// Nuvla resource data
	//LastHeartbeat string `json:"last-heartbeat,omitempty"`
	//LastTelemetry string `json:"last-telemetry,omitempty"`
	//NextHeartbeat string `json:"next-heartbeat,omitempty"`
	//NextTelemetry string `json:"next-telemetry,omitempty"`
	//Online        bool   `json:"online,omitempty"`
	//
	//Updated      string `json:"updated,omitempty"`
	//Created      string `json:"created,omitempty"`
	//ResourceType string `json:"resource-type,omitempty"`
	//
	//// ids information
	//Id        string `json:"id,omitempty"`
	//Parent    string `json:"parent,omitempty"`
	//UpdatedBy string `json:"updated-by,omitempty"`
	//CreatedBy string `json:"created-by,omitempty"`

	// Basic configuration
	//Name        string `json:"name,omitempty"`
	//Description string `json:"description,omitempty"`

	// Clustering configuration
	// Either Swarm or Kubernetes
	Orchestrator            string   `json:"orchestrator,omitempty"`
	NodeId                  string   `json:"node-id,omitempty"`
	ClusterId               string   `json:"cluster-id,omitempty"`
	ClusterManagers         []string `json:"cluster-managers,omitempty"`
	ClusterNodes            []string `json:"cluster-nodes,omitempty"`
	ClusterNodeLabels       []string `json:"cluster-node-labels,omitempty"`
	ClusterNodeRole         string   `json:"cluster-node-role,omitempty"`
	ClusterJoinAddress      string   `json:"cluster-join-address,omitempty"`
	SwarmNodeCertExpiryDate string   `json:"swarm-node-cert-expiry-date,omitempty"`

	// Host System Settings
	Architecture        string `json:"architecture,omitempty"`
	OperatingSystem     string `json:"operating-system,omitempty"`
	IpV4Address         string `json:"ip,omitempty"`
	LastBoot            string `json:"last-boot,omitempty"`
	HostName            string `json:"hostname,omitempty"`
	DockerServerVersion string `json:"docker-server-version,omitempty"`

	// NuvlaEdge Configuration
	ContainerPlugins       []string       `json:"container-plugins,omitempty"`
	CurrentTime            string         `json:"current-time,omitempty"`
	NuvlaEdgeEngineVersion string         `json:"nuvlabox-engine-version,omitempty"`
	Version                int            `json:"version,omitempty"`
	HostUserHome           string         `json:"host-user-home,omitempty"`
	InstallationParameters map[string]any `json:"installation-parameters,omitempty"` // Parameters that allow to relaunch the Nuvlaedge
	Components             []string       `json:"components,omitempty"`

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
