package jobs

import (
	"context"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/common"
	"nuvlaedge-go/types/jobs"
)

func Reboot(ctx context.Context, opts *jobs.JobOpts) error {
	ex := assertExecutor(opts)
	if ex == nil {
		return errors.New("can't determine where the program is running")
	}

	err := ex.Reboot(ctx)
	if err != nil {
		return err
	}
	return nil
}

// FetchLogs action
func FetchLogs(ctx context.Context, opts *jobs.JobOpts) error {
	ex := assertExecutor(opts)
	if ex == nil {
		return errors.New("can't determine where the program is running")
	}

	return ex.LogFetch(ctx)
}

// Update NuvlaEdge action
func Update(ctx context.Context, opts *jobs.JobOpts) error {
	return nil
}

// AddSSHKey action
func AddSSHKey(ctx context.Context, opts *jobs.JobOpts) error {
	ex := assertExecutor(opts)
	if ex == nil {
		return errors.New("can't determine where the program is running")
	}

	return ex.AddSSHKey(ctx)
}

// RevokeSSHKey action
func RevokeSSHKey(ctx context.Context, opts *jobs.JobOpts) error {
	ex := assertExecutor(opts)
	if ex == nil {
		return errors.New("can't determine where the program is running")
	}

	return ex.RevokeSSHKey(ctx)
}

func assertExecutor(opts *jobs.JobOpts) jobs.JobExecutor {
	var ex jobs.JobExecutor

	switch common.WhereAmI() {
	case common.Docker:
		ex = opts.ContainerEx
	case common.Kubernetes:
		ex = opts.ContainerEx
	case common.Host:
		ex = opts.HostEx
	default:
		log.Info("Should not be here")
		return nil
	}

	return ex
}

func ActionFactory(actionType string) jobs.JobAction {
	switch actionType {
	case jobs.RebootJob:
		return Reboot
	case jobs.FetchLogsJob:
		return FetchLogs
	case jobs.UpdateNuvlaEdgeJob:
		return Update
	case jobs.AddSSHKeyJob:
		return AddSSHKey
	case jobs.RevokeSSHKeyJob:
		return RevokeSSHKey
	default:
		return nil
	}
}

func RunJob(ctx context.Context, opts *jobs.JobOpts) error {
	j := opts.Job
	action := ActionFactory(j.JobType)

	if action == nil {
		errMsg := fmt.Sprintf("Action %s is not supported", j.JobType)
		j.Client.SetFailedState(errMsg)
		return fmt.Errorf(errMsg)
	}

	log.Infof("Running job %s", j.JobId)
	j.Client.SetInitialState()

	if err := action(ctx, opts); err != nil {
		j.Client.SetFailedState(err.Error())
		return err
	}

	j.Client.SetSuccessState()

	return nil
}
