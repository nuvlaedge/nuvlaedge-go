package deployments

import (
	"nuvlaedge-go/orchestrator"
	"nuvlaedge-go/types/jobs"
)

type DeploymentStart struct {
	*jobs.JobBase

	orchestrator orchestrator.Orchestrator
}
