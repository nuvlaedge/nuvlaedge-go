package deployments

import (
	"context"
	nuvlaApi "github.com/nuvla/api-client-go"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/engine"
	"nuvlaedge-go/orchestrator"
	"nuvlaedge-go/types"
)

type DeploymentHandler struct {
	types.WorkerBase
	coe    orchestrator.Orchestrator
	ce     engine.ContainerEngine
	client *nuvlaApi.NuvlaClient

	runningDeployments map[string]Deployment
	deploymentJobChan  chan string
}

func NewDeploymentHandler(period int, engine engine.ContainerEngine, orchestrator orchestrator.Orchestrator, nuvla *nuvlaApi.NuvlaClient) *DeploymentHandler {
	return &DeploymentHandler{
		WorkerBase:         types.NewWorkerBase(period, types.Deployments),
		ce:                 engine,
		coe:                orchestrator,
		client:             nuvla,
		runningDeployments: make(map[string]Deployment),
	}
}

func (d *DeploymentHandler) Run(ctx context.Context) error {
	d.Status = types.RUNNING
	defer d.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case job := <-d.deploymentJobChan:
			log.Info("Received deployment job: ", job)
			go d.processDeployment(job)
		// Handle deployment job
		case <-d.BaseTicker.C:
			log.Info("Deployment worker tick")
			// Report deployment status, at some point, they will go to telemetry via metrics
		}
	}
}

func (d *DeploymentHandler) processDeployment(job string) {
	if _, ok := d.runningDeployments[job]; ok {
		log.Warn("Deployment job already running: ", job)
		return
	}

	log.Info("Processing deployment: ", job)
}

func (d *DeploymentHandler) startDeployment(ctx context.Context) error {
	//d.coe.Start()
	return nil
}

func (d *DeploymentHandler) stopDeployment(ctx context.Context) error {
	return nil
}

func (d *DeploymentHandler) stateDeployment(ctx context.Context) error {
	return nil
}

func (d *DeploymentHandler) logsDeployment(ctx context.Context) error {
	return nil
}
