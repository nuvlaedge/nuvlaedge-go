package nuvlaedge

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/nuvlaedge/common/resources"
	"nuvlaedge-go/nuvlaedge/monitoring"
	"nuvlaedge-go/nuvlaedge/orchestrator"
)

type Telemetry struct {
	sentNuvlaEdgeStatus *resources.NuvlaEdgeStatus
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

func (t *Telemetry) StatusDiff(status *resources.NuvlaEdgeStatus) ([]string, []string) {
	return nil, nil
}

func (t *Telemetry) GetMapFromFields(fields []string) map[string]interface{} {
	return nil
}

// GetStatusToSend returns the status to send to Nuvla in the form of a map and a list of fields to delete
func (t *Telemetry) GetStatusToSend() (map[string]interface{}, error) {
	newStatus := &resources.NuvlaEdgeStatus{}
	err := t.metricsMonitor.GetNewFullStatus(newStatus)
	if err != nil {
		log.Warn("Error getting new full status from MetricsMonitor")
		return nil, err
	}
	b, err := json.Marshal(newStatus)
	if err != nil {
		log.Warn("Error marshaling new status")
		return nil, err
	}

	var ret map[string]interface{}
	err = json.Unmarshal(b, &ret)
	return ret, err
}
