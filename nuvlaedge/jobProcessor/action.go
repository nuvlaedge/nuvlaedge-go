package jobProcessor

import (
	nuvla "github.com/nuvla/api-client-go"
	"github.com/nuvla/api-client-go/clients"
	"nuvlaedge-go/nuvlaedge/orchestrator"
)

type ActionType string

const (
	RebootActionType          ActionType = "reboot_nuvlabox"
	StopDeploymentActionType  ActionType = "stop_deployment"
	StartDeploymentActionType ActionType = "start_deployment"
	StateDeploymentActionType ActionType = "deployment_state"
)

func GetActionTypeFromString(action string) ActionType {
	switch action {
	case "reboot_nuvlabox":
		return RebootActionType
	case "stop_deployment":
		return StopDeploymentActionType
	case "start_deployment":
		return StartDeploymentActionType
	case "deployment_state_10":
		return StateDeploymentActionType
	case "deployment_state_60":
		return StateDeploymentActionType
	default:
		return ""
	}
}

type Action interface {
	Init(opts ...ActionOptsFunc) error
	ExecuteAction() error
	GetActionType() ActionType
}

func GetAction(actionName string) Action {
	switch GetActionTypeFromString(actionName) {
	case RebootActionType:
		return &RebootAction{}
	case StopDeploymentActionType:
		return &StopDeployment{}
	case StartDeploymentActionType:
		return &StartDeployment{}
	case StateDeploymentActionType:
		return &DeploymentState{}
	default:
		return nil
	}
}

type ActionOpts struct {
	NuvlaClient *nuvla.NuvlaClient
	CoeClient   orchestrator.Coe
	JobResource *clients.JobResource
}

type ActionOptsFunc func(*ActionOpts)

func DefaultActionBaseOpts() *ActionOpts {
	return &ActionOpts{}
}

func WithNuvlaClient(nuvlaClient *nuvla.NuvlaClient) ActionOptsFunc {
	return func(opts *ActionOpts) {
		opts.NuvlaClient = nuvlaClient
	}
}

func WithCoeClient(coeClient orchestrator.Coe) ActionOptsFunc {
	return func(opts *ActionOpts) {
		opts.CoeClient = coeClient
	}
}

func WithJobResource(jobResource *clients.JobResource) ActionOptsFunc {
	return func(opts *ActionOpts) {
		opts.JobResource = jobResource
	}
}
