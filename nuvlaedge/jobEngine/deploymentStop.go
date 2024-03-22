package jobEngine

type DeploymentStopAction struct {
	ActionBase

	executor  Executor
	executors []ExecutorType
}

func NewDeploymentStopAction(opts *ActionBaseOpts) *DeploymentStopAction {
	return &DeploymentStopAction{
		ActionBase: *NewActionBase(opts),
		executors:  []ExecutorType{DockerComposeExecutorType, DockerSwarmExecutorType, DockerServiceExecutorType, DockerContainerExecutorType, KubernetesExecutorType},
	}
}

func (ds *DeploymentStopAction) Execute() error {
	return nil
}

func (ds *DeploymentStopAction) GetActionType() ActionType {
	return DeploymentStopActionType
}

func (ds *DeploymentStopAction) Init(opts *ActionBaseOpts) error {
	return nil
}

func (ds *DeploymentStopAction) GetExecutors() []ExecutorType {
	return ds.executors
}
