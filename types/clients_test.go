package types

import (
	nuvla "github.com/nuvla/api-client-go"
	"github.com/nuvla/api-client-go/clients"
	nuvlaTypes "github.com/nuvla/api-client-go/types"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func init() {
	log.SetLevel(log.PanicLevel)
}

func Test_TelemetryClient_GetEndpoint(t *testing.T) {
	// Test code here
	tc := TelemetryClient{
		NuvlaEdgeClient: clients.NewNuvlaEdgeClient("nuvlaedge-uuid", nil, nuvla.WithEndpoint("http://localhost:8080")),
	}

	assert.Equal(t, "http://localhost:8080", tc.GetEndpoint(), "Endpoints should be equal")
}

func Test_CommissionClient_GetStatusID(t *testing.T) {
	// Test code here
	cc := CommissionClient{
		NuvlaEdgeClient: clients.NewNuvlaEdgeClient("nuvlaedge-uuid", nil),
	}

	assert.Equal(t, "", cc.GetStatusId(), "Status IDs should be equal")

	cc.NuvlaEdgeStatusId = &nuvlaTypes.NuvlaID{
		Id:           "mock/1234",
		ResourceType: "mock",
		Uuid:         "1234",
	}
	assert.Equal(t, "mock/1234", cc.GetStatusId(), "Status IDs should be equal")

}
