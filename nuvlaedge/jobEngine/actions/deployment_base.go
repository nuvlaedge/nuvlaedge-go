package actions

import (
	"fmt"
	"github.com/nuvla/api-client-go/clients"
	"github.com/nuvla/api-client-go/clients/resources"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/nuvlaedge/jobEngine/executors"
	"os"
)

type DeploymentBase struct {
	ActionBase

	deploymentId       string
	deploymentResource *resources.DeploymentResource
	client             *clients.NuvlaDeploymentClient

	ipAddresses []string

	executor executors.Deployer
}

func (d *DeploymentBase) assertExecutor() error {
	ex, err := executors.GetDeployer(d.deploymentResource)
	if err != nil {
		return err
	}
	d.executor = ex
	log.Infof("Deployment action executor set to: %s", d.executor.GetName())
	return nil
}

func (d *DeploymentBase) Init(optsFn ...ActionOptsFn) error {
	// Retrieve deployment ID from job resource
	opts := GetActionOpts(optsFn...)
	if opts.JobResource == nil || opts.Client == nil {
		return fmt.Errorf("job resource or client not available")
	}

	d.deploymentId = opts.JobResource.TargetResource.Href

	// Create deployment client and update deployment resource
	d.client = clients.NewNuvlaDeploymentClient(d.deploymentId, opts.Client)
	if err := d.client.UpdateResource(); err != nil {
		return err
	}
	d.deploymentResource = d.client.GetResource()

	err := d.assertExecutor()
	if err != nil {
		log.Errorf("Error asserting executor: %s", err)
		return err
	}

	// If IPs are available, save them but not fail otherwise
	if opts.IPs != nil {
		d.ipAddresses = opts.IPs
	}

	return nil
}

func (d *DeploymentBase) ManageHostNameParam() error {
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}
	err = d.client.UpdateParameter(
		d.deploymentResource.Owner,
		resources.WithParent(d.deploymentResource.Id),
		resources.WithValue(hostname),
		resources.WithName("hostname"),
		resources.WithDescription("Hostname or IP to access the service."))
	return err
}

func (d *DeploymentBase) ManageIPsParams() error {
	// TODO: This is intended to extract the IP addresses from nuvlabox-status resource in Nuvla.
	// It should probably be done either be reading from a local file or received as a Job parameter.
	return nil
}

// getDeploymentParameters tries to retrieve the deployment parameters from the deployment resource. It's main purpose is
// preventing null pointer exceptions when trying to access the deployment parameters.
func (d *DeploymentBase) getDeploymentParameters() ([]resources.OutputParameter, error) {
	if d.deploymentResource.Module == nil || d.deploymentResource.Module.Content == nil || d.deploymentResource.Module.Content.OutputParameters == nil {
		return nil, fmt.Errorf("output parameters not available in deployment resource")
	}
	return d.deploymentResource.Module.Content.OutputParameters, nil

}

// ManageDeploymentParameters creates or updates the deployment parameters in the deployment resource if any available
func (d *DeploymentBase) ManageDeploymentParameters() error {
	params, err := d.getDeploymentParameters()
	if err != nil {
		return err
	}

	if len(params) == 0 {
		// TODO: This log should be debug at some point
		log.Infof("No deployment parameters available in deployment resource %s", d.deploymentResource.Id)
		return nil
	}

	for _, p := range params {
		if err := d.client.UpdateParameter(
			d.deploymentResource.Owner,
			resources.WithParent(d.deploymentResource.Id),
			resources.WithName(p.Name),
			resources.WithDescription(p.Description)); err != nil {
			log.Warnf("Error creating parameter %s: %s", p.Name, err)
		}
	}
	return nil
}
