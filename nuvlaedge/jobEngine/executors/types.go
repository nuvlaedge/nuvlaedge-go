package executors

import (
	"fmt"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/nuvla/api-client-go/clients/resources"
	"nuvlaedge-go/nuvlaedge/errors"
)

// Rebooter is an interface for executors that can reboot the system.
type Rebooter interface {
	Executor
	Reboot() error
}

// GetRebooter returns the appropriate Rebooter for the current environment. Question, should we trigger K8s or docker
// reboot if host is not root?
func GetRebooter(needsRoot bool) (Rebooter, error) {
	switch WhereAmI() {
	case DockerMode:
		return &Docker{}, nil
	case KubernetesMode:
		return &Kubernetes{}, nil
	case HostMode:
		if needsRoot && !IsSuperUser() {
			return nil, errors.NewActionRequiresSudoError("reboot")
		}
		return &Host{}, nil
	}
	return nil, fmt.Errorf("no executor found for mode %s", WhereAmI())
}

// Deployer is an interface for executors that can deploy Nuvla applications
type Deployer interface {
	Executor
	StartDeployment() error
	StopDeployment() error
	StateDeployment() error
	UpdateDeployment() error
	GetServices() ([]*DeploymentService, error)
}

func GetDeployer(resource *resources.DeploymentResource) (Deployer, error) {
	module := resource.Module
	compatibility := module.Compatibility
	subType := module.SubType

	switch subType {
	case "application":
		switch compatibility {
		case "docker-compose":
			return &Compose{
				deploymentResource: resource,
			}, nil
		case "swarm":
			return &Stack{
				deploymentResource: resource,
			}, nil
		default:
			return nil, errors.NewNotImplementedActionError(compatibility)
		}
	case "application_kubernetes":
		return nil, errors.NewNotImplementedActionError("kubernetes deployment")
	default:
		return nil, errors.NewNotImplementedActionError(subType)
	}
}

// DeploymentService description of a Nuvla deployment service. Should probably become an interface to allow for
// k8s and docker.
type DeploymentService struct {
	Image     string `json:"image,omitempty"`
	Name      string `json:"name,omitempty"`
	ServiceID string `json:"service-id,omitempty"`
	NodeID    string `json:"node-id,omitempty"`
	State     string `json:"state,omitempty"`
	Status    string `json:"status,omitempty"`

	ExternalPorts map[string]int // Only for external ports, protocol: port
}

func NewDeploymentServiceFromContainerSummary(c api.ContainerSummary) *DeploymentService {
	s := &DeploymentService{
		Image:     c.Image,
		Name:      c.Name,
		ServiceID: c.ID,
		NodeID:    c.Name,
		State:     c.State,
		Status:    c.Status,
	}
	if c.Publishers != nil {
		s.ExternalPorts = make(map[string]int)
		for _, p := range c.Publishers {
			s.ExternalPorts[fmt.Sprintf("%s.%d", p.Protocol, p.PublishedPort)] = p.TargetPort
		}
	}

	return s
}

type SSHKeyManager interface {
	InstallSSHKey(sshPub, user string) error
	RevokeSSKKey(sshkey string) error
}
