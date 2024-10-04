package actions

import (
	"context"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/workers/job_processor/executors"
	"time"
)

type DeploymentState struct {
	DeploymentBase
}

func (d *DeploymentState) ExecuteAction(ctx context.Context) error {
	defer CloseDeploymentClientWithLog(d.client)
	defer d.executor.Close()

	ctxTimed, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	log.Infof("Deployment state action for deployment %s", d.deploymentId)
	log.Debugf("Deployment executor: %s", d.executor.GetName())

	s, err := d.executor.GetServices(ctxTimed)
	if err != nil {
		log.Infof("Error getting services for deployment %s: %s", d.deploymentId, err)
		return err
	}

	d.CreateUserOutputParams(ctxTimed)

	log.Infof("Deployment %s services: %v", d.deploymentId, s)
	err = d.manageServiceParameters(ctxTimed, s)
	if err != nil {
		log.Warnf("Error managing service parameters for deployment %s: %s", d.deploymentId, err)
	}

	err = d.executor.StateDeployment(ctxTimed)
	if err != nil {
		log.Infof("Error getting deployment state for deployment")
		return err
	}
	return nil
}

func (d *DeploymentState) GetExecutorName() executors.ExecutorName {
	return d.executor.GetName()
}

func (d *DeploymentState) GetOutput() string {
	return d.executor.GetOutput()
}
