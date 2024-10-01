package job_processor

import (
	"context"
	"errors"
	"fmt"
	nuvla "github.com/nuvla/api-client-go"
	"github.com/nuvla/api-client-go/clients"
	"github.com/nuvla/api-client-go/clients/resources"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/engine"
	errors2 "nuvlaedge-go/types/errors"
	"nuvlaedge-go/types/jobs"
	"nuvlaedge-go/workers/job_processor/actions"
	"time"
)

const (
	JobEngineContainerImage = "sixsq/nuvlaedge:latest"
)

type Job interface {
	RunJob(ctx context.Context) error
	Init(ctx context.Context, coe engine.Coe, enableLegacy bool, legacyImage string) (Job, error)
	GetId() string
	GetJobType() string
}

func NewJob(ctx context.Context, jobId string, c *nuvla.NuvlaClient, coe engine.Coe, enableLegacy bool, legacyImage string) (Job, error) {
	job := JobBase{
		JobId:  jobId,
		Client: clients.NewJobClient(jobId, c),
	}
	j, err := job.Init(ctx, coe, enableLegacy, legacyImage)
	if err != nil {
		return nil, err
	}
	return j, nil
}

type JobBase struct {
	JobId       string
	JobType     string
	Client      *clients.NuvlaJobClient
	JobResource *resources.JobResource
}

func (j *JobBase) GetJobType() string {
	return j.JobType
}

func isNotSupportedActionError(err error) bool {
	var notImplementedActionError errors2.NotImplementedActionError
	return errors.As(err, &notImplementedActionError)
}

func (j *JobBase) Init(ctx context.Context, coe engine.Coe, enableLegacy bool, legacyImage string) (Job, error) {
	log.Infof("Initialising job %s", j.JobId)
	if err := j.Client.UpdateResource(ctx); err != nil {
		log.Errorf("Error updating job resource: %s", err)
		return nil, err
	}
	j.JobResource = j.Client.GetResource()
	j.JobType = j.JobResource.Action
	// Looks for the action in the implemented interface Action in the actions package
	a, err := actions.GetAction(j.JobResource.Action)

	if err == nil {
		return NewNativeJobFromBase(j, a, j.JobResource.Action), nil
	}

	// If the action is not supported here, try to run it in the container
	if isNotSupportedActionError(err) {
		if !enableLegacy {
			log.Infof("Legacy actions are disabled, cannot run unsupported job %s", j.JobId)
			j.Client.SetFailedState(
				ctx,
				fmt.Sprintf("NuvlaEdge-Go doesn't support action %s. "+
					"Set env JOB_LEGACY_ENABLE=true to run unsupported actions in a separate container", j.JobResource.Action))
			return nil, err
		}
		return NewContainerEngineJobFromBase(j, coe, legacyImage), nil
	} else {
		log.Errorf("Unexpected error creating new Job: %s", err)
		return nil, err
	}
}

func (j *JobBase) GetId() string {
	return j.JobId
}

type NativeJob struct {
	*JobBase
	JobName string
	JobType actions.ActionName
	Action  actions.Action
}

func NewNativeJobFromBase(jb *JobBase, action actions.Action, actionName string) *NativeJob {
	return &NativeJob{
		JobBase: jb,
		JobType: action.GetActionName(),
		Action:  action,
		JobName: actionName,
	}
}

func (j *NativeJob) RunJob(ctx context.Context) error {
	_ = j.Client.SetProgress(ctx, 30)

	// Initialise the action
	err := j.Action.Init(
		ctx,
		actions.WithActionName(j.JobName),
		actions.WithJobId(j.JobId),
		actions.WithJobResource(j.JobResource),
		actions.WithClient(j.Client.NuvlaClient))
	if err != nil {
		j.Client.SetFailedState(ctx, err.Error())
		return err
	}

	// Run the action
	if err = j.Action.ExecuteAction(ctx); err != nil {
		errMsg := j.Action.GetOutput() + "\n" + err.Error()
		j.Client.SetFailedState(ctx, errMsg)
		return err
	}

	okMsg := j.Action.GetOutput() + "\n" + "Success running job"
	j.Client.SetStatusMessage(ctx, okMsg)
	j.Client.SetSuccessState(ctx)
	return nil
}

type ContainerEngineJob struct {
	*JobBase
	coe            engine.Coe
	ContainerImage string
}

func NewContainerEngineJobFromBase(jb *JobBase, coe engine.Coe, legacyImage string) *ContainerEngineJob {
	return &ContainerEngineJob{
		JobBase:        jb,
		coe:            coe,
		ContainerImage: legacyImage,
	}
}

func (cj *ContainerEngineJob) RunJob(ctx context.Context) error {
	k, s, err := cj.Client.GetCredentials()
	if err != nil {
		log.Errorf("Error getting credentials: %s", err)
		return err
	}
	conf := &jobs.LegacyJobConf{
		JobId:            cj.JobId,
		Image:            cj.ContainerImage,
		ApiKey:           k,
		ApiSecret:        s,
		Endpoint:         cj.Client.SessionOpts.Endpoint,
		EndpointInsecure: cj.Client.SessionOpts.Insecure,
	}
	containerId, err := cj.coe.RunJobEngineContainer(ctx, conf)
	if err != nil {
		log.Errorf("Error running container: %s", err)
		return err
	}

	// Wait container to finish
	log.Infof("Waiting job to finish...")
	finishStatus, err := cj.coe.WaitContainerFinish(ctx, containerId, 60*time.Second, true)
	log.Infof("Container Job finished with status: %d", finishStatus)
	if err != nil {
		log.Errorf("Error waiting container to finish: %s", err)
		return err
	}

	if finishStatus != 0 {
		return errors.New("container job finished with error")
	}
	log.Infof("Success running container job")
	return nil
}
