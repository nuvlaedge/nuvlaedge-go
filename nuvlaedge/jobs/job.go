package jobs

import (
	nuvla "github.com/nuvla/api-client-go"
	"github.com/nuvla/api-client-go/clients"
	"github.com/nuvla/api-client-go/clients/resources"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/nuvlaedge/jobs/actions"
)

type Job struct {
	JobId  string
	Client *clients.NuvlaJobClient

	JobName     string
	JobType     actions.ActionName
	Action      actions.Action
	JobResource *resources.JobResource
}

func NewJob(jobId string, c *nuvla.NuvlaClient) *Job {
	return &Job{
		JobId:  jobId,
		Client: clients.NewJobClient(jobId, c),
	}
}

func (j *Job) Init() error {
	log.Infof("Initialising jobs %s", j.JobId)
	if err := j.Client.UpdateResource(); err != nil {
		log.Errorf("Error updating jobs resource: %s", err)
		return err
	}

	j.JobResource = j.Client.GetResource()
	j.JobName = j.JobResource.Action

	a, err := actions.GetAction(j.JobName)
	if err != nil {
		log.Errorf("Error getting action: %s", err)
		return err
	}
	j.Action = a
	j.JobType = a.GetActionName()

	return nil
}

func (j *Job) Run() error {
	_ = j.Client.SetProgress(30)

	// Initialise the action
	err := j.Action.Init(
		actions.WithActionName(j.JobName),
		actions.WithJobId(j.JobId),
		actions.WithJobResource(j.JobResource),
		actions.WithClient(j.Client.NuvlaClient))
	if err != nil {
		j.Client.SetState(resources.StateFailed)
		_ = j.Client.SetProgress(100)
		return err
	}

	// Run the action
	if err := j.Action.ExecuteAction(); err != nil {
		j.Client.SetState(resources.StateFailed)
		_ = j.Client.SetProgress(100)
		return err
	}
	j.Client.SetSuccessState()
	return nil
}
