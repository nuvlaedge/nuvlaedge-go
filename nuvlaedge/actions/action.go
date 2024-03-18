package actions

import (
	nuvla "github.com/nuvla/api-client-go/clients"
	"nuvlaedge-go/nuvlaedge/orchestrator"
)

type ActionType string

const (
	RebootActionType          ActionType = "reboot"
	DeploymentStopActionType  ActionType = "deployment-stop"
	DeploymentStartActionType ActionType = "deployment-start"
)

type Action interface {
	Execute() error
	GetActionType() ActionType
	Init(ActionBaseOpts) error
}

type ActionBase struct {
	nuvlaClient *nuvla.NuvlaEdgeClient
	coeClient   orchestrator.Coe
}

func NewActionBase(opts ActionBaseOpts) *ActionBase {
	return &ActionBase{
		nuvlaClient: opts.NuvlaClient,
		coeClient:   opts.CoeClient,
	}
}

func NewAction(actionName string, opts ...ActionBaseOptsFunc) Action {
	defaultOpts := DefaultActionBaseOpts()
	for _, fn := range opts {
		fn(defaultOpts)
	}

	switch ActionType(actionName) {
	case RebootActionType:
		return &RebootAction{}

	case DeploymentStopActionType:
		return &DeploymentStopAction{}

	case DeploymentStartActionType:
		return &DeploymentStartActions{}

	default:
		return nil
	}
}
