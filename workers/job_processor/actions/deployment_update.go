package actions

import (
	"github.com/nuvla/api-client-go/clients/resources"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/workers/job_processor/executors"
)

type DeploymentUpdate struct {
	DeploymentBase
}

func (d *DeploymentUpdate) assertExecutor() error {
	return nil
}

func (d *DeploymentUpdate) ExecuteAction() error {

	if err := d.executor.UpdateDeployment(); err != nil {
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

func (d *DeploymentUpdate) GetExecutorName() executors.ExecutorName {
	return d.executor.GetName()
}
