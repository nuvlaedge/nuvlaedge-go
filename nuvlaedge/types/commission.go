package types

type CommissioningAttributes struct {
	Tags         []string `json:"tags"`
	Capabilities []string `json:"capabilities"`

	// VPN
	VpnCsr map[string]any `json:"vpn-csr"`

	// Swarm
	SwarmEndPoint     string `json:"swarm-endpoint"`
	SwarmTokenManager string `json:"swarm-token-manager"`
	SwarmTokenWorker  string `json:"swarm-token-worker"`
	SwarmClientKey    string `json:"swarm-client-key"`
	SwarmClientCert   string `json:"swarm-client-cert"`
	SwarmClientCa     string `json:"swarm-client-ca"`

	// Minio
	MinioEndpoint  string `json:"minio-endpoint"`
	MinioAccessKey string `json:"minio-access-key"`
	MinioSecretKey string `json:"minio-secret-key"`

	// Kubernetes
	KubernetesEndpoint   string `json:"kubernetes-endpoint"`
	KubernetesClientKey  string `json:"kubernetes-client-key"`
	KubernetesClientCert string `json:"kubernetes-client-cert"`
	KubernetesClientCA   string `json:"kubernetes-client-ca"`

	// Cluster data
	ClusterID           string   `json:"cluster-id"`
	ClusterWorkerID     string   `json:"cluster-worker-id"`
	ClusterOrchestrator string   `json:"cluster-orchestrator"`
	ClusterManagers     []string `json:"cluster-managers"`
	ClusterWorkers      []string `json:"cluster-workers"`
}
