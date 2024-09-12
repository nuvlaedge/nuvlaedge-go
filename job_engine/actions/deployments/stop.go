package deployments

import (
	"context"
	"nuvlaedge-go/job_engine/actions"
)

type DeploymentStop struct {
	*DeploymentBase
}

func (a *DeploymentStop) Init(opts *actions.ActionOpts) error {
	a.DeploymentBase = NewDeploymentBase(opts)
	return nil
}

func (a *DeploymentStop) Execute(context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (a *DeploymentStop) Stop() error {
	return nil
}

func (a *DeploymentStop) stopCompose() error {
	return nil
}

func (a *DeploymentStop) stopKubernetes() error {
	return nil
}

func (a *DeploymentStop) stopSwarm() error {
	return nil
}

func (a *DeploymentStop) stopHelm() error {
	return nil
}
