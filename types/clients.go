package types

import (
	"context"
	"github.com/nuvla/api-client-go/clients"
	"github.com/nuvla/api-client-go/clients/resources"
	nuvlaTypes "github.com/nuvla/api-client-go/types"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type TelemetryClientInterface interface {
	Telemetry(ctx context.Context, data interface{}, selects []string) (*http.Response, error)
	GetEndpoint() string
}

type TelemetryClient struct {
	*clients.NuvlaEdgeClient
}

func (tc *TelemetryClient) GetEndpoint() string {
	return tc.SessionOpts.Endpoint
}

type CommissionClientInterface interface {
	Get(ctx context.Context, id string, selectFields []string) (*nuvlaTypes.NuvlaResource, error)
	Commission(ctx context.Context, data map[string]interface{}) error
	GetStatusId() string
}

type CommissionClient struct {
	*clients.NuvlaEdgeClient
}

func (cc *CommissionClient) GetStatusId() string {
	neRes := cc.GetNuvlaEdgeResource()
	if neRes.NuvlaBoxStatus == "" {

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		err := cc.UpdateResourceSelect(ctx, []string{"nuvlabox-status"})
		if err != nil {
			log.Error("Failed to update nuvlabox status ID", err)
			return ""
		}
	}
	return cc.GetNuvlaEdgeResource().NuvlaBoxStatus
}

type ConfUpdaterClient interface {
	UpdateResourceSelect(ctx context.Context, selects []string) error
	GetNuvlaEdgeResource() resources.NuvlaEdgeResource
}

//go:generate mockery --name HeartbeatClient
type HeartbeatClient interface {
	Heartbeat(ctx context.Context) (*http.Response, error)
}
