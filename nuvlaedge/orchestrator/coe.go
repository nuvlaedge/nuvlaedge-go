package orchestrator

import (
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/nuvlaedge/types"
	"os"
)

type CoeType string

const (
	DockerType     CoeType = "swarm"
	KubernetesType CoeType = "kubernetes"
)

type Coe interface {
	GetCoeType() CoeType
	GetCoeVersion() (string, error)
	String() string

	RunContainer(image string, configuration map[string]string) (string, error)
	StopContainer(containerId string, force bool) (bool, error)
	RemoveContainer(containerId string, containerName string) (bool, error)

	GetClusterData() (*ClusterData, error)
	GetOrchestratorCredentials(*types.CommissioningAttributes) error

	TelemetryStart() error
	TelemetryStatus() (int, error)
	TelemetryStop() (bool, error)
}

func NewCoe(coeType CoeType) (Coe, error) {
	log.Infof("Creating new %s COE", coeType)
	switch coeType {
	case DockerType:
		return NewDockerCoe(), nil
	case KubernetesType:
		log.Errorf("Kubernetes COE not implemented yet")
		os.Exit(1)
	}
	log.Errorf("Unknown COE type: %s", coeType)
	os.Exit(1)
	return nil, nil
}
