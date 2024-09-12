package orchestrator

import (
	"context"
	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/compose"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/types"
	"strings"
)

type Compose struct {
	service types.ComposeService

	// We keep the DockerCLI instance to close it when the orchestrator is closed. Not reachable from
	// the service
	dCli command.Cli
}

func NewComposeOrchestrator(dClient client.APIClient) (*Compose, error) {
	if dClient == nil {
		dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			return nil, err
		}
		dClient = dockerClient
	}
	dCli, err := command.NewDockerCli(command.WithAPIClient(dClient))
	if err != nil {
		return nil, err
	}

	opts := &flags.ClientOptions{Context: "default", LogLevel: "info"}
	err = dCli.Initialize(opts)
	if err != nil {
		return nil, err
	}

	cs := compose.NewComposeService(dCli)
	c := &Compose{
		service: cs,
		dCli:    dCli,
	}
	return c, nil
}

func (c *Compose) Start(ctx context.Context, opts *types.StartOpts) error {
	// Validate opts here for Compose...
	return c.start(ctx, opts.CFiles, opts.Env, opts.ProjectName, opts.WorkingDir)
}

func (c *Compose) start(ctx context.Context, files []string,
	env []string,
	projectName string,
	workingDir string) error {

	pOptions, err := cli.NewProjectOptions(
		files,
		cli.WithWorkingDirectory(workingDir),
		cli.WithOsEnv,
		cli.WithEnv(env),
		cli.WithName(projectName),
		cli.WithDefaultConfigPath)
	if err != nil {
		return err
	}

	project, err := pOptions.LoadProject(ctx)
	if err != nil {
		return err
	}
	log.Info("Project working directory: ", project.WorkingDir)

	for i, s := range project.Services {
		s.CustomLabels = map[string]string{
			api.ProjectLabel:     pOptions.Name,
			api.ServiceLabel:     s.Name,
			api.VersionLabel:     api.ComposeVersion,
			api.WorkingDirLabel:  "/",
			api.ConfigFilesLabel: strings.Join(project.ComposeFiles, ","),
			api.OneoffLabel:      "False", // default, will be overridden by `run` command
		}
		attach := false
		s.Attach = &attach
		project.Services[i] = s
	}

	if err := c.service.Pull(ctx, project, api.PullOptions{}); err != nil {
		return err
	}

	if err = c.service.Up(ctx, project, api.UpOptions{}); err != nil {
		return err
	}

	return nil
}

func (c *Compose) Stop(ctx context.Context, opts *types.StopOpts) error {
	return c.stop(ctx, opts.ProjectName)
}

func (c *Compose) stop(ctx context.Context, projectName string) error {
	if err := c.service.Down(ctx, projectName, api.DownOptions{}); err != nil {
		return err
	}
	return nil
}

func (c *Compose) GetProjectStatus(ctx context.Context, projectName string) (interface{}, error) {

	containers, err := c.service.Ps(ctx, projectName, api.PsOptions{})
	if err != nil {
		return nil, err
	}
	return containers, nil
}

func (c *Compose) List(ctx context.Context) ([]api.Stack, error) {
	return c.service.List(ctx, api.ListOptions{})
}

func (c *Compose) Install(ctx context.Context) error {
	return nil
}

func (c *Compose) Remove(ctx context.Context) error {

	return nil
}

func (c *Compose) Close() error {
	return c.dCli.Client().Close()
}

func (c *Compose) Logs(ctx context.Context, opts *types.LogOpts) error {
	// First, get services from the project

	contSum, err := c.service.Ps(ctx, opts.ProjectName, api.PsOptions{All: true})
	if err != nil {
		return err
	}
	for _, cont := range contSum {
		log.Infof("Container: %s", cont.Name)
		log.Infof("Service: %s", cont.Service)

		reader, err := c.dCli.Client().ContainerLogs(ctx, cont.ID, container.LogsOptions{})
		if err != nil {
			return err
		}
		defer reader.Close()

		// Read the logs
		buf := make([]byte, 1024)
		reader.Read(buf)
	}

	return c.service.Logs(ctx, opts.ProjectName, opts.LogConsumer, opts.LogOptions)
}

var _ Orchestrator = &Compose{}
