package nuvlaedge

type Job struct {
	ID          string
	containerID string
}

type JobProcessor struct {
	runningJobs []string
	jobChan     chan string // Job channel. Receives job IDs from the agent
	exitChan    chan bool   // Exit channel. Receives exit signal from the agent
}

func NewJobProcessor(jobChan chan string) *JobProcessor {
	return &JobProcessor{
		jobChan: jobChan,
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
	log.Infof("Running Job Engine")
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
	log.Infof("Job Processor starting new job with id %s", j)

}

func (p *JobProcessor) stopJob(j string) {

}

func (p *JobProcessor) getJob(j string) *Job {
	return nil
}

func (p *JobProcessor) getRunningJobs() []string {
	return nil
}
