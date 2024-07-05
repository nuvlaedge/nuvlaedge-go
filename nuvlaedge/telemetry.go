package nuvlaedge

import (
	"nuvlaedge-go/nuvlaedge/common"
	"nuvlaedge-go/nuvlaedge/common/resources"
	"nuvlaedge-go/nuvlaedge/monitoring"
	"nuvlaedge-go/nuvlaedge/orchestrator"
)

type Telemetry struct {
	sentNuvlaEdgeStatus resources.NuvlaEdgeStatus
	metricsMonitor      *monitoring.MetricsMonitor
}

func NewTelemetry(coeClient orchestrator.Coe, updatePeriod int) *Telemetry {
	metricsMonitor := monitoring.NewMetricsMonitor(coeClient, updatePeriod)

	return &Telemetry{
		metricsMonitor: metricsMonitor,
	}
}

func (t *Telemetry) Start() error {
	go t.metricsMonitor.Run()
	return nil
}

func (t *Telemetry) GetMapFromFields(fields []string) map[string]interface{} {
	return nil
}

// GetStatusToSend returns the status to send to Nuvla in the form of a map and a list of fields to delete
func (t *Telemetry) GetStatusToSend() (map[string]interface{}, []string) {
	newStatus := t.metricsMonitor.GetNewFullStatus()

	// Get the diff between the new status and the last sent status
	diff, deletedFields := common.GetStructDiff(t.sentNuvlaEdgeStatus, newStatus)

	// Update the last sent status
	t.sentNuvlaEdgeStatus = newStatus

	return diff, deletedFields
}
