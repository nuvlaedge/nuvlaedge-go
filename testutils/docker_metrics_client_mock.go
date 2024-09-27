package testutils

import (
	"context"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/api/types/system"
	"github.com/docker/docker/api/types/volume"
)

// TestDockerMetricsClient implements DockerMetricsClient for testing purposes.
type TestDockerMetricsClient struct {
	SwarmInspectMu    sync.Mutex
	SwarmInspectCount int
	InspectReturn     swarm.Swarm
	InspectErr        error

	InfoMu     sync.Mutex
	InfoCount  int
	InfoReturn system.Info
	InfoErr    error

	NodeListMu     sync.Mutex
	NodeListCount  int
	NodeListReturn []swarm.Node
	NodeListErr    error

	PluginListMu     sync.Mutex
	PluginListCount  int
	PluginListReturn types.PluginsListResponse
	PluginListErr    error

	ContainerListMu     sync.Mutex
	ContainerListCount  int
	ContainerListReturn []types.Container
	ContainerListErr    error

	ContainerStatsMu     sync.Mutex
	ContainerStatsCount  int
	ContainerStatsReturn container.StatsResponseReader
	ContainerStatsErr    error

	ContainerInspectMu      sync.Mutex
	ContainerInspectCount   int
	ContainerInspectReturn  types.ContainerJSON
	ContainerInspectErr     error
	ContainerInspectErrFunc func(containerID string) error

	ImageListMu     sync.Mutex
	ImageListCount  int
	ImageListReturn []image.Summary
	ImageListErr    error

	VolumeListMu     sync.Mutex
	VolumeListCount  int
	VolumeListReturn volume.ListResponse
	VolumeListErr    error

	NetworkListMu     sync.Mutex
	NetworkListCount  int
	NetworkListReturn []network.Summary
	NetworkListErr    error

	ServiceListMu     sync.Mutex
	ServiceListCount  int
	ServiceListReturn []swarm.Service
	ServiceListErr    error

	TaskListMu     sync.Mutex
	TaskListCount  int
	TaskListReturn []swarm.Task
	TaskListErr    error

	ConfigListMu     sync.Mutex
	ConfigListCount  int
	ConfigListReturn []swarm.Config
	ConfigListErr    error

	SecretListMu     sync.Mutex
	SecretListCount  int
	SecretListReturn []swarm.Secret
	SecretListErr    error

	CloseMu    sync.Mutex
	CloseCount int
	CloseErr   error
}

func (c *TestDockerMetricsClient) SwarmInspect(ctx context.Context) (swarm.Swarm, error) {
	c.SwarmInspectMu.Lock()
	defer c.SwarmInspectMu.Unlock()
	c.SwarmInspectCount++
	return c.InspectReturn, c.InspectErr
}

func (c *TestDockerMetricsClient) Info(ctx context.Context) (system.Info, error) {
	c.InfoMu.Lock()
	defer c.InfoMu.Unlock()
	c.InfoCount++
	return c.InfoReturn, c.InfoErr
}

func (c *TestDockerMetricsClient) NodeList(ctx context.Context, options types.NodeListOptions) ([]swarm.Node, error) {
	c.NodeListMu.Lock()
	defer c.NodeListMu.Unlock()
	c.NodeListCount++
	return c.NodeListReturn, c.NodeListErr
}

func (c *TestDockerMetricsClient) PluginList(ctx context.Context, args filters.Args) (types.PluginsListResponse, error) {
	c.PluginListMu.Lock()
	defer c.PluginListMu.Unlock()
	c.PluginListCount++
	return c.PluginListReturn, c.PluginListErr
}

func (c *TestDockerMetricsClient) ContainerList(ctx context.Context, options container.ListOptions) ([]types.Container, error) {
	c.ContainerListMu.Lock()
	defer c.ContainerListMu.Unlock()
	c.ContainerListCount++
	return c.ContainerListReturn, c.ContainerListErr
}

func (c *TestDockerMetricsClient) ContainerStats(ctx context.Context, containerID string, stream bool) (container.StatsResponseReader, error) {
	c.ContainerStatsMu.Lock()
	defer c.ContainerStatsMu.Unlock()
	c.ContainerStatsCount++
	return c.ContainerStatsReturn, c.ContainerStatsErr
}

func (c *TestDockerMetricsClient) ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	c.ContainerInspectMu.Lock()
	defer c.ContainerInspectMu.Unlock()
	c.ContainerInspectCount++
	if c.ContainerInspectErrFunc != nil {
		return types.ContainerJSON{}, c.ContainerInspectErrFunc(containerID)
	}
	return c.ContainerInspectReturn, c.ContainerInspectErr
}

func (c *TestDockerMetricsClient) ImageList(ctx context.Context, opts image.ListOptions) ([]image.Summary, error) {
	c.ImageListMu.Lock()
	defer c.ImageListMu.Unlock()
	c.ImageListCount++
	return c.ImageListReturn, c.ImageListErr
}

func (c *TestDockerMetricsClient) VolumeList(ctx context.Context, opts volume.ListOptions) (volume.ListResponse, error) {
	c.VolumeListMu.Lock()
	defer c.VolumeListMu.Unlock()
	c.VolumeListCount++
	return c.VolumeListReturn, c.VolumeListErr
}

func (c *TestDockerMetricsClient) NetworkList(ctx context.Context, opts network.ListOptions) ([]types.NetworkResource, error) {
	c.NetworkListMu.Lock()
	defer c.NetworkListMu.Unlock()
	c.NetworkListCount++
	return c.NetworkListReturn, c.NetworkListErr
}

func (c *TestDockerMetricsClient) ServiceList(ctx context.Context, opts types.ServiceListOptions) ([]swarm.Service, error) {
	c.ServiceListMu.Lock()
	defer c.ServiceListMu.Unlock()
	c.ServiceListCount++
	return c.ServiceListReturn, c.ServiceListErr
}

func (c *TestDockerMetricsClient) TaskList(ctx context.Context, opts types.TaskListOptions) ([]swarm.Task, error) {
	c.TaskListMu.Lock()
	defer c.TaskListMu.Unlock()
	c.TaskListCount++
	return c.TaskListReturn, c.TaskListErr
}

func (c *TestDockerMetricsClient) ConfigList(ctx context.Context, opts types.ConfigListOptions) ([]swarm.Config, error) {
	c.ConfigListMu.Lock()
	defer c.ConfigListMu.Unlock()
	c.ConfigListCount++
	return c.ConfigListReturn, c.ConfigListErr
}

func (c *TestDockerMetricsClient) SecretList(ctx context.Context, opts types.SecretListOptions) ([]swarm.Secret, error) {
	c.SecretListMu.Lock()
	defer c.SecretListMu.Unlock()
	c.SecretListCount++
	return c.SecretListReturn, c.SecretListErr
}

func (c *TestDockerMetricsClient) Close() error {
	c.CloseMu.Lock()
	defer c.CloseMu.Unlock()
	c.CloseCount++
	return c.CloseErr
}
