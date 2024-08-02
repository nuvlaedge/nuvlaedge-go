package jobs

import (
	nuvlaApi "github.com/nuvla/api-client-go"
	"github.com/nuvla/api-client-go/clients"
	"nuvlaedge-go/jobs/actions"
	"nuvlaedge-go/orchestrator"
)

type LegacyJob struct {
	actions.JobBase
}

func NewLegacyJob(jobId string, nuvla *nuvlaApi.NuvlaClient) *LegacyJob {
	return &LegacyJob{
		JobBase: actions.NewJobBase(jobId, "", clients.NewJobClient(jobId, nuvla))}
}

func (j *LegacyJob) RunJob(coe orchestrator.Orchestrator, engine interface{}) error {
	return nil
}

func (j *LegacyJob) Init() {

}
