package jobProcessor

import (
	"fmt"
	"github.com/nuvla/api-client-go/clients"
	log "github.com/sirupsen/logrus"
	"strings"
)

type DeployerType string

const (
	ComposeDeployerType    DeployerType = "compose"
	SwarmDeployerType      DeployerType = "swarm"
	KubernetesDeployerType DeployerType = "kubernetes"
	UnixDeployerType       DeployerType = "unix"
)

type NotSupportedDeployer string

func (n NotSupportedDeployer) Error() string {
	return string(n)
}

func GetDeployerFromModule(res *clients.DeploymentResource) (Deployer, error) {
	module := res.Module

	compatibility := module.Compatibility
	subType := module.SubType

	switch subType {
	case "application":
		log.Infof("Deployment is an application subType meaning it is a docker deployment")
		switch compatibility {

		case "docker-compose":
			log.Infof("Deployment is a docker-compose deployment")
			// Try to get the docker-compose file
			compose, ok := module.Content["docker-compose"]
			if !ok {
				log.Errorf("Error getting docker-compose file from deployment")
				return nil, fmt.Errorf("error getting docker compose file for %s", res.Id)
			}
			return NewComposeDeployer(res.Id, compose.(string)), nil

		case "swarm":
			log.Infof("Deployment is a swarm deployment, which is not supported at the moment")
			return nil, NotSupportedDeployer("swarm deployment is not supported")

		default:
			log.Infof("Deployment is a %s deployment, which is not recognised at the moment", compatibility)
			return nil, NotSupportedDeployer(fmt.Sprintf("docker %s deployment is not supported", compatibility))
		}

	case "application_kubernetes":
		log.Infof("Deployment is a kubernetes deployment, which is not supported at the moment")
		return nil, NotSupportedDeployer("kubernetes deployment is not supported")

	case "application_unix":
		log.Infof("Deployment is a unix deployment, which is not supported at the moment")
		return nil, NotSupportedDeployer("unix deployment is not supported")

	default:
		log.Infof("Deployment is a %s-%s deployment, which is not recognised at the moment", subType, compatibility)
		return nil, NotSupportedDeployer(fmt.Sprintf("%s-%s deployment is not supported", subType, compatibility))
	}
}

type Deployer interface {
	GetType() DeployerType
	StartDeployment() error
	StopDeployment() error
	RemoveDeployment() error
	DeploymentState() error
}

type ComposeDeployer struct {
	fileContent  string
	file         string
	deploymentId string // Deployment ID will act as compose project name
}

func NewComposeDeployer(deploymentId string, fileContent string) *ComposeDeployer {
	return &ComposeDeployer{
		deploymentId: deploymentId,
		fileContent:  fileContent,
		file:         "docker-compose.yml",
	}
}

func (c *ComposeDeployer) GetType() DeployerType {
	return ComposeDeployerType
}

func (c *ComposeDeployer) StartDeployment() error {
	// 1. Write docker-compose file to disk
	deploymentName := strings.Replace(c.deploymentId, "/", "-", -1)
	err := saveFileToDeploymentDir(deploymentName, c.file, c.fileContent)
	if err != nil {
		log.Errorf("Error writing docker-compose file: %s", err)
		return err
	}
	command := []string{"-f", fmt.Sprintf("/tmp/%s/docker-compose.yml", deploymentName), "up", "-d"}
	log.Infof("Executing command: docker compose %s", command)
	_, err = executeCommand("docker-compose", command...)
	if err != nil {
		log.Infof("Error starting deployment: %s", err)
		return err
	}
	return nil
}

func (c *ComposeDeployer) StopDeployment() error {
	deploymentName := strings.Replace(c.deploymentId, "/", "-", -1)
	command := []string{"-f", fmt.Sprintf("/tmp/%s/docker-compose.yml", deploymentName), "down"}
	log.Infof("Executing command: docker compose %s", command)
	_, err := executeCommand("docker-compose", command...)
	if err != nil {
		log.Infof("Error starting deployment: %s", err)
		return err
	}
	return nil
}

func (c *ComposeDeployer) RemoveDeployment() error {
	return nil
}

func (c *ComposeDeployer) DeploymentState() error {
	return nil
}
