package actions

import (
	"context"
	"errors"
	"nuvlaedge-go/common"
	"nuvlaedge-go/types/jobs"
)

type LegacyJobAction struct {
	job         *jobs.JobBase
	legacyImage string
}

func (lja *LegacyJobAction) legacyJobDocker(ctx context.Context) error {
	//connector.connectorNewDockerEngine()
	return nil
}

func (lja *LegacyJobAction) legacyJobKubernetes(ctx context.Context) error {
	return nil
}

func (lja *LegacyJobAction) Execute(ctx context.Context) error {
	switch common.WhereAmI() {
	case common.Docker:
		return lja.legacyJobDocker(ctx)
	case common.Kubernetes:
		return lja.legacyJobKubernetes(ctx)
	case common.Host:
		return lja.legacyJobDocker(ctx)
	default:
		return errors.New("unknown environment")
	}
}

func (lja *LegacyJobAction) Init(opts *ActionOpts) error {
	lja.job = opts.Job
	lja.legacyImage = opts.LegacyJobImage
	return nil
}
