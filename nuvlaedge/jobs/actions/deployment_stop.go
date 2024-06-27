package actions

import (
	"github.com/nuvla/api-client-go/clients/resources"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/nuvlaedge/jobs/executors"
)

type DeploymentStop struct {
	DeploymentBase
}

func (d *DeploymentStop) ExecuteAction() error {
	defer CloseDeploymentClientWithLog(d.client)

	if err := d.client.SetState(resources.StateStopping); err != nil {
		log.Warnf("Error setting deployment state to stopping: %s", err)
	}

	if err := d.executor.StopDeployment(); err != nil {
		if stateErr := d.client.SetState(resources.StateError); stateErr != nil {
			log.Warnf("Error setting deployment state to error: %s", stateErr)
		}
		return err
	}
	if err := d.client.SetState(resources.StateStopped); err != nil {
		log.Warnf("Error setting deployment state to stopped: %s", err)
	}
	return nil
}

func (d *DeploymentStop) GetExecutorName() executors.ExecutorName {
	return d.executor.GetName()
}
