package telemetry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"nuvlaedge-go/common"
	"nuvlaedge-go/telemetry/monitor"
	"nuvlaedge-go/types"
	"nuvlaedge-go/types/metrics"
	"time"
)

type Telemetry struct {
	types.WorkerBase

	localStatus metrics.NuvlaEdgeStatus
	lastStatus  metrics.NuvlaEdgeStatus

	//jobChan     chan string
	metricsChan chan metrics.Metric
	period      int

	nuvla types.TelemetryClientInterface

	monitors map[string]monitor.NuvlaEdgeMonitor

	jobChan chan string // Sends a job ID if any to job processor
}

func NewTelemetry(
	period int,
	neClient types.TelemetryClientInterface,
	dockerClient types.DockerMetricsClient,
	commissionerChan chan types.CommissionData,
	jobChan chan string) *Telemetry {

	t := &Telemetry{
		nuvla:       neClient,
		metricsChan: make(chan metrics.Metric, 10), // Buffer size 10 to allow all different metric types to be sent without blocking
		period:      period,
		jobChan:     jobChan,
	}

	t.monitors = map[string]monitor.NuvlaEdgeMonitor{
		"engine":       monitor.NewDockerMonitor(dockerClient, t.period, t.metricsChan, neClient.GetEndpoint(), commissionerChan),
		"system":       monitor.NewSystemMonitor(t.period, t.metricsChan),
		"resources":    monitor.NewResourceMonitor(t.period, t.metricsChan),
		"installation": monitor.NewInstallationMonitor(t.period, dockerClient, t.metricsChan),
	}

	return t
}

func (t *Telemetry) StartMonitors(ctx context.Context) error {
	for k, m := range t.monitors {
		log.Infof("Starting Monitor: %s", k)
		go func(mon monitor.NuvlaEdgeMonitor) {
			if err := mon.Run(ctx); err != nil {
				log.Errorf("Error running monitor: %s", err)
			}
		}(m)
	}
	return nil
}

func (t *Telemetry) monitorStatus(ctx context.Context) {
	for k, m := range t.monitors {
		log.Infof("Monitor %s status: %t", k, m.Running())
		if !m.Running() {
			log.Warnf("Monitor %s is not running, restarting...", k)
			go func(mon monitor.NuvlaEdgeMonitor) {
				if err := mon.Run(ctx); err != nil {
					log.Errorf("Error restarting monitor: %s", err)
				}
			}(m)
		}
	}
}

func (t *Telemetry) Run(ctx context.Context) error {
	log.Info("Starting telemetry...")
	updateTimer := time.NewTicker(time.Duration(t.period) * time.Second)
	defer updateTimer.Stop()

	statusTimer := time.NewTicker(10 * time.Second)
	defer statusTimer.Stop()

	ctxMonitors, cancel := context.WithCancel(ctx)
	defer cancel()
	if err := t.StartMonitors(ctxMonitors); err != nil {
		log.Errorf("Error starting monitors: %s", err)
		return err
	}

	for {
		select {
		case <-ctx.Done():
			log.Info("Telemetry stopped, stopping monitors...")
			cancel()
			return ctx.Err()
		case <-statusTimer.C:
			t.monitorStatus(ctxMonitors)
		case <-updateTimer.C:
			if err := t.sendTelemetry(); err != nil {
				// Report error to status handler
				log.Errorf("Error sending telemetry: %s", err)
			}

		case m := <-t.metricsChan:
			// Process metrics
			if err := m.WriteToStatus(&t.localStatus); err != nil {
				log.Errorf("Error writing metric to status: %s", err)
			}
		}
	}
}

func (t *Telemetry) sendTelemetry() error {
	if t.nuvla == nil {
		return errors.New("telemetry client not initialized, cannot send telemetry")
	}

	// Get data diff from telemetries
	data, sel := common.GetStructDiff(t.lastStatus, t.localStatus)
	if (data == nil && sel == nil) || (len(data) == 0 && len(sel) == 0) {
		return nil
	}

	if log.GetLevel() == log.DebugLevel && (data != nil || len(data) == 0) {
		d, _ := json.MarshalIndent(data, "", "  ")
		log.Infof("Telemetry data to send: %s \n", string(d))
	}

	// Send telemetry to client
	log.Info("Sending telemetry...")
	res, err := t.nuvla.Telemetry(data, sel)
	defer func() {
		if res != nil {
			if err := res.Body.Close(); err != nil {
				log.Errorf("Error closing telemetry response body: %s", err)
			}
		}
	}()
	if err != nil {
		return err
	}

	if res.StatusCode != 200 && res.StatusCode != 201 {
		b, _ := io.ReadAll(res.Body)
		var m map[string]interface{}
		if err := json.Unmarshal(b, &m); err == nil {
			log.Errorf("telemetry failed with message: %s--%s", res.Status, m["message"])
		}

		return fmt.Errorf("telemetry failed with status code: %d", res.StatusCode)
	}

	// Update last status
	t.lastStatus = t.localStatus

	// Process jobs...

	return nil
}

func (t *Telemetry) Stop() {
	log.Info("Stopping telemetry...")
	for k, m := range t.monitors {
		if err := m.Close(); err != nil {
			log.Errorf("Error closing monitor %s: %s", k, err)
		}
	}
}

func (t *Telemetry) GetType() types.WorkerType {
	return types.Telemetry
}

var _ types.Worker = &Telemetry{}
