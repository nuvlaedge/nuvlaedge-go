package types

import (
	"github.com/nuvla/api-client-go/clients"
	"github.com/nuvla/api-client-go/clients/resources"
	nuvlaTypes "github.com/nuvla/api-client-go/types"
	log "github.com/sirupsen/logrus"
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

type CommissionClientInterface interface {
	Get(id string, selectFields []string) (*nuvlaTypes.NuvlaResource, error)
	Commission(data map[string]interface{}) error
	GetStatusId() string
}

type CommissionClient struct {
	*clients.NuvlaEdgeClient
}

func (cc *CommissionClient) GetStatusId() string {
	neRes := cc.GetNuvlaEdgeResource()
	if neRes.NuvlaBoxStatus == "" {
		err := cc.UpdateResourceSelect([]string{"nuvlabox-status"})
		if err != nil {
			log.Error("Failed to update nuvlabox status ID", err)
			return ""
		}
	}
	return cc.GetNuvlaEdgeResource().NuvlaBoxStatus
}

type ConfUpdaterClient interface {
	UpdateResourceSelect(selects []string) error
	GetNuvlaEdgeResource() resources.NuvlaEdgeResource
}
