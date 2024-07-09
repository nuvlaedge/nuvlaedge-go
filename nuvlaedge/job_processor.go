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
	jobChan        chan string        // NativeJob channel. Receives jobs IDs from the agent
	exitChan       chan bool          // Exit channel. Receives exit signal from the agent
	client         *nuvla.NuvlaClient // Nuvla session required in the jobs and deployment clients
	coe            orchestrator.Coe   // COE client required in the jobs and deployment clients
	enableLegacy   bool
	legacyJobImage string

	runningJobs *JobRegistry
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
		runningJobs:    NewRunningJobs(),
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

	if p.runningJobs.Exists(j) {
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

	ok := p.runningJobs.Add(&RunningJob{
		jobId:   j,
		jobType: job.GetJobType(),
		running: true,
	})
	if !ok {
		log.Errorf("Job %s is already running...", j)
		return
	}
	log.Infof("Currently running jobs: \n %s", p.runningJobs)
	defer p.runningJobs.Remove(j)

	// 2. Run the jobs
	err = job.RunJob()
	if err != nil {
		log.Errorf("Error running job %s: %s", j, err)
		return
	}
	log.Infof("Running job %s... Success.", j)

}

type RunningJob struct {
	jobId   string
	jobType string
	running bool
}

type JobRegistry struct {
	jobs map[string]*RunningJob
	lock *sync.Mutex
}

func NewRunningJobs() *JobRegistry {
	return &JobRegistry{
		jobs: make(map[string]*RunningJob),
		lock: &sync.Mutex{},
	}
}

func (r *JobRegistry) Add(job *RunningJob) bool {
	if r.Exists(job.jobId) {
		return false
	}
	r.lock.Lock()
	defer r.lock.Unlock()

	r.jobs[job.jobId] = job
	return true
}

func (r *JobRegistry) Remove(jobId string) bool {
	if !r.Exists(jobId) {
		return false
	}
	r.lock.Lock()
	defer r.lock.Unlock()

	delete(r.jobs, jobId)
	return true
}

func (r *JobRegistry) Get(jobId string) (*RunningJob, bool) {
	r.lock.Lock()
	defer r.lock.Unlock()
	job, ok := r.jobs[jobId]
	return job, ok
}

func (r *JobRegistry) Exists(jobId string) bool {
	r.lock.Lock()
	defer r.lock.Unlock()
	_, ok := r.jobs[jobId]
	return ok
}

func (r *JobRegistry) String() string {
	r.lock.Lock()
	defer r.lock.Unlock()
	var jobSummary string
	for _, job := range r.jobs {
		jobSummary += "ID: " + job.jobId + " Type: " + job.jobType + "\n"
	}
	return jobSummary
}
