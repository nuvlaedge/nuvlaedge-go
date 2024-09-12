package job_engine

import (
	"context"
	"github.com/nuvla/api-client-go/clients"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/job_engine/actions"
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
}

func (j *JobProcessor) Init(opts *worker.WorkerOpts, conf *worker.WorkerConfig) error {
	// Init
	j.WorkerBase = worker.NewWorkerBase(worker.JobProcessor)
	j.jobChan = opts.JobCh
	j.deploymentJobChan = opts.DeploymentCh
	j.runningJobs = opts.Jobs

	// Clients setup
	j.client = opts.NuvlaClient

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
			go j.processJob(ctx, job)

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

func (j *JobProcessor) processJob(ctx context.Context, jobId string) {
	if j.runningJobs.Exists(jobId) {
		log.Infof("Job %s is already running", jobId)
		return
	}
	defer j.runningJobs.Remove(jobId)

	job, err := jobs.NewJobBase(jobId, j.client.NuvlaClient)
	if err != nil {
		log.Errorf("Error creating jobId %s: %s", jobId, err)
		return
	}

	j.runningJobs.Add(&jobs.RunningJob{
		JobId:   jobId,
		JobType: job.JobType,
	})

	if !jobs.IsSupportedJob(job.JobType) && j.enableLegacy {
		// Run legacy container...
		job.JobType = "legacy_job"
	}

	ctxTime, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	err = actions.RunJob(ctxTime, job, j.legacyImage)
	if err != nil {
		log.Errorf("Error running job %s: %s", jobId, err)
		return
	}

}
