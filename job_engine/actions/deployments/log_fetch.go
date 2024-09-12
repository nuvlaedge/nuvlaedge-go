package deployments

import (
	"context"
	"nuvlaedge-go/job_engine/actions"
)

type DeploymentLogFetch struct {
	*DeploymentBase
}

func (a *DeploymentLogFetch) Init(opts *actions.ActionOpts) error {
	a.DeploymentBase = NewDeploymentBase(opts)

	return nil
}

func (a *DeploymentLogFetch) Execute(context.Context) error {
	//TODO implement me
	panic("implement me")
}
