package actions

import (
	"context"
	"github.com/nuvla/api-client-go/clients/resources"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/workers/job_processor/executors"
	"time"
)

type DeploymentStart struct {
	DeploymentBase
}

func (d *DeploymentStart) ExecuteAction(ctx context.Context) error {
	defer CloseDeploymentClientWithLog(d.client)
	defer d.executor.Close()

	ctxTimed, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	if err := d.client.SetState(ctxTimed, resources.StateStarting); err != nil {
		log.Warnf("Error setting deployment state to starting: %s", err)
	}

	if err := d.executor.StartDeployment(ctxTimed); err != nil {
		if stateErr := d.client.SetState(ctxTimed, resources.StateError); stateErr != nil {
			log.Warnf("Error setting deployment state to error: %s", stateErr)
		}
		return err
	}

	// Creates nuvla output params if they don't exist or updates them
	d.CreateUserOutputParams(ctxTimed)

	if err := d.client.SetState(ctxTimed, resources.StateStarted); err != nil {
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
