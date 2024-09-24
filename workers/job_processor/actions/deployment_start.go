package actions

import (
	"github.com/nuvla/api-client-go/clients/resources"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/workers/job_processor/executors"
)

type DeploymentStart struct {
	DeploymentBase
}

func (d *DeploymentStart) ExecuteAction() error {
	defer CloseDeploymentClientWithLog(d.client)
	defer d.executor.Close()

	if err := d.client.SetState(resources.StateStarting); err != nil {
		log.Warnf("Error setting deployment state to starting: %s", err)
	}

	if err := d.executor.StartDeployment(); err != nil {
		if stateErr := d.client.SetState(resources.StateError); stateErr != nil {
			log.Warnf("Error setting deployment state to error: %s", stateErr)
		}
		return err
	}

	// Creates nuvla output params if they don't exist or updates them
	d.CreateUserOutputParams()

	if err := d.client.SetState(resources.StateStarted); err != nil {
		log.Warnf("Error setting deployment state to started: %s", err)
	}
	return nil
}

func (d *DeploymentStart) GetExecutorName() executors.ExecutorName {
	return d.executor.GetName()
}
