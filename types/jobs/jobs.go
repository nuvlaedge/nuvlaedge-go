package jobs

import (
	"slices"
	"strings"
	"sync"
)

type Job interface {
	GetId() string
	GetJobType() string
}

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
	if r.Exists(job.JobId) {
		return false
	}
	r.lock.Lock()
	defer r.lock.Unlock()

	r.jobs[job.JobId] = job
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
		jobSummary += "ID: " + job.JobId + " Type: " + job.JobType + "\n"
	}
	return jobSummary
}

type RunningJob struct {
	JobId   string
	JobType string
}

const (
	RebootJob           string = "reboot_nuvlabox"
	StopDeploymentJob   string = "stop_deployment"
	StartDeploymentJob  string = "start_deployment"
	StateDeploymentJob  string = "deployment_state"
	UpdateDeploymentJob string = "update_deployment"
	UpdateNuvlaEdgeJob  string = "nuvlabox_update"
	FetchLogsJob        string = "fetch_nuvlabox_log"
	AddSSHKeyJob        string = "add_ssh_key"
	RevokeSSHKeyJob     string = "revoke_ssh_key"
	UnknownJob          string = "unknown"
)

var SupportedJobTypes = []string{
	RebootJob,
	FetchLogsJob,
	StartDeploymentJob,
	StateDeploymentJob,
	StopDeploymentJob,
	UpdateDeploymentJob,
	UpdateNuvlaEdgeJob,
}

func IsSupportedJob(jobType string) bool {
	return slices.Contains(SupportedJobTypes, jobType)
}

func IsDeployment(jobType string) bool {
	return strings.Contains(jobType, "deployment")
}
