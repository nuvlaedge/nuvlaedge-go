package nuvlaedge

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/types/worker"
	"nuvlaedge-go/workers"
	"nuvlaedge-go/workers/job_processor"
	"nuvlaedge-go/workers/telemetry"
)

// A worker is a module of NuvlaEdge that executes a periodic task (or periodically triggered task).

// Workers:

// Periodic:
// - Telemetry
//.     - 60s
// - Heartbeat
//.     - 20s
// - CleanUp
//      - 60s
// - Commission
//      - 60s
//      - (Future) VPN Handler

// Triggered:
// - ConfUpdate
//.  	- Telemetry
//.     - Heartbeat
// - JobProcessor
//.  	- Telemetry
//.     - Heartbeat
// - Deployment
//.  	- JobProcessor

func generateWorkers() Workers {
	return Workers{
		// Timed
		worker.Telemetry:       &telemetry.Telemetry{},
		worker.Heartbeat:       &workers.Heartbeat{},
		worker.ResourceCleaner: &workers.DockerCleaner{},
		worker.Commissioner:    &workers.Commissioner{},

		// Triggered
		worker.JobProcessor: &job_processor.JobProcessor{},
		//worker.Deployments:  &deployments.DeploymentProcessor{},
		worker.ConfUpdater: &workers.ConfUpdater{},
	}
}

func WorkerGenerator(opts *worker.WorkerOpts, conf *worker.WorkerConfig) (Workers, error) {
	workerMap := generateWorkers()

	var errList []error
	var confChannels []chan *worker.WorkerConfig
	// A bit of overhead since ATM no worker returns an error on Init, but the structure is in place and might be useful
	for n, w := range workerMap {
		if n == worker.ConfUpdater {
			// We need to initialise the conf updater last to provide it with the conf channels
			continue
		}
		log.Infof("Initializing worker %s", n)
		if err := w.Init(opts, conf); err != nil {
			log.Errorf("Error initializing worker %s: %s", w.GetName(), err)
			errList = append(errList, err)
		}
		confChannels = append(confChannels, w.GetConfChannel())
	}
	log.Infof("Initializing worker %s", worker.ConfUpdater)
	// Init the conf updater last
	opts.ConfigChannels = confChannels
	if err := workerMap[worker.ConfUpdater].Init(opts, conf); err != nil {
		log.Errorf("Error initializing worker %s: %s", worker.ConfUpdater, err)
		errList = append(errList, err)
	}
	return workerMap, errors.Join(errList...)
}
