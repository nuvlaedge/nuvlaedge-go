package job_processor

import (
	"context"
	nuvla "github.com/nuvla/api-client-go"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/engine"
	"nuvlaedge-go/types/jobs"
	"nuvlaedge-go/types/worker"
	"sync"
)

type JobProcessor struct {
	worker.WorkerBase
	jobChan        chan string        // NativeJob channel. Receive jobs IDs from the agent
	client         *nuvla.NuvlaClient // Nuvla session required in the jobs and deployment clients
	coe            engine.Coe         // COE client required in the jobs and deployment clients
	enableLegacy   bool
	legacyJobImage string

	runningJobs *jobs.JobRegistry
}

func (p *JobProcessor) Start(ctx context.Context) error {
	log.Infof("Nothing to start in the jobs processor, passing...")
	go func() {
		err := p.Run(ctx)
		if err != nil {
			log.Errorf("Error running Job Processor: %s", err)
		}
	}()
	return nil
}

func (p *JobProcessor) Stop(ctx context.Context) error {
	// Send exit signal to the jobs processor
	log.Info("Stopping NativeJob Processor")

	return nil
}

func (p *JobProcessor) Init(opts *worker.WorkerOpts, conf *worker.WorkerConfig) error {
	p.WorkerBase = worker.NewWorkerBase(worker.JobProcessor)
	p.jobChan = opts.JobCh
	p.runningJobs = opts.Jobs

	// Clients setup
	p.client = opts.NuvlaClient.NuvlaClient

	// Config
	p.enableLegacy = conf.EnableJobLegacy
	p.legacyJobImage = conf.LegacyJobImage
	p.coe = engine.NewDockerEngine()
	return nil
}

func (p *JobProcessor) Reconfigure(conf *worker.WorkerConfig) error {
	p.legacyJobImage = conf.LegacyJobImage
	p.enableLegacy = conf.EnableJobLegacy
	return nil
}

func (p *JobProcessor) Run(ctx context.Context) error {
	log.Info("Running Job Engine")

	for {
		select {
		case job := <-p.jobChan:
			go p.processJob(ctx, job)
		case <-ctx.Done():
			log.Info("Context done. Exiting...")
			return ctx.Err()
		case conf := <-p.ConfChan:
			log.Debug("Received configuration in Job Processor: ", conf)
			if err := p.Reconfigure(conf); err != nil {
				log.Error("Failed to reconfigure job processor: ", err)
			}
		}

	}
}

func (p *JobProcessor) processJob(ctx context.Context, j string) {
	jobCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	if p.runningJobs.Exists(j) {
		log.Infof("NativeJob %s is already running", j)
		return
	}

	log.Infof("NativeJob Processor starting new jobs with id %s", j)

	// 1. Create NativeJob structure
	job, err := NewJob(jobCtx, j, p.client, p.coe, p.enableLegacy, p.legacyJobImage)
	if err != nil {
		log.Errorf("Error creating job %s: %s", j, err)
		return
	}

	ok := p.runningJobs.Add(&jobs.RunningJob{
		JobId:   j,
		JobType: job.GetJobType(),
	})
	if !ok {
		log.Errorf("Job %s is already running...", j)
		return
	}
	log.Infof("Currently running jobs: \n %s", p.runningJobs)
	defer p.runningJobs.Remove(j)

	// 2. Run the jobs
	err = job.RunJob(jobCtx)
	if err != nil {
		log.Errorf("Error running job %s: %s", j, err)
		return
	}
	log.Infof("Running job %s... Success.", j)

}

type RunningJob struct {
	jobId   string
	jobType string
}

type JobRegistry struct {
	jobs map[string]*RunningJob
	lock *sync.Mutex
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

var _ worker.Worker = &JobProcessor{}
