package nuvlaedge

import (
	nuvla "github.com/nuvla/api-client-go"
	"github.com/nuvla/api-client-go/clients"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/nuvlaedge/jobEngine"
	"nuvlaedge-go/nuvlaedge/orchestrator"
	"sync"
)

type Job struct {
	ID          string
	containerID string
}

type JobProcessor struct {
	runningJobs sync.Map
	jobChan     chan string        // Job channel. Receives job IDs from the agent
	exitChan    chan bool          // Exit channel. Receives exit signal from the agent
	client      *nuvla.NuvlaClient // Nuvla session required in the jobs and deployment clients
	coe         orchestrator.Coe   // COE client required in the jobs and deployment clients
}

func NewJobProcessor(jobChan chan string, client *nuvla.NuvlaClient, coe orchestrator.Coe) *JobProcessor {
	return &JobProcessor{
		jobChan: jobChan,
		client:  client,
		coe:     coe,
	}
}

func (p *JobProcessor) Start() error {
	log.Infof("Nothing to start in the job processor, passing...")
	return nil
}

func (p *JobProcessor) Stop() error {
	return nil
}

func (p *JobProcessor) Run() error {
	log.Info("Running Job Engine")
	go func() {
		for {
			select {
			case job := <-p.jobChan:
				go p.processJob(job)
			case <-p.exitChan:
				log.Warn("Job Processor received exit signal")
				return
			}
		}
	}()
	return nil
}

func (p *JobProcessor) processJob(j string) {
	if _, ok := p.runningJobs.Load(j); ok {
		log.Infof("Job %s is already running", j)
		return
	}

	log.Infof("Job Processor starting new job with id %s", j)

	// 1. Create JobClient struct
	jobClient := clients.NewJobClient(j, p.client)
	log.Warnf("JobClient: %v", jobClient)
	err := jobClient.UpdateResource()
	if err != nil {
		log.Errorf("Error retrieving Job %s: %s", j, err)
		log.Errorf("Job %s will not be processed", j)
		return
	}

	jobClient.PrintResource()
	log.Infof("")
	jobClient.SetInitialState()

	if err != nil {
		log.Errorf("Error updating job %s: %s", j, err)
		log.Errorf("Job %s will not be processed", j)
		return
	}
	p.runningJobs.Store(j, jobClient.GetResource())
	defer p.runningJobs.Delete(j)

	requestedAction := jobClient.GetActionName()
	if requestedAction == "" {
		log.Errorf("Job %s has no action, thus cannot be started", j)
		// TODO: Should we here update the job status to failed/finished/...?
		return
	}
	log.Debugf("Job %s has requested action %s", j, requestedAction)
	action := jobEngine.NewAction(
		requestedAction,
		jobEngine.WithNuvlaClient(p.client),
		jobEngine.WithCoeClient(p.coe),
		jobEngine.WithJobResource(jobClient.GetResource()))

	if action == nil {
		log.Warnf("Error creating action %s for job %s", requestedAction, j)
		// Set success state to close job
		jobClient.SetSuccessState()
		log.Warnf("Action %s not supported (yet), passing... ", requestedAction)
		return
	}

	log.Infof("Starting action %s for job %s", requestedAction, j)
	err = action.Execute()
	if err != nil {
		log.Errorf("Error executing action %s for job %s: %s", requestedAction, j, err)
		// TODO: Report job as failed
		jobClient.SetState(clients.StateFailed)
		return
	}

	// 2. If the job is correct, add it to the running jobs and start it
	// Remove the job from the running jobs
	jobClient.SetSuccessState()
	log.Infof("Job %s finished successfully", j)
}

func (p *JobProcessor) stopJob(j string) {

}

func (p *JobProcessor) getJob(j string) *Job {
	return nil
}

func (p *JobProcessor) getRunningJobs() []string {
	return nil
}
