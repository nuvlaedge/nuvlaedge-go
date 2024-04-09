package jobProcessor

import (
	"fmt"
	nuvla "github.com/nuvla/api-client-go"
	"github.com/nuvla/api-client-go/clients"
	"github.com/nuvla/api-client-go/types"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/nuvlaedge/orchestrator"
)

type Job struct {
	ID     *types.NuvlaID
	client *clients.NuvlaJobClient
	coe    orchestrator.Coe

	action     Action
	actionType ActionType
}

func NewJob(jobId *types.NuvlaID, client *nuvla.NuvlaClient, coe orchestrator.Coe) *Job {
	return &Job{
		ID:     jobId,
		client: clients.NewJobClient(jobId.Id, client),
		coe:    coe,
	}
}

func (j *Job) Start() error {
	log.Infof("Starting job %s", j.ID.Id)
	err := j.client.UpdateResource()

	j.client.SetInitialState()
	if err != nil {
		log.Errorf("Error updating job resource: %s, cannot continue.", err)
		return err
	}

	j.action = GetAction(j.client.GetActionName())
	if j.action == nil {
		log.Errorf("Error getting action %s for job %s", j.client.GetActionName(), j.ID.Id)
		return fmt.Errorf("error getting action for job %s", j.ID.Id)
	}

	j.actionType = j.action.GetActionType()
	log.Infof("Initialising action: %s...", j.actionType)

	err = j.action.Init(WithCoeClient(j.coe), WithNuvlaClient(j.client.NuvlaClient), WithJobResource(j.client.GetResource()))
	if err != nil {
		log.Errorf("Error initialising action %s for job %s", err, j.ID.Id)
	}
	log.Infof("Initialising action: %s... Success", j.actionType)

	return nil
}

func (j *Job) Run() error {
	_ = j.client.SetProgress(30)

	if err := j.action.ExecuteAction(); err != nil {
		log.Errorf("Error executing action %s for job %s", err, j.ID.Id)
		// TODO: If execute action returns an error we need to push it to the job
		j.client.SetState(clients.StateFailed)
		_ = j.client.SetProgress(100)
		return err
	}

	j.client.SetSuccessState()

	return nil
}

func (j *Job) Stop() error {
	return nil
}

func (j *Job) GetState() string {
	return ""
}

func (j *Job) String() string {
	return fmt.Sprintf("Job with ID %s running on STATE %s", j.ID, j.GetState())
}
