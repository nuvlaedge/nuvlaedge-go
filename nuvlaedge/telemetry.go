package nuvlaedge

import (
	"nuvlaedge-go/nuvlaedge/coe"
	"nuvlaedge-go/nuvlaedge/common/resources"
	"nuvlaedge-go/nuvlaedge/monitoring"
)

type Telemetry struct {
	sentNuvlaEdgeStatus *resources.NuvlaEdgeStatus
	metricsMonitor      *monitoring.MetricsMonitor
}

func NewTelemetry(coeClient coe.Coe, updatePeriod int) *Telemetry {
	metricsMonitor := monitoring.NewMetricsMonitor(coeClient, updatePeriod)

	return &Telemetry{
		metricsMonitor: metricsMonitor,
	}
}

func (t *Telemetry) Start() error {
	t.metricsMonitor.Run()
	return nil
}

func (t *Telemetry) StatusDiff(status *resources.NuvlaEdgeStatus) ([]string, []string) {
	return nil, nil
}

func (t *Telemetry) GetMapFromFields(fields []string) map[string]interface{} {
	return nil
}

// GetStatusToSend returns the status to send to Nuvla in the form of a map and a list of fields to delete
func (t *Telemetry) GetStatusToSend() (map[string]interface{}, []string) {
	newStatus := &resources.NuvlaEdgeStatus{}
	err := t.metricsMonitor.GetNewFullStatus(newStatus)
	if err != nil {
		log.Warn("Error getting new full status from MetricsMonitor")
		return nil, nil
	}

	return nil, nil
}
