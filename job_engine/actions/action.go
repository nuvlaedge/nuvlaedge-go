package actions

import (
	"context"
	"errors"
	"nuvlaedge-go/types/jobs"
)

type Action interface {
	Execute(ctx context.Context) error
	Init(opts *ActionOpts) error
}

type ActionOpts struct {
	Job            *jobs.JobBase
	LegacyJobImage string
}

func ActionFactory(actionType string) Action {
	switch actionType {
	case "nuvlabox_update":
		return &UpdateAction{}
	case "reboot_nuvlabox":
		return &RebootAction{}
	case "legacy_job":
		return &LegacyJobAction{}
	default:
		return nil
	}
}

func RunJob(ctx context.Context, job *jobs.JobBase, legacyImage string) error {
	action := ActionFactory(job.JobType)
	if action == nil {
		return errors.New("unknown action")
	}

	if err := action.Init(&ActionOpts{Job: job, LegacyJobImage: legacyImage}); err != nil {
		return err
	}

	return action.Execute(ctx)
}
