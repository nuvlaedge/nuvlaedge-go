package deployments

import (
	"context"
	nuvlaApi "github.com/nuvla/api-client-go"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/orchestrator"
	"nuvlaedge-go/types/jobs"
	"nuvlaedge-go/types/worker"
)

type DeploymentHandler struct {
	worker.TimedWorker
	coe    orchestrator.Orchestrator
	client *nuvlaApi.NuvlaClient

	runningDeployments map[string]Deployment
	deploymentJobChan  chan jobs.Job
}

func (d *DeploymentHandler) Init(opts *worker.WorkerOpts, conf *worker.WorkerConfig) error {
	d.TimedWorker = worker.NewTimedWorker(conf.TelemetryPeriod, worker.Deployments)
	return nil
}

func (d *DeploymentHandler) Start(ctx context.Context) error {
	//TODO implement me
	log.Info("Starting Deployment Handler")
	go d.Run(ctx)
	return nil
}

func (d *DeploymentHandler) Reconfigure(conf *worker.WorkerConfig) error {
	//TODO implement me
	return nil
}

func (d *DeploymentHandler) Run(ctx context.Context) error {
	log.Info("Running Deployment Handler")
	for {
		select {
		case <-ctx.Done():
			log.Info("Stopping Deployment Handler")
			if err := d.Stop(ctx); err != nil {
				return err
			}
			return ctx.Err()

		case <-d.BaseTicker.C:
			log.Info("Scan deployments...")
		case <-d.deploymentJobChan:
			log.Info("Received deployment job")

		case conf := <-d.ConfChan:
			log.Info("Received configuration in handler: ", conf)
			if err := d.Reconfigure(conf); err != nil {
				//TODO log error
			}
		}
	}
}

func (d *DeploymentHandler) Stop(ctx context.Context) error {
	//TODO implement me
	return nil
}

func (d *DeploymentHandler) scanDeployments() {

}
