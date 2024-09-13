package testutils

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/api/types/system"
)

// TestDockerMetricsClient implements DockerMetricsClient for testing purposes.
type TestDockerMetricsClient struct {
	SwarmInspectCount int
	InspectReturn     swarm.Swarm
	InspectErr        error

	InfoCount  int
	InfoReturn system.Info
	InfoErr    error

	NodeListCount  int
	NodeListReturn []swarm.Node
	NodeListErr    error

	PluginListCount  int
	PluginListReturn types.PluginsListResponse
	PluginListErr    error

	ContainerListCount  int
	ContainerListReturn []types.Container
	ContainerListErr    error

	ContainerStatsCount  int
	ContainerStatsReturn container.StatsResponseReader
	ContainerStatsErr    error

	ContainerInspectCount   int
	ContainerInspectReturn  types.ContainerJSON
	ContainerInspectErr     error
	ContainerInspectErrFunc func(containerID string) error

	CloseCount int
	CloseErr   error
}

func (c *TestDockerMetricsClient) SwarmInspect(ctx context.Context) (swarm.Swarm, error) {
	c.SwarmInspectCount++
	return c.InspectReturn, c.InspectErr
}

func (c *TestDockerMetricsClient) Info(ctx context.Context) (system.Info, error) {
	c.InfoCount++
	return c.InfoReturn, c.InfoErr
}

func (c *TestDockerMetricsClient) NodeList(ctx context.Context, options types.NodeListOptions) ([]swarm.Node, error) {
	c.NodeListCount++
	return c.NodeListReturn, c.NodeListErr
}

func (c *TestDockerMetricsClient) PluginList(ctx context.Context, args filters.Args) (types.PluginsListResponse, error) {
	c.PluginListCount++
	return c.PluginListReturn, c.PluginListErr
}

func (c *TestDockerMetricsClient) ContainerList(ctx context.Context, options container.ListOptions) ([]types.Container, error) {
	c.ContainerListCount++
	return c.ContainerListReturn, c.ContainerListErr
}

func (c *TestDockerMetricsClient) ContainerStats(ctx context.Context, containerID string, stream bool) (container.StatsResponseReader, error) {
	c.ContainerStatsCount++
	return c.ContainerStatsReturn, c.ContainerStatsErr
}

func (c *TestDockerMetricsClient) ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	c.ContainerInspectCount++
	if c.ContainerInspectErrFunc != nil {
		return types.ContainerJSON{}, c.ContainerInspectErrFunc(containerID)
	}
	return c.ContainerInspectReturn, c.ContainerInspectErr
}

func (c *TestDockerMetricsClient) Close() error {
	// Implement close logic if needed.
	c.CloseCount++
	return c.CloseErr
}
