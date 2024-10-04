package testutils

import (
	"context"
	"net/http"
)

type MockTelemetryClient struct {
	TelemetryCnt      int
	TelemetryErr      error
	TelemetryResponse *http.Response

	GetEndpointCnt      int
	GetEndpointResponse string
}

func (mtc *MockTelemetryClient) Telemetry(ctx context.Context, data interface{}, Select []string) (*http.Response, error) {
	mtc.TelemetryCnt++
	return mtc.TelemetryResponse, mtc.TelemetryErr
}

func (mtc *MockTelemetryClient) GetEndpoint() string {
	mtc.GetEndpointCnt++
	return mtc.GetEndpointResponse
}
