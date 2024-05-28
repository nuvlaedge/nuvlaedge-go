package actions

import (
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/nuvlaedge/jobEngine/executors"
)

type DeploymentState struct {
	DeploymentBase
}

func (d *DeploymentState) ExecuteAction() error {
	log.Infof("Deployment state action for deployment %s", d.deploymentId)
	s, err := d.executor.GetServices()
	if err != nil {
		log.Infof("Error getting services for deployment %s: %s", d.deploymentId, err)
		return err
	}
	log.Infof("Deployment %s services: %v", d.deploymentId, s)

	err = d.executor.StateDeployment()
	if err != nil {
		log.Infof("Error getting deployment state for deployment")
		return err
	}
	return nil
}

func (d *DeploymentState) GetExecutorName() executors.ExecutorName {
	return d.executor.GetName()
}