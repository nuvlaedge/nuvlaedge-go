package workers

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

type MockHeartBeatClient struct {
	hbCounter int
	hbError   error
}

func (m *MockHeartBeatClient) Heartbeat(ctx context.Context) (*http.Response, error) {

	return nil, m.hbError
}

func TestHeartbeat_sendHeartbeat(t *testing.T) {
	h := &Heartbeat{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockClient := MockHeartBeatClient{}
	mockClient.hbError = errors.New("error")

	h.client = &mockClient

	err := h.sendHeartbeat(ctx)
	assert.NotNil(t, err)
}
