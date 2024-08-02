package orchestrator

import (
	"context"
	"nuvlaedge-go/types"
)

type Orchestrator interface {
	Start(ctx context.Context, opts *types.StartOpts) error
	Stop(ctx context.Context, opts *types.StopOpts) error
	Install(ctx context.Context) error
	Remove(ctx context.Context) error

	// Close should close any client connection or dangling resource. It is specially important on docker
	Close() error
}
