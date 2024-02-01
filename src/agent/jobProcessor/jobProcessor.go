package jobProcessor

import log "github.com/sirupsen/logrus"

type JobProcessor struct {
	JobIncome   chan string
	runningJobs []string
}

func New(inChannel chan string) *JobProcessor {
	return &JobProcessor{
		JobIncome: inChannel,
	}
}

func (p *JobProcessor) run() {
	for j := range p.JobIncome {
		log.Infof("Job Processor starting new job with id %s", j)
	}
}
