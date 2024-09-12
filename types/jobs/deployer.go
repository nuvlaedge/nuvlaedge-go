package jobs

import (
	"context"
	"github.com/nuvla/api-client-go/clients/resources"
)

type Logs map[string]string

type Deployer interface {
	StartDeployment(ctx context.Context, resource *resources.DeploymentResource) error
	StopDeployment(ctx context.Context, deploymentId string) error
	UpdateDeployment(ctx context.Context, resource *resources.DeploymentResource) error
	GetDeploymentLogs(ctx context.Context, deploymentId string) (*Logs, error)
	GetDeploymentState(ctx context.Context, deploymentId string) error
}
