package nuvlaedge

import (
	nuvla "github.com/nuvla/api-client-go"
	"github.com/nuvla/api-client-go/types"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/nuvlaedge/jobProcessor"
	"nuvlaedge-go/nuvlaedge/orchestrator"
	"sync"
)

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
	// Send exit signal to the job processor
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

	log.Infof("Job Processor starting new job with id %s", j)

	// 1. Create Job structure
	job := jobProcessor.NewJob(types.NewNuvlaIDFromId(j), p.client, p.coe)
	p.runningJobs.Store(j, job)
	defer p.runningJobs.Delete(j)

	log.Debugf("Job created: %s", job.String())

	log.Infof("Initialising job... %s...", j)
	err := job.Start()
	if err != nil {
		log.Errorf("Error starting job %s: %s", j, err)
		return
	}
	log.Infof("Initialising job %s... Success.", j)

	// 2. Run the job
	log.Infof("Running job %s...", j)
	err = job.Run()
	if err != nil {
		log.Errorf("Error running job %s: %s", j, err)
		return
	}
	log.Infof("Running job %s... Success.", j)

}
