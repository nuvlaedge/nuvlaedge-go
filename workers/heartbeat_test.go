package workers

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

type MockHeartBeatClient struct {
	hbCounter int
	hbError   error
}

func (m *MockHeartBeatClient) Heartbeat() (*http.Response, error) {

	return nil, m.hbError
}

func TestHeartbeat_sendHeartbeat(t *testing.T) {
	h := &Heartbeat{}
	mockClient := MockHeartBeatClient{}
	mockClient.hbError = errors.New("error")

	h.client = &mockClient

	err := h.sendHeartbeat()
	assert.NotNil(t, err)
}
