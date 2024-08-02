package testutils

import "nuvlaedge-go/types/metrics"

type MetricMock struct {
	IncCnt int
	IncErr error
}

func (mm *MetricMock) WriteToStatus(status *metrics.NuvlaEdgeStatus) error {
	mm.IncCnt++
	return mm.IncErr
}
