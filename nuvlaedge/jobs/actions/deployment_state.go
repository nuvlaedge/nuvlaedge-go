package actions

import (
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/nuvlaedge/jobs/executors"
)

type DeploymentState struct {
	DeploymentBase
}

func (d *DeploymentState) ExecuteAction() error {
	log.Infof("Dummy deployment state")
	return nil
	//log.Infof("Deployment state action for deployment %s", d.deploymentId)
	//s, err := d.executor.GetServices()
	//temp, err := json.MarshalIndent(s, "", "  ")
	//log.Infof("Deployment %s services: \n%s", d.deploymentId, temp)
	//if err != nil {
	//	log.Infof("Error getting services for deployment %s: %s", d.deploymentId, err)
	//	return err
	//}
	//log.Infof("Deployment %s services: %v", d.deploymentId, s)
	//err = d.manageServiceParameters(s)
	//if err != nil {
	//	log.Warnf("Error managing service parameters for deployment %s: %s", d.deploymentId, err)
	//}
	//
	//err = d.executor.StateDeployment()
	//if err != nil {
	//	log.Infof("Error getting deployment state for deployment")
	//	return err
	//}
	//return nil
}

func (d *DeploymentState) GetExecutorName() executors.ExecutorName {
	return d.executor.GetName()
}
