package deployments

import (
	"context"
	"nuvlaedge-go/job_engine/actions"
)

type DeploymentState struct {
	*DeploymentBase
}

func (a *DeploymentState) Init(opts *actions.ActionOpts) error {
	a.DeploymentBase = NewDeploymentBase(opts)
	return nil
}

func (a *DeploymentState) Execute(context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (a *DeploymentState) State() error {
	return nil
}

func (a *DeploymentState) stateCompose() error {
	return nil
}

func (a *DeploymentState) stateKubernetes() error {
	return nil
}

func (a *DeploymentState) stateSwarm() error {
	return nil
}

func (a *DeploymentState) stateHelm() error {
	return nil
}
