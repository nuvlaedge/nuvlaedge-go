package nuvlaedge

import (
	nuvla "github.com/nuvla/api-client-go"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/nuvlaedge/jobs"
	"nuvlaedge-go/nuvlaedge/orchestrator"
	"sync"
)

type JobProcessor struct {
	runningJobs sync.Map
	jobChan     chan string        // Job channel. Receives jobs IDs from the agent
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
	log.Infof("Nothing to start in the jobs processor, passing...")
	return nil
}

func (p *JobProcessor) Stop() error {
	// Send exit signal to the jobs processor
	log.Info("Stopping Job Processor")
	p.exitChan <- true
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

	log.Infof("Job Processor starting new jobs with id %s", j)

	// 1. Create Job structure
	job := jobs.NewJob(j, p.client)
	p.runningJobs.Store(j, job)
	defer p.runningJobs.Delete(j)

	log.Debugf("Job created: %s", job.JobId)

	log.Infof("Initialising jobs... %s...", j)
	err := job.Init()
	if err != nil {
		log.Errorf("Error starting jobs %s: %s", j, err)
		return
	}
	log.Infof("Initialising jobs %s... Success.", j)

	// 2. Run the jobs
	log.Infof("Running jobs %s...", j)
	err = job.Run()
	if err != nil {
		log.Errorf("Error running jobs %s: %s", j, err)
		return
	}
	log.Infof("Running jobs %s... Success.", j)

}
