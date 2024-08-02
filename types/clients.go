package types

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/api/types/system"
	"github.com/nuvla/api-client-go/clients"
	nuvlaTypes "github.com/nuvla/api-client-go/types"
	"net/http"
)

type TelemetryClientInterface interface {
	Telemetry(data map[string]interface{}, Select []string) (*http.Response, error)
	GetEndpoint() string
}

type TelemetryClient struct {
	*clients.NuvlaEdgeClient
}

func (tc *TelemetryClient) GetEndpoint() string {
	return tc.SessionOpts.Endpoint
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

type CommissionClientInterface interface {
	Get(id string, selectFields []string) (*nuvlaTypes.NuvlaResource, error)
	Commission(data map[string]interface{}) error
	GetStatusId() string
}

type CommissionClient struct {
	*clients.NuvlaEdgeClient
}

func (cc *CommissionClient) GetStatusId() string {
	if cc.NuvlaEdgeStatusId == nil {
		return ""
	}
	return cc.NuvlaEdgeStatusId.String()
}
