package metrics

type CoeResources struct {
	// Docker
	DockerResources     DockerResources     `json:"docker,omitempty"`
	KubernetesResources KubernetesResources `json:"kubernetes,omitempty"`
}

func (dr CoeResources) WriteToStatus(status *NuvlaEdgeStatus) error {
	status.CoeResources = &dr
	return nil
}

type DockerResources struct {
	// Docker
	Containers []map[string]interface{} `json:"containers"`
	Images     []map[string]interface{} `json:"images"`
	Volumes    []map[string]interface{} `json:"volumes"`
	Networks   []map[string]interface{} `json:"networks"`

	// Swarm
	Services []map[string]interface{} `json:"services"`
	Tasks    []map[string]interface{} `json:"tasks"`
	Configs  []map[string]interface{} `json:"configs"`
	Secrets  []map[string]interface{} `json:"secrets"`
}

type KubernetesResources struct {
}
