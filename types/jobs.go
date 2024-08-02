package types

import (
	"nuvlaedge-go/engine"
	"sync"
)

type JobRegistry struct {
	jobs map[string]*RunningJob
	lock *sync.Mutex
}

func NewRunningJobs() JobRegistry {
	return JobRegistry{
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
		jobSummary += "ID: " + job.jobId + " Type: " + string(job.jobType) + "\n"
	}
	return jobSummary
}

type RunningJob struct {
	jobId   string
	jobType JobType
	running bool
}

type JobType string

func (jt JobType) String() string {
	return string(jt)
}

const (
	RebootJob           JobType = "reboot_nuvlabox"
	StopDeploymentJob   JobType = "stop_deployment"
	StartDeploymentJob  JobType = "start_deployment"
	StateDeploymentJob  JobType = "deployment_state"
	UpdateDeploymentJob JobType = "update_deployment"
	UpdateNuvlaEdgeJob  JobType = "nuvlabox_update"
	UnknownJob          JobType = "unknown"
)

var SupportedJobTypes = []JobType{
	RebootJob,
	StartDeploymentJob,
	StateDeploymentJob,
	StopDeploymentJob,
	UpdateDeploymentJob,
	UpdateNuvlaEdgeJob,
}

type JobOpts struct {
	JobId string
	Ce    engine.ContainerEngine
}
