package jobEngine

import (
	nuvla "github.com/nuvla/api-client-go"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/nuvlaedge/orchestrator"
)

type ActionType string

const (
	RebootActionType          ActionType = "reboot_nuvlabox"
	DeploymentStopActionType  ActionType = "deployment-stop"
	DeploymentStartActionType ActionType = "deployment-start"
)

type Action interface {
	Execute() error
	GetActionType() ActionType
	Init(*ActionBaseOpts) error
	//GetExecutor() ExecutorType
}

type ActionBase struct {
	nuvlaClient *nuvla.NuvlaClient
	coeClient   orchestrator.Coe
}

func NewActionBase(opts *ActionBaseOpts) *ActionBase {
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
	var a Action

	switch ActionType(actionName) {
	case RebootActionType:
		a = &RebootAction{}

	case DeploymentStopActionType:
		a = &DeploymentStopAction{}

	case DeploymentStartActionType:
		a = &DeploymentStartActions{}

	default:
		return nil
	}

	log.Infof("Initialising action: %s...", actionName)
	if err := a.Init(defaultOpts); err != nil {
		log.Errorf("Error creating the new action")
		return nil
		// TODO: Maybe here handle an error or customise errors...
	}
	log.Infof("Initialising action: %s... Success", actionName)
	return a
}
