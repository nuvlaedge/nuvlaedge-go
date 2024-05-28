package actions

import "nuvlaedge-go/nuvlaedge/jobs/executors"

type DeploymentUpdate struct {
	DeploymentBase
}

func (d *DeploymentUpdate) assertExecutor() error {
	return nil
}

func (d *DeploymentUpdate) ExecuteAction() error {
	return nil
}

func (d *DeploymentUpdate) GetExecutorName() executors.ExecutorName {
	return d.executor.GetName()
}
