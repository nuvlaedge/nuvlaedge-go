package deployments

import (
	"nuvlaedge-go/job_engine/actions"
)

type DeploymentBase struct {
}

func NewDeploymentBase(opts *actions.ActionOpts) *DeploymentBase {
	return &DeploymentBase{}
}
