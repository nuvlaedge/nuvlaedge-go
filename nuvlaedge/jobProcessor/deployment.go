package jobProcessor

import (
	"github.com/nuvla/api-client-go/clients"
	log "github.com/sirupsen/logrus"
)

// GetRequestedConnector returns the requested connector for the deployment. If the connector is not supported,
// it returns NotSupportedDeploymentConnector error and the requested connector.
func GetRequestedConnector() (string, error) {
	return "", nil
}

type DeploymentBase struct {
	DeploymentClient   *clients.NuvlaDeploymentClient
	DeploymentResource *clients.DeploymentResource
	JobResource        *clients.JobResource
	deployer           Deployer
}

func (d *DeploymentBase) Init(opts ...ActionOptsFunc) error {
	defaultOpts := DefaultActionBaseOpts()
	for _, fn := range opts {
		fn(defaultOpts)
	}
	d.DeploymentClient = clients.NewNuvlaDeploymentClient(
		defaultOpts.JobResource.TargetResource.Href,
		defaultOpts.NuvlaClient)

	d.JobResource = defaultOpts.JobResource

	err := d.DeploymentClient.UpdateResource()
	if err != nil {
		log.Infof("Error updating deployment resource, cannot deploy: %s", err)
		return err
	}

	// For debugging purposes
	d.DeploymentClient.PrintResource()
	d.DeploymentResource = d.DeploymentClient.GetResource()

	// Get the deployment connector
	log.Infof("Creating deployer...")
	dep, err := GetDeployerFromModule(d.DeploymentResource)
	if err != nil {
		log.Errorf("Error getting deployer from module: %s", err)
		return err
	}
	d.deployer = dep

	return nil
}

// --------------------------------------------
// Deployment Start
// --------------------------------------------

type StartDeployment struct {
	DeploymentBase
}

func (d *StartDeployment) ExecuteAction() error {
	log.Infof("Starting deployment %s...", d.DeploymentResource.Id)
	err := d.deployer.StartDeployment()
	if err != nil {
		log.Errorf("Error starting deployment: %s", err)
		_ = d.DeploymentClient.SetState(clients.StateError)
		return err
	}
	log.Infof("Starting deployment %s... Success", d.DeploymentResource.Id)
	_ = d.DeploymentClient.SetStateStarted()
	return nil
}

func (d *StartDeployment) GetActionType() ActionType {
	return StartDeploymentActionType
}

// --------------------------------------------
// Deployment Update
// --------------------------------------------

type UpdateDeployment struct {
	DeploymentBase
}

func (d *UpdateDeployment) ExecuteAction() error {
	log.Infof("Updating deployment %s...", d.DeploymentResource.Id)
	err := d.deployer.UpdateDeployment()
	if err != nil {
		log.Errorf("Error updating deployment: %s", err)
		_ = d.DeploymentClient.SetState(clients.StateError)
		return err
	}
	log.Infof("Updating deployment %s... Success", d.DeploymentResource.Id)
	_ = d.DeploymentClient.SetStateStarted()
	return nil
}

func (d *UpdateDeployment) GetActionType() ActionType {
	return UpdateDeploymentActionType
}

// --------------------------------------------
// Deployment Stop
// --------------------------------------------

type StopDeployment struct {
	DeploymentBase
}

func (d *StopDeployment) ExecuteAction() error {
	log.Infof("Stopping deployment %s...", d.DeploymentResource.Id)
	err := d.deployer.StopDeployment()
	if err != nil {
		log.Errorf("Error stopping deployment: %s", err)
		_ = d.DeploymentClient.SetState(clients.StateError)
		return err
	}
	log.Infof("Stopping deployment %s... Success", d.DeploymentResource.Id)
	_ = d.DeploymentClient.SetState(clients.StateStopped)
	return nil
}

func (d *StopDeployment) GetActionType() ActionType {
	return StopDeploymentActionType
}

// --------------------------------------------
// Deployment State
// --------------------------------------------

type DeploymentState struct {
	DeploymentBase
}

func (d *DeploymentState) ExecuteAction() error {
	return nil
}

func (d *DeploymentState) GetActionType() ActionType {
	return StateDeploymentActionType
}
