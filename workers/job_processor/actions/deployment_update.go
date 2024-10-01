package actions

import (
	"context"
	"github.com/nuvla/api-client-go/clients/resources"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/workers/job_processor/executors"
	"time"
)

type DeploymentUpdate struct {
	DeploymentBase
}

func (d *DeploymentUpdate) assertExecutor() error {
	return nil
}

func (d *DeploymentUpdate) ExecuteAction(ctx context.Context) error {
	ctxTimeout, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	if err := d.executor.UpdateDeployment(ctxTimeout); err != nil {
		if stateErr := d.client.SetState(ctxTimeout, resources.StateError); stateErr != nil {
			log.Warnf("Error setting deployment state to error: %s", stateErr)
		}
		return err
	}
	if err := d.client.SetState(ctxTimeout, resources.StateStarted); err != nil {
		log.Warnf("Error setting deployment state to started: %s", err)
	}
	return nil
}

func (d *DeploymentUpdate) GetExecutorName() executors.ExecutorName {
	return d.executor.GetName()
}

func (d *DeploymentUpdate) GetOutput() string {
	return d.executor.GetOutput()
}
