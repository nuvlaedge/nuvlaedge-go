package jobEngine

type DeploymentStartActions struct {
	*ActionBase
}

func (ds *DeploymentStartActions) Execute() error {
	return nil
}

func (ds *DeploymentStartActions) GetActionType() ActionType {
	return DeploymentStartActionType
}

func (ds *DeploymentStartActions) Init(opts *ActionBaseOpts) error {
	return nil
}

func (ds *DeploymentStartActions) GetExecutors() []ExecutorType {
	return nil
}
