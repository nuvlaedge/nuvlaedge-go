package telemetry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/wI2L/jsondiff"
	"io"
	"nuvlaedge-go/common"
	"nuvlaedge-go/common/constants"
	"nuvlaedge-go/common/version"
	"nuvlaedge-go/types"
	"nuvlaedge-go/types/metrics"
	"nuvlaedge-go/types/worker"
	"nuvlaedge-go/workers/telemetry/monitor"
	"time"
)

type Telemetry struct {
	worker.TimedWorker

	localStatus metrics.NuvlaEdgeStatus
	lastStatus  metrics.NuvlaEdgeStatus

	//jobChan     chan string
	metricsChan chan metrics.Metric

	nuvla types.TelemetryClientInterface

	monitors map[string]monitor.NuvlaEdgeMonitor

	jobChan chan string // Sends a job ID if any to job processor
}

func (t *Telemetry) Init(opts *worker.WorkerOpts, conf *worker.WorkerConfig) error {
	// Configure telemetry
	t.TimedWorker = worker.NewTimedWorker(conf.TelemetryPeriod, worker.Telemetry)
	t.nuvla = &types.TelemetryClient{NuvlaEdgeClient: opts.NuvlaClient}

	// Init telemetry
	t.metricsChan = make(chan metrics.Metric, 10)
	t.jobChan = opts.JobCh

	t.monitors = map[string]monitor.NuvlaEdgeMonitor{
		"engine":       monitor.NewDockerMonitor(opts.DockerClient, t.GetPeriod(), t.metricsChan, t.nuvla.GetEndpoint(), opts.CommissionCh),
		"system":       monitor.NewSystemMonitor(t.GetPeriod(), t.metricsChan),
		"resources":    monitor.NewResourceMonitor(t.GetPeriod(), t.metricsChan),
		"installation": monitor.NewInstallationMonitor(t.GetPeriod(), opts.DockerClient, t.metricsChan),
	}
	return nil
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
		log.Debugf("Monitor %s status: %t", k, m.Running())
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

func (t *Telemetry) Start(ctx context.Context) error {
	log.Info("Starting telemetry...")

	// Part of the telemetry that will be fixed and only defined once
	t.setInitialStatus()

	if err := t.StartMonitors(ctx); err != nil {
		log.Errorf("Error starting monitors: %s", err)
		return err
	}

	go func() {
		err := t.Run(ctx)
		if err != nil {
			log.Errorf("Error running Commissioner: %s", err)
		}
	}()

	return nil
}

func (t *Telemetry) Run(ctx context.Context) error {
	log.Debug("Running telemetry...")

	statusTimer := time.NewTicker(60 * time.Second)
	defer statusTimer.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("Telemetry stopped, stopping monitors...")
			return ctx.Err()
		case <-statusTimer.C:
			t.monitorStatus(ctx)

		case <-t.BaseTicker.C:
			log.Info("Try sending telemetry...")
			patch, data, attrsToDelete := t.getTelemetryDiff()
			var patchErr error
			if patch != nil {
				if patchErr = t.sendTelemetry(patch, attrsToDelete); patchErr != nil {
					// Report error to status handler
					log.Errorf("Error sending telemetry patch: %s", patchErr)
				}
			}
			if patch == nil || patchErr != nil {
				if err := t.sendTelemetry(data, attrsToDelete); err != nil {
					// Report error to status handler
					log.Errorf("Error sending telemetry: %s", err)
				}
			}

		case m := <-t.metricsChan:
			// Process metrics
			if err := m.WriteToStatus(&t.localStatus); err != nil {
				log.Errorf("Error writing metric to status: %s", err)
			}

		case conf := <-t.ConfChan:
			log.Debug("Received configuration in telemetry: ", conf)
			if err := t.Reconfigure(conf); err != nil {
				log.Errorf("Error reconfiguring telemetry: %s", err)
			}
		}
	}
}

func (t *Telemetry) setInitialStatus() {
	t.localStatus.NuvlaEdgeEngineVersion = version.GetVersion() + "-go"
	t.localStatus.Status = "OPERATIONAL"
	t.localStatus.Version = 2
}

func (t *Telemetry) getTelemetryDiff() (jsondiff.Patch, map[string]interface{}, []string) {
	// Update current time
	t.localStatus.CurrentTime = time.Now().Format(constants.DatetimeFormat)

	data, attrsToDelete := common.GetStructDiff(t.lastStatus, t.localStatus)

	patch, err := jsondiff.Compare(t.lastStatus, t.localStatus, jsondiff.Factorize())
	if err != nil {
		log.Errorf("Error creating telemetry patch: %v", err)
		return nil, data, attrsToDelete
	}

	return patch, data, attrsToDelete
}

func (t *Telemetry) sendTelemetry(data interface{}, attrsToDelete []string) error {
	if t.nuvla == nil {
		return errors.New("telemetry client not initialized, cannot send telemetry")
	}

	if data == nil && (attrsToDelete == nil || len(attrsToDelete) == 0) {
		return nil
	}

	if log.GetLevel() == log.DebugLevel && data != nil {
		d, _ := json.MarshalIndent(data, "", "  ")
		log.Infof("Telemetry data to send: %s \n", string(d))
	}

	// Send telemetry to client
	log.Info("Sending telemetry...")
	res, err := t.nuvla.Telemetry(data, attrsToDelete)
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
	if err := common.ProcessResponse(res, t.jobChan, nil); err != nil {
		log.Errorf("Error processing telemetry response: %s", err)
	}

	return nil
}

func (t *Telemetry) Reconfigure(conf *worker.WorkerConfig) error {
	if conf.TelemetryPeriod != t.GetPeriod() {
		t.SetPeriod(conf.TelemetryPeriod)
	}
	return nil
}

func (t *Telemetry) Stop(_ context.Context) error {
	log.Info("Stopping telemetry...")
	var errList []error
	for k, m := range t.monitors {
		if err := m.Close(); err != nil {
			log.Errorf("Error closing monitor %s: %s", k, err)
			errList = append(errList, err)
		}
	}
	return errors.Join(errList...)
}

var _ worker.Worker = &Telemetry{}
