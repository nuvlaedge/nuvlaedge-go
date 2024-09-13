package types

import (
	"context"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/compose/v2/pkg/api"
)

type ComposeService interface {
	Up(ctx context.Context, p *types.Project, opts api.UpOptions) error
	Down(ctx context.Context, p string, opts api.DownOptions) error
	Ps(ctx context.Context, p string, opts api.PsOptions) ([]api.ContainerSummary, error)
	Pull(ctx context.Context, p *types.Project, opts api.PullOptions) error
	List(ctx context.Context, opts api.ListOptions) ([]api.Stack, error)
	Logs(ctx context.Context, projectName string, consumer api.LogConsumer, options api.LogOptions) error
}

// StartOpts are shared options for any orchestrator start operation, thus, each start will access its required fields
type StartOpts struct {
	// Compose Opts
	CFiles      []string
	Env         []string
	ProjectName string
	WorkingDir  string
}

// StopOpts are shared options for any orchestrator stop operation, thus, each stop will access its required fields
type StopOpts struct {
	// Compose Opts
	ProjectName string
}

type LogOpts struct {
	ProjectName string
	LogConsumer api.LogConsumer
	LogOptions  api.LogOptions
}
