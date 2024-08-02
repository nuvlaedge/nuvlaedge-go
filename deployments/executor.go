package deployments

import (
	"context"
	"nuvlaedge-go/orchestrator"
)

type DeploymentExecutor interface {
	StartDeployment(ctx context.Context) error
	StopDeployment(ctx context.Context) error
	StateDeployment(ctx context.Context) (string, error)
	UpdateDeployment(ctx context.Context) error
	LogsDeployment(ctx context.Context) (string, error)
}

type ExecutorBase struct {
	coe orchestrator.Orchestrator
}

type ComposeExecutor struct {
	*ExecutorBase
}

func (ce *ComposeExecutor) StartDeployment(ctx context.Context) error {

	return nil
}
