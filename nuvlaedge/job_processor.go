package nuvlaedge

import (
	"context"
	nuvla "github.com/nuvla/api-client-go"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/nuvlaedge/jobs"
	"nuvlaedge-go/nuvlaedge/orchestrator"
	"sync"
)

type JobProcessor struct {
	ctx            context.Context
	runningJobs    sync.Map
	jobChan        chan string        // NativeJob channel. Receives jobs IDs from the agent
	exitChan       chan bool          // Exit channel. Receives exit signal from the agent
	client         *nuvla.NuvlaClient // Nuvla session required in the jobs and deployment clients
	coe            orchestrator.Coe   // COE client required in the jobs and deployment clients
	enableLegacy   bool
	legacyJobImage string
}

func NewJobProcessor(
	ctx context.Context,
	jobChan chan string,
	client *nuvla.NuvlaClient,
	coe orchestrator.Coe,
	enableLegacy bool,
	legacyImage string) *JobProcessor {
	return &JobProcessor{
		ctx:            ctx,
		jobChan:        jobChan,
		client:         client,
		coe:            coe,
		enableLegacy:   enableLegacy,
		legacyJobImage: legacyImage,
		runningJobs:    sync.Map{},
	}
}

func (p *JobProcessor) Start() error {
	log.Infof("Nothing to start in the jobs processor, passing...")
	return nil
}

func (p *JobProcessor) Stop() error {
	// Send exit signal to the jobs processor
	log.Info("Stopping NativeJob Processor")
	p.exitChan <- true
	return nil
}

func (p *JobProcessor) Run() error {
	log.Info("Running NativeJob Engine")
	go func() {
		for {
			select {
			case job := <-p.jobChan:
				go p.processJob(job)
			case <-p.ctx.Done():
				log.Warn("NativeJob Processor received exit signal")
				return
			}
		}
	}()
	return nil
}

func (p *JobProcessor) processJob(j string) {
	if _, ok := p.runningJobs.Load(j); ok {
		log.Infof("NativeJob %s is already running", j)
		return
	}

	log.Infof("NativeJob Processor starting new jobs with id %s", j)

	// 1. Create NativeJob structure
	job, err := jobs.NewJob(j, p.client, p.coe, p.enableLegacy, p.legacyJobImage)
	if err != nil {
		log.Errorf("Error creating job %s: %s", j, err)
		return
	}
	p.runningJobs.Store(j, job)
	defer p.runningJobs.Delete(j)

	// 2. Run the jobs
	log.Infof("Running jobs %s...", j)
	err = job.RunJob()
	if err != nil {
		log.Errorf("Error running job %s: %s", j, err)
		return
	}
	log.Infof("Running jobs %s... Success.", j)

}
