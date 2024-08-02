package engine

import (
	"context"
)

// ContainerEngine is an interface for the container engine
// All methods should take context as the first argument to allow for cancellation since it is a blocking operation
type ContainerEngine interface {
	// GetLogs returns the logs of a container
	GetLogs(ctx context.Context, containerId string) (string, error) // Probably should take a
	// RunContainer runs a container
	RunContainer(ctx context.Context) error
	// StopContainer stops a container
	StopContainer(ctx context.Context, containerId string) error
	// RemoveContainer removes a container
	RemoveContainer(ctx context.Context, containerId string) error

	// Prune removes all stopped containers, unused networks, dangling images, and build cache
	Prune(ctx context.Context) error
}
