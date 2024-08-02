package actions

import (
	"context"
	"nuvlaedge-go/engine"
	"nuvlaedge-go/types"
)

type RebootJob struct {
	JobBase

	// CE is the container engine
	ce engine.ContainerEngine
}

func (rj *RebootJob) RunJob(ctx context.Context) error {
	// This is where the actual job is run
	return nil
}

func (rj *RebootJob) Init(opts types.JobOpts) error {
	rj.ce = opts.Ce
	return nil
}

// Compile time check if RebootJob implements Job interface
var _ Job = &RebootJob{}
