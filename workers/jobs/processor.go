package jobs

import (
	"context"
	"github.com/nuvla/api-client-go/clients"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/common/constants"
	"nuvlaedge-go/types/jobs"
	"nuvlaedge-go/types/worker"
	"time"
)

type JobProcessor struct {
	worker.WorkerBase
	jobChan           chan string
	deploymentJobChan chan jobs.Job

	client *clients.NuvlaEdgeClient

	runningJobs  *jobs.JobRegistry
	enableLegacy bool
	legacyImage  string

	containerExecutor jobs.JobExecutor
	hostExecutor      jobs.JobExecutor
}

func (j *JobProcessor) Init(opts *worker.WorkerOpts, conf *worker.WorkerConfig) error {
	// Init
	j.WorkerBase = worker.NewWorkerBase(worker.JobProcessor)
	j.jobChan = opts.JobCh
	j.deploymentJobChan = opts.DeploymentCh
	j.runningJobs = opts.Jobs

	// Clients setup
	j.client = opts.NuvlaClient
	j.containerExecutor = NewDockerExecutor(opts.DockerClient)

	// Config
	j.enableLegacy = conf.EnableJobLegacy
	j.legacyImage = conf.LegacyJobImage
	return nil
}

func (j *JobProcessor) Start(ctx context.Context) error {
	go j.Run(ctx)
	return nil
}

func (j *JobProcessor) Reconfigure(conf *worker.WorkerConfig) error {
	j.enableLegacy = conf.EnableJobLegacy
	j.legacyImage = conf.LegacyJobImage

	return nil
}

func (j *JobProcessor) Run(ctx context.Context) error {

	for {
		select {
		case <-ctx.Done():
			if err := j.Stop(ctx); err != nil {
				return err
			}
			return ctx.Err()

		case job := <-j.jobChan:
			go j.processJob(job)

		case conf := <-j.ConfChan:
			log.Info("Received configuration in Job Processor: ", conf)
			if err := j.Reconfigure(conf); err != nil {
				log.Error("Failed to reconfigure job processor: ", err)
			}
		}
	}
}

func (j *JobProcessor) Stop(ctx context.Context) error {
	//TODO implement me
	return nil
}

func (j *JobProcessor) processJob(jobId string) {
	if j.runningJobs.Exists(jobId) {
		log.Infof("Job %s is already running", jobId)
		return
	}

	job, err := jobs.NewJobBase(jobId, j.client.NuvlaClient)
	if err != nil {
		log.Errorf("Error creating jobId %s: %s", jobId, err)
		j.runningJobs.Remove(jobId)
		return
	}

	log.Infof("Job Processor starting new jobId with id %s", jobId)
	if jobs.IsDeployment(job.JobType) {
		log.Infof("Job %s is a deployment action, sending to Deployment Handler", jobId)
		j.deploymentJobChan <- job
		return
	}

	if !jobs.IsSupportedJob(job.JobType) {
		// Run legacy container...
		log.Infof("Job %s is not supported natively, running legacy container", jobId)
		j.runningJobs.Remove(jobId)
		return
	}

	jobOpts := &jobs.JobOpts{
		Job:         job,
		ContainerEx: j.containerExecutor,
		HostEx:      j.hostExecutor,
	}

	ctx, cancel := context.WithTimeout(context.Background(), constants.DefaultJobTimeout*time.Second)
	defer cancel()
	if err := RunJob(ctx, jobOpts); err != nil {
		log.Errorf("Error running job %s: %s", jobId, err)
	}
	j.runningJobs.Remove(jobId)
}
