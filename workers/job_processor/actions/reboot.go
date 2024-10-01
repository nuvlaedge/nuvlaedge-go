package actions

import (
	"context"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/workers/job_processor/executors"
)

// RebootAction reboots the system. It can do it from a container or from the host.
// The reboot action should be scheduled to allow the jobs to complete.
type RebootAction struct {
	ActionBase

	executor executors.Rebooter
}

func (r *RebootAction) ExecuteAction(_ context.Context) error {
	err := r.executor.Reboot()
	if err != nil {
		log.Errorf("Error executing reboot: %s", err)
		return err
	}
	// TODO: Reboot is scheduled to allow the jobs to complete
	log.Infof("Reboot successfully scheduled")
	return nil
}

func (r *RebootAction) GetExecutorName() executors.ExecutorName {
	return r.executor.GetName()
}

func (r *RebootAction) assertExecutor() error {
	ex, err := executors.GetRebooter(true)
	if err != nil {
		return err
	}
	r.executor = ex
	log.Infof("Reboot action executor set to: %s", r.GetExecutorName())
	return nil
}

func (r *RebootAction) Init(_ context.Context, _ ...ActionOptsFn) error {
	err := r.assertExecutor()
	if err != nil {
		return err
	}
	log.Infof("Reboot actions initialised with executor: %s", r.GetExecutorName())
	return nil
}

func (r *RebootAction) GetOutput() string {
	return ""
}
