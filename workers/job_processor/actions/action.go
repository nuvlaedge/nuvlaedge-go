package actions

import (
	"context"
	nuvla "github.com/nuvla/api-client-go"
	"github.com/nuvla/api-client-go/clients/resources"
	"nuvlaedge-go/types/errors"
	"nuvlaedge-go/workers/job_processor/executors"
)

type Action interface {
	// Init allows implementations of the interface to each create their own client from the generic NuvlaClient
	// and retrieve any field required from the jobs resource.
	Init(ctx context.Context, opts ...ActionOptsFn) error
	// GetExecutorName returns the name of the executor that will be used to execute the action. This will be
	// initialised after asserting the executor. It should return an empty sting if the executor is not set.
	GetExecutorName() executors.ExecutorName // TODO: Should be change to ExecutorName type when it is defined
	// assertExecutor creates the specific required executor for the given actions and sets it in the executor field.
	// Each action will have their onw executor extension interface.
	assertExecutor() error
	// ExecuteAction will execute the action using the executor.
	ExecuteAction(ctx context.Context) error
	// GetActionName returns the name of the action being implemented
	GetActionName() ActionName

	// GetOutput Action stdout and stderr
	GetOutput() string
}

type ActionBase struct {
	actionName ActionName
}

func (a *ActionBase) GetActionName() ActionName {
	return a.actionName
}

type ActionName string

const (
	RebootActionName           ActionName = "reboot_nuvlabox"
	StopDeploymentActionName   ActionName = "stop_deployment"
	StartDeploymentActionName  ActionName = "start_deployment"
	StateDeploymentActionName  ActionName = "deployment_state"
	UpdateDeploymentActionName ActionName = "update_deployment"
	UpdateNuvlaEdge            ActionName = "nuvlabox_update"
	CoeResourceActions         ActionName = "coe_resource_actions"
	UnknownActionName          ActionName = "unknown"
)

var ActionNameMap = map[ActionName]ActionName{
	"reboot_nuvlabox":     RebootActionName,
	"stop_deployment":     StopDeploymentActionName,
	"start_deployment":    StartDeploymentActionName,
	"deployment_state_10": StateDeploymentActionName,
	"deployment_state_60": StateDeploymentActionName,
	"update_deployment":   UpdateDeploymentActionName,
	"nuvlabox_update":     UpdateNuvlaEdge,
}

func getActionNameFromString(action string) ActionName {
	a, ok := ActionNameMap[ActionName(action)]
	if !ok {
		return UnknownActionName
	}
	return a
}

func GetAction(actionName string) (Action, error) {
	switch getActionNameFromString(actionName) {
	case RebootActionName:
		return &RebootAction{}, nil
	case StopDeploymentActionName:
		return &DeploymentStop{}, nil
	case StartDeploymentActionName:
		return &DeploymentStart{}, nil
	case StateDeploymentActionName:
		return &DeploymentState{}, nil
	case UpdateDeploymentActionName:
		return &DeploymentUpdate{}, nil
	case CoeResourceActions:
		return &COEResourceActions{}, nil
	//case UpdateNuvlaEdge:
	//	return &Update{}, nil
	default:
		return nil, errors.NewNotImplementedActionError(actionName)
	}
}

type ActionOpts struct {
	ActionName  string                 `json:"action-name,omitempty"`
	JobId       string                 `json:"jobs-id,omitempty"`
	JobResource *resources.JobResource `json:"jobs-resource,omitempty"`
	Client      *nuvla.NuvlaClient     `json:"client,omitempty"`
	IPs         []string               `json:"ips,omitempty"`
}

func NewDefaultActionOpts() *ActionOpts {
	return &ActionOpts{
		ActionName:  "",
		JobId:       "",
		IPs:         nil,
		JobResource: nil,
		Client:      nil,
	}
}

type ActionOptsFn func(*ActionOpts)

func WithActionName(name string) ActionOptsFn {
	return func(opts *ActionOpts) {
		opts.ActionName = name
	}
}

func WithJobId(id string) ActionOptsFn {
	return func(opts *ActionOpts) {
		opts.JobId = id
	}
}

func WithJobResource(resource *resources.JobResource) ActionOptsFn {
	return func(opts *ActionOpts) {
		opts.JobResource = resource
	}
}

func WithClient(client *nuvla.NuvlaClient) ActionOptsFn {
	return func(opts *ActionOpts) {
		opts.Client = client
	}
}

func GetActionOpts(optsFn ...ActionOptsFn) *ActionOpts {
	opts := NewDefaultActionOpts()
	for _, fn := range optsFn {
		fn(opts)
	}
	return opts
}
