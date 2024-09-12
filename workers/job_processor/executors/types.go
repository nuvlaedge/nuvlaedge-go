package executors

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/docker/api/types/swarm"
	"github.com/nuvla/api-client-go/clients/resources"
	"nuvlaedge-go/types/errors"
	"strconv"
	"strings"
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
	GetServices() ([]DeploymentService, error)
	// Close TODO: For the moment, we only need to close dockerCLI
	Close() error
}

func GetDeployer(resource *resources.DeploymentResource) (Deployer, error) {
	module := resource.Module
	compatibility := module.Compatibility
	subType := module.SubType

	switch subType {
	case "application":
		switch compatibility {
		case "docker-compose":
			return &ComposeExecutor{
				ctx:                context.Background(),
				ExecutorBase:       ExecutorBase{Name: ComposeExecutorName},
				deploymentResource: resource,
			}, nil
		case "swarm":
			return &Stack{
				ExecutorBase:       ExecutorBase{Name: StackExecutorName},
				deploymentResource: resource,
				context:            context.Background(),
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

type DeploymentService interface {
	GetServiceMap() map[string]string
	GetPorts() map[string]int
}

// DeploymentComposeService description of a Nuvla deployment service. Should probably become an interface to allow for
// k8s and docker.
type DeploymentComposeService struct {
	Image     string `json:"image,omitempty"`
	Name      string `json:"name,omitempty"`
	ServiceID string `json:"service-id,omitempty"`
	NodeID    string `json:"node-id,omitempty"`
	State     string `json:"state,omitempty"`
	Status    string `json:"status,omitempty"`

	ExternalPorts map[string]int `json:"-"` // Only for external ports, protocol: port
}

func (s *DeploymentComposeService) GetServiceMap() map[string]string {
	// Convert struct to Map
	m := make(map[string]string)
	b, _ := json.Marshal(s)
	_ = json.Unmarshal(b, &m)
	return m
}

func (s *DeploymentComposeService) GetPorts() map[string]int {
	return s.ExternalPorts
}

func NewDeploymentServiceFromContainerSummary(c api.ContainerSummary) *DeploymentComposeService {
	s := &DeploymentComposeService{
		Image:     c.Image,
		Name:      c.Name,
		ServiceID: c.ID,
		NodeID:    c.Service,
		State:     c.State,
		Status:    c.Status,
	}
	if c.Publishers != nil {
		s.ExternalPorts = make(map[string]int)
		for _, p := range c.Publishers {
			s.ExternalPorts[fmt.Sprintf("%s.%d", p.Protocol, p.TargetPort)] = p.PublishedPort
		}
	}

	return s
}

type DeploymentStackService struct {
	ServiceID string `json:"service-id,omitempty"`
	Mode      string `json:"mode,omitempty"`
	Image     string `json:"image,omitempty"`
	NodeID    string `json:"node-id,omitempty"`
	Desired   string `json:"replicas.desired,omitempty"`
	Running   string `json:"replicas.running,omitempty"`

	Ports map[string]int `json:"-"`
}

func (s *DeploymentStackService) GetServiceMap() map[string]string {
	// Convert struct to Map
	m := make(map[string]string)
	b, _ := json.Marshal(s)
	_ = json.Unmarshal(b, &m)
	return m
}

func (s *DeploymentStackService) GetPorts() map[string]int {
	return s.Ports
}

func NewDeploymentStackServiceFromServiceSummary(s swarm.Service) *DeploymentStackService {
	dService := &DeploymentStackService{
		ServiceID: s.ID,
	}

	// Mode Extraction
	if s.Spec.Mode.Replicated != nil {
		dService.Mode = "replicated"
	}

	// Image extraction
	if image, ok := s.Spec.Labels["com.docker.stack.image"]; ok {
		dService.Image = image
	} else {
		i := s.Spec.TaskTemplate.ContainerSpec.Image
		if strings.Contains(i, "@") {
			dService.Image = strings.Split(i, "@")[0]
		} else {
			dService.Image = i
		}
	}

	// Extract Node ID from Service Name
	if strings.Contains(s.Spec.Name, "_") {
		dService.NodeID = strings.Split(s.Spec.Name, "_")[1]
	} else {
		dService.NodeID = s.Spec.Name
	}

	// Replicas
	dService.Desired = strconv.FormatUint(s.ServiceStatus.DesiredTasks, 10)
	dService.Running = strconv.FormatUint(s.ServiceStatus.RunningTasks, 10)

	// Ports
	dService.Ports = make(map[string]int)
	for _, p := range s.Endpoint.Ports {
		dService.Ports[fmt.Sprintf("%s.%d", p.Protocol, p.TargetPort)] = int(p.PublishedPort)
	}
	return dService
}

type SSHKeyManager interface {
	InstallSSHKey(sshPub, user string) error
	RevokeSSKKey(sshkey string) error
}
