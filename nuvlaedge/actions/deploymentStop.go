package actions

type DeploymentStopAction struct {
}

func (ds *DeploymentStopAction) Execute() error {
	return nil
}

func (ds *DeploymentStopAction) GetActionType() ActionType {
	return DeploymentStopActionType
}

func (ds *DeploymentStopAction) Init(opts ActionBaseOpts) error {
	return nil
}
