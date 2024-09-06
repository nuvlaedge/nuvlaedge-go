package deployments

import (
	"context"
	"nuvlaedge-go/orchestrator"
)

type Deployment interface {
	Execute(ctx context.Context) error
}

func DeploymentFactory(action string, orchestrator orchestrator.Orchestrator) Deployment {
	switch action {
	case "start":
		return &DeploymentStart{orchestrator: orchestrator}
	case "stop":
		return &DeploymentStop{orchestrator: orchestrator}
	case "update":
		return &DeploymentUpdate{orchestrator: orchestrator}
	default:
		return nil
	}
}
