package actions

import (
	"context"
	"fmt"
	"github.com/nuvla/api-client-go/clients"
	"github.com/nuvla/api-client-go/clients/resources"
	"nuvlaedge-go/types"
	"slices"
)

type Job interface {
	RunJob(ctx context.Context) error
	Init(opts types.JobOpts) error
	GetId() string
	GetJobType() types.JobType
}

func NewJob(jobType types.JobType) (Job, error) {
	if slices.Contains(types.SupportedJobTypes, jobType) {
		return nil, nil
	} else {
		return nil, fmt.Errorf("job type not supported") // nolint: goerr113 // This is a placeholder for the actual error message
	}
}

type JobBase struct {
	JobId    string
	JobType  types.JobType
	Client   *clients.NuvlaJobClient
	Resource *resources.JobResource
}

func NewJobBase(jobId string, jobType types.JobType, client *clients.NuvlaJobClient) JobBase {
	return JobBase{
		JobId:    jobId,
		JobType:  jobType,
		Client:   client,
		Resource: &resources.JobResource{},
	}
}

func (j *JobBase) GetJobType() types.JobType {
	return j.JobType
}

func (j *JobBase) GetId() string {
	return j.JobId
}
