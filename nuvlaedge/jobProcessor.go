package nuvlaedge

type Job struct {
	ID          string
	containerID string
}

type JobProcessor struct {
	runningJobs []string
	jobChan     chan []string // Job channel. Receives job IDs from the agent
	exitChan    chan bool     // Exit channel. Receives exit signal from the agent
}

func NewJobProcessor(jobChan chan []string) *JobProcessor {
	return &JobProcessor{
		jobChan: jobChan,
	}
}

func (p *JobProcessor) Start() error {
	return nil
}

func (p *JobProcessor) Stop() error {
	return nil
}

func (p *JobProcessor) Run() error {
	for {
		select {
		case jobs := <-p.jobChan:
			for _, j := range jobs {
				go p.processJob(j)
			}
		case <-p.exitChan:
			return nil
		}
	}
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
