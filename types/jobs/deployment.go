package jobs

import (
	"context"
	"github.com/nuvla/api-client-go/clients"
)

type DeploymentOpts struct {
	Context          context.Context
	Job              *JobBase
	DeploymentClient *clients.NuvlaDeploymentClient
	Orchestrator     Deployer
}
