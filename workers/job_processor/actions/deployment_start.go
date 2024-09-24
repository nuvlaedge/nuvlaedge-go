package actions

import (
	"github.com/nuvla/api-client-go/clients/resources"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/workers/job_processor/executors"
)

type DeploymentStart struct {
	DeploymentBase
}

func (d *DeploymentStart) CreateUserOutputParams() {
	// Fixed parameters for all deployments, hostname and IPs. TODO: IP should be created by Nuvla...
	if err := d.ManageHostNameParam(); err != nil {
		log.Warnf("Error creating hostname parameter: %s", err)
	}

	if err := d.ManageIPsParams(); err != nil {
		log.Warnf("Error creating IPs parameters: %s", err)
	}

	if err := d.ManageDeploymentParameters(); err != nil {
		log.Warnf("Error creating deployment parameters: %s", err)
	}
}

func (d *DeploymentStart) ExecuteAction() error {
	defer CloseDeploymentClientWithLog(d.client)
	defer d.executor.Close()

	if err := d.client.SetState(resources.StateStarting); err != nil {
		log.Warnf("Error setting deployment state to starting: %s", err)
	}

	// Creates nuvla output params if they don't exist or updates them
	d.CreateUserOutputParams()

	if err := d.executor.StartDeployment(); err != nil {
		if stateErr := d.client.SetState(resources.StateError); stateErr != nil {
			log.Warnf("Error setting deployment state to error: %s", stateErr)
		}
		return err
	}
	if err := d.client.SetState(resources.StateStarted); err != nil {
		log.Warnf("Error setting deployment state to started: %s", err)
	}

	return nil
}

func (d *DeploymentStart) GetExecutorName() executors.ExecutorName {
	return d.executor.GetName()
}

func (d *DeploymentStart) GetOutput() string {
	return d.executor.GetOutput()
}
