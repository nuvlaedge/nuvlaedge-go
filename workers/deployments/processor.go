package deployments

import (
	"context"
	"errors"
	"fmt"
	nuvla "github.com/nuvla/api-client-go"
	"github.com/nuvla/api-client-go/clients"
	"github.com/nuvla/api-client-go/clients/resources"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/orchestrator"
	"nuvlaedge-go/types/jobs"
	"nuvlaedge-go/types/worker"
	"strings"
	"time"
)

type DeploymentProcessor struct {
	worker.TimedWorker

	orchestrators map[string]jobs.Deployer

	//ComposeOrchestrator orchestrator.Orchestrator
	//SwarmOrchestrator   orchestrator.Orchestrator
	//
	//HelmOrchestrator       orchestrator.Orchestrator
	//KubernetesOrchestrator orchestrator.Orchestrator

	nuvlaClient *nuvla.NuvlaClient // Might not be needed...

	currentJobs *jobs.JobRegistry // Shared with Job Processor so that it can remove jobs when they are done

	// Keeps track of the currently running deployments in the different available orchestrators
	currentDeployments interface{}
	deploymentChannel  chan jobs.Job
}

/** Worker interface implementation as a timed worker */

func (d *DeploymentProcessor) Init(opts *worker.WorkerOpts, conf *worker.WorkerConfig) error {
	log.Info("Initialising Deployment Processor...")
	d.TimedWorker = worker.NewTimedWorker(conf.TelemetryPeriod, worker.Deployments)
	d.deploymentChannel = opts.DeploymentCh
	d.currentJobs = opts.Jobs

	var errs []error
	// Initialise orchestrators:
	d.orchestrators = make(map[string]jobs.Deployer)
	log.Infof("Creating orchestrators...")
	compose, err := orchestrator.NewComposeOrchestrator(opts.DockerClient)
	if err != nil {
		log.Warn("Error creating Compose orchestrator: ", err)
		errs = append(errs, err)
	}
	d.orchestrators["compose"] = NewComposeDeployer(compose)
	log.Infof("Compose created")

	//swarm, err := orchestrator.NewSwarmOrchestrator(opts.DockerClient)
	//if err != nil {
	//	log.Warn("Error creating Swarm orchestrator: ", err)
	//	errs = append(errs, err)
	//}
	//d.orchestrators["swarm"] = swarm
	//log.Infof("Swarm created")

	return errors.Join(errs...)
}

func (d *DeploymentProcessor) Start(ctx context.Context) error {
	log.Info("Starting Deployment Processor...")
	go d.Run(ctx)
	return nil
}

func (d *DeploymentProcessor) Reconfigure(conf *worker.WorkerConfig) error {
	if d.GetPeriod() != conf.TelemetryPeriod {
		d.SetPeriod(conf.TelemetryPeriod)
	}
	return nil
}

func (d *DeploymentProcessor) Run(ctx context.Context) error {
	for {
		log.Info("Deployment Processor running...")
		select {
		case <-ctx.Done():
			log.Warn("Deployment Processor received exit signal")
			return nil
		case <-d.BaseTicker.C:
			// Here goes the deployment Monitoring
			//TODO implement me
			log.Debug("Deployment Processor running...")

		// Send new telemetry
		case job := <-d.deploymentChannel:
			log.Info("Received deployment job: ", job.GetId())
			jBase, ok := job.(*jobs.JobBase)
			if !ok {
				log.Error("Received job is not of type JobBase")
				continue
			}
			d.processDeployment(ctx, jBase)

		case conf := <-d.ConfChan:
			log.Info("Received configuration in Deployment Processor: ", conf)
			if err := d.Reconfigure(conf); err != nil {
				log.Error("Failed to reconfigure deployment processor: ", err)
			}
		}
	}
}

func (d *DeploymentProcessor) Stop(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (d *DeploymentProcessor) processDeployment(ctx context.Context, job *jobs.JobBase) {
	if !d.currentJobs.Exists(job.GetId()) {
		log.Infof("Job should have already been registered: %s", job.GetId())
		return
	}
	defer d.currentJobs.Remove(job.GetId())

	deploymentCli, err := getDeploymentClientFromJob(job)
	if err != nil {
		log.Errorf("Error getting deployment client: %s", err)
		return
	}

	// Find required orchestrator
	o, err := d.getOrchestrator(deploymentCli.GetResource().Module)
	if err != nil {
		log.Errorf("Error getting orchestrator: %s", err)
		job.Client.SetFailedState(err.Error())
		return
	}

	job.Client.SetInitialState()

	depCtx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	opts := &jobs.DeploymentOpts{
		DeploymentClient: deploymentCli,
		Orchestrator:     o,
		Context:          depCtx,
		Job:              job,
	}
	err = d.runDeploymentAction(opts)

	if err != nil {
		log.Errorf("Error processing deployment: %s", err)
		job.Client.SetFailedState(err.Error())
		return
	}
	job.Client.SetSuccessState()
}

func (d *DeploymentProcessor) runDeploymentAction(opts *jobs.DeploymentOpts) error {
	// Run deployment action
	var err error
	switch opts.Job.GetJobType() {
	case "start_deployment":
		err = d.StartDeployment(opts)
	case "stop_deployment":
		err = d.StopDeployment(opts)
	case "update_deployment":
		err = d.UpdateDeployment(opts)
	case "fetch_deployment_log":
		err = d.GetDeploymentLogs(opts)
	case "deployment_state":
		err = d.GetDeploymentState(opts)
	default:
		err = fmt.Errorf("job type %s not supported", opts.Job.GetJobType())
	}
	return err
}

func (d *DeploymentProcessor) getOrchestrator(module *resources.ModuleResource) (jobs.Deployer, error) {
	if module == nil {
		return nil, errors.New("module is nil")
	}
	compatibility := module.Compatibility
	subType := module.SubType

	switch subType {
	case "application":
		switch compatibility {
		case "docker-compose":
			orch, ok := d.orchestrators["compose"]
			if !ok {
				return nil, errors.New("Compose orchestrator not available")
			}
			return orch, nil
		case "swarm":
			orch, ok := d.orchestrators["swarm"]
			if !ok {
				return nil, errors.New("Swarm orchestrator not available")
			}
			return orch, nil
		default:
			return nil, fmt.Errorf("compatibility %s not supported", compatibility)
		}
	case "application_kubernetes":
		return nil, errors.New("Kubernetes deployments are not supported yet")
	default:
		return nil, fmt.Errorf("subType %s not supported", subType)
	}
}

/**
Deployment Actions:
- Deploy
- Update
- Remove
- Stop
- Retrieve logs
*/

func (d *DeploymentProcessor) StartDeployment(opts *jobs.DeploymentOpts) error {
	log.Info("Starting deployment...")
	if err := opts.DeploymentClient.SetState(resources.StateStarting); err != nil {
		log.Warnf("Error setting deployment state to started: %s", err)
	}

	// Create user output params
	dep := opts.DeploymentClient.GetResource()

	if err := opts.Orchestrator.StartDeployment(opts.Context, dep); err != nil {
		if stateErr := opts.DeploymentClient.SetState(resources.StateError); stateErr != nil {
			log.Warnf("Error setting deployment state to error: %s", stateErr)
		}
		return err
	}

	if err := opts.DeploymentClient.SetState(resources.StateStarted); err != nil {
		log.Warnf("Error setting deployment state to started: %s", err)
	}

	return nil
}

func (d *DeploymentProcessor) StopDeployment(opts *jobs.DeploymentOpts) error {
	log.Info("Stopping deployment...")
	if err := opts.DeploymentClient.SetState(resources.StateStopping); err != nil {
		log.Warnf("Error setting deployment state to stopping: %s", err)
	}

	if err := opts.Orchestrator.StopDeployment(opts.Context, opts.DeploymentClient.GetResource().Id); err != nil {
		if stateErr := opts.DeploymentClient.SetState(resources.StateError); stateErr != nil {
			log.Warnf("Error setting deployment state to error: %s", stateErr)
		}
		return err
	}

	if err := opts.DeploymentClient.SetState(resources.StateStopped); err != nil {
		log.Warnf("Error setting deployment state to stopped: %s", err)
	}

	return nil
}

func (d *DeploymentProcessor) UpdateDeployment(opts *jobs.DeploymentOpts) error {
	log.Info("Updating deployment...")

	return d.StartDeployment(opts)
}

func (d *DeploymentProcessor) DeleteDeployment(opts *jobs.DeploymentOpts) error {
	log.Info("Deleting deployment...")
	return nil
}

func (d *DeploymentProcessor) GetDeploymentLogs(opts *jobs.DeploymentOpts) error {
	resourceLogId := opts.Job.Resource.TargetResource.Href
	log.Infof("Resource log id: %s", resourceLogId)

	log.Info("Getting deployment logs...")
	logs, err := opts.Orchestrator.GetDeploymentLogs(opts.Context, opts.DeploymentClient.GetResource().Id)
	if err != nil {
		log.Errorf("Error getting deployment logs: %s", err)
		return err
	}
	log.Infof("Deployment logs: %s", logs)

	//opts.Job.Client.Edit(resourceLogId, logs)
	return err
}

func (d *DeploymentProcessor) GetDeploymentState(opts *jobs.DeploymentOpts) error {
	log.Info("Getting deployment state...")
	return nil
}

func getDeploymentClientFromJob(job *jobs.JobBase) (*clients.NuvlaDeploymentClient, error) {
	var dId string
	if strings.HasPrefix(job.Resource.TargetResource.Href, "resource-log/") {
		tId, err := extractDeploymentFromLogResource(job.Resource.TargetResource.Href, job.Client.NuvlaClient)
		if err != nil {
			log.Errorf("Error extracting deployment from log resource: %s", err)
			return nil, err
		}
		dId = tId
	} else {
		dId = job.Resource.TargetResource.Href
	}

	dCli := clients.NewNuvlaDeploymentClient(dId, job.Client.NuvlaClient)
	if err := dCli.UpdateResource(); err != nil {
		log.Errorf("Error updating deployment resource: %s", err)
		return nil, err
	}
	res := dCli.GetResource()
	sessionOpts := dCli.NuvlaClient.SessionOpts
	nuvlaClient := nuvla.NewNuvlaClient(nil, &sessionOpts)
	if err := nuvlaClient.LoginApiKeys(res.ApiCredentials.ApiKey, res.ApiCredentials.ApiSecret); err != nil {
		return nil, err
	}

	dCli.NuvlaClient = nuvlaClient

	return dCli, nil
}

func extractDeploymentFromLogResource(resourceLogId string, client *nuvla.NuvlaClient) (string, error) {
	return "", nil
}
