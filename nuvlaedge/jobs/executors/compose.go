package executors

import (
	"context"
	"fmt"
	"github.com/compose-spec/compose-go/v2/loader"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	composeAPI "github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/compose"
	"github.com/nuvla/api-client-go/clients/resources"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/nuvlaedge/common"
	"strings"
)

type Compose struct {
	ExecutorBase

	deploymentResource *resources.DeploymentResource
	projectName        string

	tempDir     string
	composePath string

	ctx            context.Context
	composeConfig  *types.ConfigDetails
	composeProject *types.Project
	composeService composeAPI.Service

	// Services summary
	services []*DeploymentService
}

func (c *Compose) StartDeployment() error {
	c.ctx = context.TODO()
	c.projectName = GetProjectNameFromDeploymentId(c.deploymentResource.Id)

	if err := c.prepareComposeUp(); err != nil {
		return err
	}

	if err := c.composeService.Up(c.ctx, c.composeProject, composeAPI.UpOptions{}); err != nil {
		return err
	}

	return nil
}

func (c *Compose) StopDeployment() error {
	c.ctx = context.TODO()
	c.projectName = GetProjectNameFromDeploymentId(c.deploymentResource.Id)

	if err := c.prepareComposeDown(); err != nil {
		return err
	}

	if err := c.composeService.Down(c.ctx, c.projectName, composeAPI.DownOptions{}); err != nil {
		log.Errorf("Error stopping deployment: %s", err)
		return err
	}

	return nil
}

func (c *Compose) GetServices() ([]*DeploymentService, error) {
	c.ctx = context.TODO()
	c.projectName = GetProjectNameFromDeploymentId(c.deploymentResource.Id)
	if err := c.setUpService(); err != nil {
		return nil, err
	}

	containers, err := c.composeService.Ps(c.ctx, c.projectName, composeAPI.PsOptions{
		All: true,
	})

	if err != nil {
		log.Infof("Error getting services: %s", err)
	}

	c.services = make([]*DeploymentService, 0)
	for _, container := range containers {
		c.services = append(c.services, NewDeploymentServiceFromContainerSummary(container))
	}
	return c.services, nil
}

func (c *Compose) StateDeployment() error {
	c.projectName = GetProjectNameFromDeploymentId(c.deploymentResource.Id)
	_, err := c.GetServices()
	if err != nil {
		log.Warnf("Error getting services for deployment %s: %s", c.deploymentResource.Id, err)
	}
	return nil
}

func (c *Compose) UpdateDeployment() error {
	return nil
}

func (c *Compose) getComposeFromDeployment() (string, error) {
	if c.deploymentResource.Module == nil ||
		c.deploymentResource.Module.Content == nil ||
		c.deploymentResource.Module.Content.DockerCompose == "" {
		return "", NewComposeNotAvailableError(c.deploymentResource.Id, string(c.GetName()))
	}

	return c.deploymentResource.Module.Content.DockerCompose, nil
}

func (c *Compose) setUpProjectConfig() error {
	composeContent, err := c.getComposeFromDeployment()
	if err != nil {
		return err
	}
	c.composeConfig = &types.ConfigDetails{
		WorkingDir: c.tempDir,
		ConfigFiles: []types.ConfigFile{
			{Filename: "docker-compose.yml",
				Content: []byte(composeContent)},
		},
		Environment: nil,
	}
	return nil
}

func (c *Compose) setUpProject() error {
	if c.composeConfig == nil {
		return fmt.Errorf("compose config is not set, cannot create the project")
	}
	p, err := loader.LoadWithContext(c.ctx, *c.composeConfig, func(options *loader.Options) {
		options.SetProjectName(c.projectName, true)
	})
	if err != nil {
		return err
	}

	for i, s := range p.Services {
		s.CustomLabels = map[string]string{
			composeAPI.ProjectLabel:     p.Name,
			composeAPI.ServiceLabel:     s.Name,
			composeAPI.VersionLabel:     composeAPI.ComposeVersion,
			composeAPI.WorkingDirLabel:  "/",
			composeAPI.ConfigFilesLabel: strings.Join(p.ComposeFiles, ","),
			composeAPI.OneoffLabel:      "False", // default, will be overridden by `run` command
			"nuvla.deployment":          c.deploymentResource.Id,
		}
		attach := false
		s.Attach = &attach
		p.Services[i] = s
	}
	c.composeProject = p

	return nil
}

func (c *Compose) setUpService() error {
	dockerCli, err := command.NewDockerCli()
	if err != nil {
		return err
	}

	myOpts := &flags.ClientOptions{Context: "default", LogLevel: common.LogLevel.String()}
	err = dockerCli.Initialize(myOpts)
	if err != nil {
		return err
	}

	c.composeService = compose.NewComposeService(dockerCli)
	return nil
}

func (c *Compose) prepareComposeUp() error {
	if err := c.setUpProjectConfig(); err != nil {
		return err
	}

	if err := c.setUpProject(); err != nil {
		return err
	}

	if err := c.setUpService(); err != nil {
		return err
	}
	return nil
}

func (c *Compose) prepareComposeDown() error {
	if err := c.setUpService(); err != nil {
		return err
	}
	return nil
}
