package actions

import (
	"context"
	"fmt"
	"github.com/nuvla/api-client-go/clients/resources"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/common/constants"
	"nuvlaedge-go/workers/job_processor/executors"
	"time"
)

type DeploymentStop struct {
	DeploymentBase
}

func (d *DeploymentStop) ExecuteAction(ctx context.Context) error {
	defer CloseDeploymentClientWithLog(d.client)
	defer d.executor.Close()

	ctxTimed, cancel := context.WithTimeout(ctx, constants.DefaultJobTimeout*time.Second)
	defer cancel()

	if err := d.client.SetState(ctxTimed, resources.StateStopping); err != nil {
		log.Warnf("Error setting deployment state to stopping: %s", err)
	}

	if err := d.executor.StopDeployment(ctxTimed); err != nil {
		if stateErr := d.client.SetState(ctxTimed, resources.StateError); stateErr != nil {
			log.Warnf("Error setting deployment state to error: %s", stateErr)
		}
		return fmt.Errorf("error stopping deployment: %s", err)
	}
	if err := d.client.SetState(ctxTimed, resources.StateStopped); err != nil {
		log.Warnf("Error setting deployment state to stopped: %s", err)
	}

	return nil
}

func (d *DeploymentStop) GetExecutorName() executors.ExecutorName {
	return d.executor.GetName()
}

func (d *DeploymentStop) GetOutput() string {
	return d.executor.GetOutput()
}
