package jobs

import (
	"context"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/types"
)

type JobProcessor struct {
	types.WorkerBase
	jobChan           chan string
	deploymentJobChan chan string

	runningJobs  types.JobRegistry
	enableLegacy bool
}

func NewJobProcessor(jobCh chan string, deploymentJobCh chan string, period int, enableLegacy bool) *JobProcessor {
	return &JobProcessor{
		WorkerBase:        types.NewWorkerBase(period, types.JobProcessor),
		jobChan:           jobCh,
		deploymentJobChan: deploymentJobCh,
		runningJobs:       types.NewRunningJobs(),
		enableLegacy:      enableLegacy,
	}
}

func (p *JobProcessor) Run(ctx context.Context) error {
	p.Status = types.RUNNING
	defer p.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("Context closed, stopping job processor")
			return ctx.Err()
		case job := <-p.jobChan:
			log.Info("Received job: ", job)
			go p.processJob(ctx, job)
		}
	}
}

func (p *JobProcessor) processJob(ctx context.Context, job string) {
	if p.runningJobs.Exists(job) {
		log.Warn("Job already running: ", job)
		return
	}
	log.Info("Processing job: ", job)

}
