package types

import (
	nuvla "github.com/nuvla/api-client-go"
	"github.com/nuvla/api-client-go/clients"
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
