package deployments

import (
	"context"
	"nuvlaedge-go/job_engine/actions"
)

type DeploymentUpdate struct {
	*DeploymentBase
}

func (a *DeploymentUpdate) Init(opts *actions.ActionOpts) error {
	a.DeploymentBase = NewDeploymentBase(opts)
	return nil
}

func (a *DeploymentUpdate) Execute(context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (a *DeploymentUpdate) Update() error {
	return nil
}

func (a *DeploymentUpdate) updateCompose() error {
	return nil
}

func (a *DeploymentUpdate) updateKubernetes() error {
	return nil
}

func (a *DeploymentUpdate) updateSwarm() error {
	return nil
}

func (a *DeploymentUpdate) updateHelm() error {
	return nil
}
