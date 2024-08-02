package types

type CommissionAttributes struct {
	Tags         []string `json:"tags,omitempty"`
	Capabilities []string `json:"capabilities,omitempty"`

	// VPN
	VpnCsr map[string]any `json:"vpn-csr,omitempty"`

	// Swarm
	SwarmEndPoint     string `json:"swarm-endpoint,omitempty"`
	SwarmTokenManager string `json:"swarm-token-manager,omitempty"`
	SwarmTokenWorker  string `json:"swarm-token-worker,omitempty"`
	SwarmClientKey    string `json:"swarm-client-key,omitempty"`
	SwarmClientCert   string `json:"swarm-client-cert,omitempty"`
	SwarmClientCa     string `json:"swarm-client-ca,omitempty"`

	// Minio
	MinioEndpoint  string `json:"minio-endpoint,omitempty"`
	MinioAccessKey string `json:"minio-access-key,omitempty"`
	MinioSecretKey string `json:"minio-secret-key,omitempty"`

	// Kubernetes
	KubernetesEndpoint   string `json:"kubernetes-endpoint,omitempty"`
	KubernetesClientKey  string `json:"kubernetes-client-key,omitempty"`
	KubernetesClientCert string `json:"kubernetes-client-cert,omitempty"`
	KubernetesClientCA   string `json:"kubernetes-client-ca,omitempty"`

	// Cluster data
	ClusterID           string   `json:"cluster-id,omitempty"`
	ClusterWorkerID     string   `json:"cluster-worker-id,omitempty"`
	ClusterOrchestrator string   `json:"cluster-orchestrator,omitempty"`
	ClusterManagers     []string `json:"cluster-managers,omitempty"`
	ClusterWorkers      []string `json:"cluster-workers,omitempty"`
}

type CommissionData interface {
	WriteToAttrs(attrs *CommissionAttributes) error
}
