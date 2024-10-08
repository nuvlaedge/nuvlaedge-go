package executors

import (
	"context"
	"fmt"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/stack/loader"
	"github.com/docker/cli/cli/command/stack/options"
	"github.com/docker/cli/cli/command/stack/swarm"
	composetypes "github.com/docker/cli/cli/compose/types"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/cli/opts"
	"github.com/nuvla/api-client-go/clients/resources"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"io"
	"nuvlaedge-go/updater/common"
	"os"
	"path/filepath"
)

type Stack struct {
	ExecutorBase

	deploymentResource *resources.DeploymentResource
	projectName        string

	stackConfig *composetypes.Config
	dockerCli   *command.DockerCli
	stackOpts   *options.Deploy

	// Temporary file locations
	tempDir     string
	composeFile string

	dockerOutPut io.Writer
}

func (s *Stack) StartDeployment(ctx context.Context) error {
	s.projectName = GetProjectNameFromDeploymentId(s.deploymentResource.Id)
	log.Infof("Starting deployment for project %s", s.projectName)

	// Prepare docker client and context. Cannot fail
	if err := s.setUpDockerCLI(); err != nil {
		return err
	}

	// Prepare config files. Cannot fail
	if err := s.setUpFiles(); err != nil {
		return err
	}
	for _, s := range s.stackConfig.Services {
		log.Infof("Starting Stack service %s", s.Name)
	}
	defer s.CleanUp()

	// Start deployment
	if err := s.deploy(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Stack) StopDeployment(ctx context.Context) error {
	s.projectName = GetProjectNameFromDeploymentId(s.deploymentResource.Id)

	if err := s.setUpDockerCLI(); err != nil {
		return err
	}

	if err := s.remove(ctx); err != nil {
		log.Errorf("Error removing stack: %s", err)
		return err
	}

	return nil
}

func (s *Stack) StateDeployment(_ context.Context) error {
	return nil
}

func (s *Stack) UpdateDeployment(ctx context.Context) error {
	return s.StartDeployment(ctx)
}

func (s *Stack) GetServices(ctx context.Context) ([]DeploymentService, error) {
	if err := s.setUpDockerCLI(); err != nil {
		return nil, err
	}
	s.projectName = GetProjectNameFromDeploymentId(s.deploymentResource.Id)

	//var services []*DeploymentComposeService
	swarmServices, err := swarm.GetServices(ctx, s.dockerCli, options.Services{
		Namespace: s.projectName,
		Format:    "json",
		Filter:    opts.NewFilterOpt(),
	})

	if err != nil {
		log.Error("Error retrieving stack services")
		return nil, err
	}

	services := make([]DeploymentService, 0)
	for _, ser := range swarmServices {
		services = append(services, NewDeploymentStackServiceFromServiceSummary(ser))
	}

	return services, nil
}

func (s *Stack) deploy(ctx context.Context) error {
	// Deploy the stack
	err := swarm.RunDeploy(ctx, s.dockerCli, &pflag.FlagSet{}, s.stackOpts, s.stackConfig)
	if err != nil {
		log.Errorf("Error deploying stack: %s", err)
		return err
	}

	return nil
}

func (s *Stack) remove(ctx context.Context) error {
	err := swarm.RunRemove(
		ctx,
		s.dockerCli,
		options.Remove{
			Namespaces: []string{s.projectName},
			Detach:     true,
		})

	return err
}

func (s *Stack) setUpStackOpts() {
	s.stackOpts = &options.Deploy{
		Composefiles: []string{s.composeFile},
		Namespace:    s.projectName,
		Prune:        true,
		Detach:       true,
		Quiet:        true,
		ResolveImage: swarm.ResolveImageAlways,
	}
}

func (s *Stack) setUpFiles() error {
	if s.deploymentResource.Module.Content.DockerCompose == "" {
		return fmt.Errorf("no docker-compose file provided")
	}

	// Write docker-compose file
	if err := s.buildTempDir(); err != nil {
		return err
	}

	s.composeFile = filepath.Join(s.tempDir, "docker-compose.yml")

	contentWithEnv := ExpandEnvMapWithDefaults(
		s.deploymentResource.Module.Content.DockerCompose,
		GetEnvironmentMappingFromContent(s.deploymentResource.Module.Content))
	log.Infof("Writing docker-compose file to %s", contentWithEnv)
	err := common.SaveFile(s.composeFile, "", contentWithEnv)
	if err != nil {
		return err
	}
	// Prepare stack options
	s.setUpStackOpts()

	c, err := loader.LoadComposefile(s.dockerCli, *s.stackOpts)
	if err != nil {
		log.Errorf("Error loading compose file: %s", err)
		return err
	}
	s.stackConfig = c

	// Setup config files if they exist
	if s.deploymentResource.Module.Content.Files != nil {
		log.Infof("Processing config files")
		for _, f := range s.deploymentResource.Module.Content.Files {
			err = common.SaveFile(f.FileName, s.tempDir, f.FileContent)
			if err != nil {
				log.Errorf("Error writing file %s: %s", f.FileName, err)
			}
		}
	}

	return nil
}

// setUpDockerCLI prepares the docker client and its context
func (s *Stack) setUpDockerCLI() error {
	s.dockerOutPut = NewCaptureWriter()

	dCli, err := command.NewDockerCli(command.WithCombinedStreams(s.dockerOutPut))
	if err != nil {
		log.Errorf("Error creating docker cli")
		return err
	}

	err = dCli.Initialize(&flags.ClientOptions{
		Context:  "default",
		LogLevel: "info"},
	)
	if err != nil {
		log.Errorf("Error initializing docker cli")
		return err
	}

	s.dockerCli = dCli

	return nil
}

func (s *Stack) Close() error {
	if s.dockerCli != nil {
		return s.dockerCli.Client().Close()
	}
	return nil
}

func (s *Stack) CleanUp() {
	// Clean up temporary files
	if err := os.RemoveAll(s.tempDir); err != nil {
		log.Errorf("Error cleaning up temporary directory %s: %s", s.tempDir, err)
	}
}

func (s *Stack) buildTempDir() error {
	// Create temporary directory
	s.setTempDirName()
	err := os.MkdirAll(s.tempDir, 0750)
	if err != nil {
		return err
	}
	return nil
}

func (s *Stack) GetOutput() string {
	return fmt.Sprintf("%s", s.dockerOutPut)
}

func (s *Stack) setTempDirName() string {
	s.tempDir = filepath.Join(DefaultTemporaryDirectory, s.projectName)
	return s.tempDir
}
