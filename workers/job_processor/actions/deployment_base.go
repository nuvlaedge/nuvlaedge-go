package actions

import (
	"encoding/json"
	"errors"
	"fmt"
	nuvla "github.com/nuvla/api-client-go"
	"github.com/nuvla/api-client-go/clients"
	"github.com/nuvla/api-client-go/clients/resources"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/types/metrics"
	"nuvlaedge-go/workers/job_processor/executors"
	"strings"
)

type DeploymentBase struct {
	ActionBase

	deploymentId       string
	deploymentResource *resources.DeploymentResource
	client             *clients.NuvlaDeploymentClient
	nuvlaClient        *nuvla.NuvlaClient

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
	// Retrieve deployment ID from jobs resource
	opts := GetActionOpts(optsFn...)
	if opts.JobResource == nil || opts.Client == nil {
		return fmt.Errorf("jobs resource or client not available")
	}

	d.deploymentId = opts.JobResource.TargetResource.Href

	d.nuvlaClient = opts.Client

	// Create deployment client and update deployment resource
	d.client = clients.NewNuvlaDeploymentClient(d.deploymentId, opts.Client)
	if err := d.client.UpdateResource(); err != nil {
		return err
	}
	d.deploymentResource = d.client.GetResource()

	// After retrieving the deployment resource, update the session with the deployment credentials.
	// Features such as deployment-parameters are only available for deployments and users and the received clients
	// is logged in as a NuvlaEdge
	if err := d.client.UpdateSessionFromDeploymentCredentials(); err != nil {
		log.Errorf("Error refleshing session from deployment credentials: %s", err)
		return err
	}

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

func (d *DeploymentBase) ManageHostNameParam(ip string) error {
	return d.client.UpdateParameter(
		d.deploymentResource.Owner,
		resources.WithParent(d.deploymentResource.Id),
		resources.WithValue(ip),
		resources.WithName("hostname"),
		resources.WithDescription("Hostname or IP to access the service."))
}

func (d *DeploymentBase) ManageIPsParams(ips deploymentIps) error {
	// Iterate over ips.Network.ips and create a parameter for each one
	var itIp map[string]string
	b, err := json.Marshal(ips.Network.IPs)
	if err != nil {
		return fmt.Errorf("error marshaling IPs: %s", err)
	}

	if err := json.Unmarshal(b, &itIp); err != nil {
		return fmt.Errorf("error unmarshaling IPs: %s", err)
	}

	if len(itIp) == 0 {
		log.Infof("No IP addresses available for deployment %s", d.deploymentId)
		return nil
	}

	var errList []error
	for k, v := range itIp {
		if v == "" || k == "" {
			log.Debugf("IP address %s is empty", k)
			continue
		}

		k = strings.ToLower(k)
		paramName := fmt.Sprintf("ip.%s", k)

		err := d.client.UpdateParameter(
			d.deploymentResource.Owner,
			resources.WithParent(d.deploymentResource.Id),
			resources.WithName(paramName),
			resources.WithValue(v))

		if err != nil {
			errList = append(errList, err)
		}
	}

	return errors.Join(errList...)
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

// manageServiceParameters updates the parameters corresponding to the services started by the deployment
func (d *DeploymentBase) manageServiceParameters(services []executors.DeploymentService) error {
	for _, s := range services {
		if err := d.updateServiceParameter(s); err != nil {
			log.Warnf("Error updating service %s parameter: %s", s.GetServiceMap()["name"], err)
		}
	}
	return nil
}

func (d *DeploymentBase) updateParamInCurrentDeployment(paramName, value, nodeId string) error {
	return d.client.UpdateParameter(
		d.deploymentResource.Owner,
		resources.WithParent(d.deploymentResource.Id),
		resources.WithName(paramName),
		resources.WithValue(value),
		resources.WithNodeId(nodeId))
}

func (d *DeploymentBase) updateServiceParameter(s executors.DeploymentService) error {
	serviceMap := s.GetServiceMap()
	nodeId := serviceMap["node-id"]
	for k, v := range s.GetServiceMap() {
		paramName := fmt.Sprintf("%s.%s", nodeId, k)
		if err := d.updateParamInCurrentDeployment(paramName, v, nodeId); err != nil {
			log.Warnf("Error updating parameter %s: %s", paramName, err)
		}
	}

	for k, v := range s.GetPorts() {
		paramName := fmt.Sprintf("%s.%s", nodeId, k)
		if err := d.updateParamInCurrentDeployment(paramName, fmt.Sprintf("%d", v), nodeId); err != nil {
			log.Warnf("Error updating parameter %s: %s", paramName, err)
		}
	}

	return nil
}

func CloseDeploymentClientWithLog(client *clients.NuvlaDeploymentClient) {
	if err := client.Logout(); err != nil {
		log.Errorf("Error logging out deployment client: %s", err)
	}
}

func (d *DeploymentBase) CreateUserOutputParams() {
	// Fixed parameters for all deployments, hostname and IPs. TODO: IP should be created by Nuvla...
	ips, err := d.getIps()
	if err == nil {
		if err := d.ManageHostNameParam(ips.IP); err != nil {
			log.Warnf("Error creating hostname parameter: %s", err)
		}

		if err := d.ManageIPsParams(ips); err != nil {
			log.Warnf("Error creating IPs parameters: %s", err)
		}

	} else {
		log.Warnf("Error getting IPs: %s", err)
	}

	if err := d.ManageDeploymentParameters(); err != nil {
		log.Warnf("Error creating deployment parameters: %s", err)
	}
}

func (d *DeploymentBase) getIps() (deploymentIps, error) {
	var ips deploymentIps

	neId := d.deploymentResource.Nuvlabox
	log.Debugf("Getting IPs for NuvlaBox %s", neId)

	cli := d.nuvlaClient

	neRes, err := cli.Get(neId, []string{"nuvlabox-status"})
	if err != nil {
		log.Errorf("Error getting NuvlaBox %s: %s", neId, err)
		return ips, fmt.Errorf("error getting NuvlaBox Status ID from Ne %s: %s", neId, err)
	}

	neStatusId, ok := neRes.Data["nuvlabox-status"]
	if !ok {
		log.Errorf("NuvlaBox %s does not have a status", neId)
		return ips, fmt.Errorf("NuvlaBox %s does not have a status", neId)
	}

	neStatusRes, err := cli.Get(neStatusId.(string), []string{"ip", "network"})
	if err != nil {
		log.Errorf("Error getting NuvlaBox %s status: %s", neId, err)
		return ips, fmt.Errorf("error getting NuvlaBox %s status: %s", neId, err)
	}

	neIp, ipFound := neStatusRes.Data["ip"]
	neNetwork, netFound := neStatusRes.Data["network"]

	if !ipFound && !netFound {
		return ips, fmt.Errorf("NuvlaBox %s status does not have an IP or network", neId)
	}

	if ipFound {
		ips.IP = neIp.(string)
	}

	if netFound {
		b, err := json.Marshal(neNetwork)
		if err == nil {
			_ = json.Unmarshal(b, &ips.Network)
		}
	}

	return ips, nil
}

type deploymentIps struct {
	IP      string `json:"ip"`
	Network metrics.NetworkMetrics
}
