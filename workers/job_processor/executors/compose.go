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
	"io"
	"nuvlaedge-go/common"
	"strings"
)

type ComposeExecutor struct {
	ExecutorBase

	deploymentResource *resources.DeploymentResource
	projectName        string

	tempDir string

	composeConfig  *types.ConfigDetails
	composeProject *types.Project
	composeService composeAPI.Service
	dockerCli      *command.DockerCli

	dockerOutPut io.Writer
}

func (ce *ComposeExecutor) StartDeployment(ctx context.Context) error {

	ce.projectName = GetProjectNameFromDeploymentId(ce.deploymentResource.Id)

	if err := ce.prepareComposeUp(ctx); err != nil {
		return err
	}

	if err := ce.composeService.Up(ctx, ce.composeProject, composeAPI.UpOptions{}); err != nil {
		return err
	}

	return nil
}

func (ce *ComposeExecutor) StopDeployment(ctx context.Context) error {
	ce.projectName = GetProjectNameFromDeploymentId(ce.deploymentResource.Id)

	if err := ce.prepareComposeDown(); err != nil {
		return err
	}

	if err := ce.composeService.Down(ctx, ce.projectName, composeAPI.DownOptions{}); err != nil {
		log.Errorf("Error stopping deployment: %s", err)
		return err
	}

	return nil
}

func (ce *ComposeExecutor) GetServices(ctx context.Context) ([]DeploymentService, error) {
	ce.projectName = GetProjectNameFromDeploymentId(ce.deploymentResource.Id)

	if err := ce.setUpService(); err != nil {
		return nil, err
	}
	defer ce.dockerCli.Client().Close()

	containers, err := ce.composeService.Ps(ctx, ce.projectName, composeAPI.PsOptions{
		All: true,
	})
	if err != nil {
		log.Infof("Error getting services: %s", err)
	}

	services := make([]DeploymentService, 0)
	for _, container := range containers {
		services = append(services, NewDeploymentServiceFromContainerSummary(container))
	}

	return services, nil
}

func (ce *ComposeExecutor) StateDeployment(_ context.Context) error {
	ce.projectName = GetProjectNameFromDeploymentId(ce.deploymentResource.Id)

	return nil
}

func (ce *ComposeExecutor) UpdateDeployment(ctx context.Context) error {
	return ce.StartDeployment(ctx)
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

func (ce *ComposeExecutor) setUpProject(ctx context.Context) error {
	if ce.composeConfig == nil {
		return fmt.Errorf("compose config is not set, cannot create the project")
	}
	p, err := loader.LoadWithContext(ctx, *ce.composeConfig, func(options *loader.Options) {
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
	ce.dockerOutPut = NewCaptureWriter()

	dockerCli, err := command.NewDockerCli(command.WithCombinedStreams(ce.dockerOutPut))

	if err != nil {
		return err
	}
	ce.dockerCli = dockerCli

	myOpts := &flags.ClientOptions{Context: "default", LogLevel: common.LogLevel.String()}
	err = ce.dockerCli.Initialize(myOpts)
	if err != nil {
		return err
	}

	ce.composeService = compose.NewComposeService(dockerCli)
	return nil
}

func (ce *ComposeExecutor) prepareComposeUp(ctx context.Context) error {
	if err := ce.setUpProjectConfig(); err != nil {
		return err
	}

	if err := ce.setUpProject(ctx); err != nil {
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

func (ce *ComposeExecutor) GetOutput() string {
	return fmt.Sprintf("%s", ce.dockerOutPut)
}

var _ Deployer = &ComposeExecutor{}
