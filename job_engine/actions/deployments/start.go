package deployments

import (
	"context"
	"nuvlaedge-go/job_engine/actions"
	"nuvlaedge-go/job_engine/connector"
)

type DeploymentStart struct {
	*DeploymentBase
}

func (a *DeploymentStart) Init(opts *actions.ActionOpts) error {
	a.DeploymentBase = NewDeploymentBase(opts)
	return nil
}

func (a *DeploymentStart) Execute(context.Context) error {
	//TODO implement me
	return a.Start()
}

func (a *DeploymentStart) Start() error {

	return nil
}

func (a *DeploymentStart) startCompose() error {
	_, err := connector.NewComposeConnector(nil)
	if err != nil {
		return err
	}

	return nil
}

func (a *DeploymentStart) startKubernetes() error {
	return nil
}

func (a *DeploymentStart) startSwarm() error {
	return nil
}

func (a *DeploymentStart) startHelm() error {
	return nil
}
