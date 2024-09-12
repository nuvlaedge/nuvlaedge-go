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
	"strings"
)

type ComposeExecutor struct {
	ExecutorBase

	deploymentResource *resources.DeploymentResource
	projectName        string

	tempDir     string
	composePath string

	ctx            context.Context
	composeConfig  *types.ConfigDetails
	composeProject *types.Project
	composeService composeAPI.Service
	dockerCli      *command.DockerCli
}

func (ce *ComposeExecutor) StartDeployment() error {
	ce.ctx = context.TODO()
	ce.projectName = GetProjectNameFromDeploymentId(ce.deploymentResource.Id)

	if err := ce.prepareComposeUp(); err != nil {
		return err
	}

	if err := ce.composeService.Up(ce.ctx, ce.composeProject, composeAPI.UpOptions{}); err != nil {
		return err
	}

	return nil
}

func (ce *ComposeExecutor) StopDeployment() error {
	ce.ctx = context.TODO()
	ce.projectName = GetProjectNameFromDeploymentId(ce.deploymentResource.Id)

	if err := ce.prepareComposeDown(); err != nil {
		return err
	}

	if err := ce.composeService.Down(ce.ctx, ce.projectName, composeAPI.DownOptions{}); err != nil {
		log.Errorf("Error stopping deployment: %s", err)
		return err
	}

	return nil
}

func (ce *ComposeExecutor) GetServices() ([]DeploymentService, error) {
	ce.projectName = GetProjectNameFromDeploymentId(ce.deploymentResource.Id)
	if err := ce.setUpService(); err != nil {
		return nil, err
	}
	defer ce.dockerCli.Client().Close()

	containers, err := ce.composeService.Ps(ce.ctx, ce.projectName, composeAPI.PsOptions{
		All: true,
	})
	for _, container := range containers {
		log.Infof("Container: %s", container.Name)
	}

	if err != nil {
		log.Infof("Error getting services: %s", err)
	}

	services := make([]DeploymentService, 0)
	for _, container := range containers {
		services = append(services, NewDeploymentServiceFromContainerSummary(container))
	}
	return services, nil
}

func (ce *ComposeExecutor) StateDeployment() error {
	ce.projectName = GetProjectNameFromDeploymentId(ce.deploymentResource.Id)
	_, err := ce.GetServices()
	if err != nil {
		log.Warnf("Error getting services for deployment %s: %s", ce.deploymentResource.Id, err)
	}
	return nil
}

func (ce *ComposeExecutor) UpdateDeployment() error {
	return ce.StartDeployment()
}

func (ce *ComposeExecutor) getComposeFromDeployment() (string, error) {
	if ce.deploymentResource.Module == nil ||
		ce.deploymentResource.Module.Content == nil ||
		ce.deploymentResource.Module.Content.DockerCompose == "" {
		return "", NewComposeNotAvailableError(ce.deploymentResource.Id, string(ce.GetName()))
	}

	return ce.deploymentResource.Module.Content.DockerCompose, nil
}

func (ce *ComposeExecutor) setUpProjectConfig() error {
	composeContent, err := ce.getComposeFromDeployment()
	if err != nil {
		return err
	}
	ce.composeConfig = &types.ConfigDetails{
		WorkingDir: ce.tempDir,
		ConfigFiles: []types.ConfigFile{
			{Filename: "docker-compose.yml",
				Content: []byte(composeContent)},
		},
		Environment: nil,
	}

	if ce.deploymentResource.Module.Content.EnvironmentVariables != nil {
		ce.composeConfig.Environment = GetEnvironmentMappingFromContent(ce.deploymentResource.Module.Content)
	}
	return nil
}

func GetEnvironmentMappingFromContent(content *resources.ModuleApplicationResource) types.Mapping {
	envMap := make(types.Mapping)
	for _, e := range content.EnvironmentVariables {
		if e.Value != "" {
			envMap[e.Name] = e.Value
		}
	}
	return envMap
}

func (ce *ComposeExecutor) setUpProject() error {
	if ce.composeConfig == nil {
		return fmt.Errorf("compose config is not set, cannot create the project")
	}
	p, err := loader.LoadWithContext(ce.ctx, *ce.composeConfig, func(options *loader.Options) {
		options.SetProjectName(ce.projectName, true)
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
			"nuvla.deployment":          ce.deploymentResource.Id,
		}
		attach := false
		s.Attach = &attach
		p.Services[i] = s
	}
	ce.composeProject = p

	return nil
}

func (ce *ComposeExecutor) setUpService() error {
	dockerCli, err := command.NewDockerCli()
	if err != nil {
		return err
	}
	ce.dockerCli = dockerCli

	myOpts := &flags.ClientOptions{Context: "default", LogLevel: "info"}
	err = ce.dockerCli.Initialize(myOpts)
	if err != nil {
		return err
	}

	ce.composeService = compose.NewComposeService(dockerCli)
	return nil
}

func (ce *ComposeExecutor) prepareComposeUp() error {
	if err := ce.setUpProjectConfig(); err != nil {
		return err
	}

	if err := ce.setUpProject(); err != nil {
		return err
	}

	if err := ce.setUpService(); err != nil {
		return err
	}
	return nil
}

func (ce *ComposeExecutor) prepareComposeDown() error {
	if err := ce.setUpService(); err != nil {
		return err
	}
	return nil
}

func (ce *ComposeExecutor) Close() error {
	if ce.dockerCli != nil {
		return ce.dockerCli.Client().Close()
	}
	return nil
}
