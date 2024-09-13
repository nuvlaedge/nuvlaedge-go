package types

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/api/types/system"
	"github.com/docker/docker/api/types/volume"
)

type CleanerClient interface {
	ContainersPrune(ctx context.Context, args filters.Args) (container.PruneReport, error)
	ImagesPrune(ctx context.Context, args filters.Args) (image.PruneReport, error)
	VolumesPrune(ctx context.Context, args filters.Args) (volume.PruneReport, error)
	NetworksPrune(ctx context.Context, args filters.Args) (network.PruneReport, error)
}

type DockerMetricsClient interface {
	SwarmInspect(ctx context.Context) (swarm.Swarm, error)
	Info(ctx context.Context) (system.Info, error)
	NodeList(ctx context.Context, options types.NodeListOptions) ([]swarm.Node, error)
	PluginList(ctx context.Context, args filters.Args) (types.PluginsListResponse, error)
	Close() error
	ContainerList(ctx context.Context, options container.ListOptions) ([]types.Container, error)
	ContainerStats(ctx context.Context, containerID string, stream bool) (container.StatsResponseReader, error)
	ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error)
}

type InstallationParametersClient interface {
	ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error)
}