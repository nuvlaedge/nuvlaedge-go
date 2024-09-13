package orchestrator

import (
	"context"
	"errors"
	composeTypes "github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/stack/loader"
	"github.com/docker/cli/cli/command/stack/options"
	"github.com/docker/cli/cli/command/stack/swarm"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/common"
	"nuvlaedge-go/types"
	"strings"
)

type Swarm struct {
	dCli command.Cli

	swarmService types.SwarmService
}

func NewSwarmOrchestrator(dClient client.APIClient) (*Swarm, error) {
	dCli, err := command.NewDockerCli(command.WithAPIClient(dClient))
	if err != nil {
		return nil, err
	}

	opts := &flags.ClientOptions{Context: "default", LogLevel: "info"}
	err = dCli.Initialize(opts)
	if err != nil {
		return nil, err
	}

	return &Swarm{
		dCli:         dCli,
		swarmService: &types.CustomSwarmService{},
	}, nil
}

func (s *Swarm) Start(ctx context.Context, opts *types.StartOpts) error {

	dOpts := options.Deploy{
		Namespace:    opts.ProjectName,
		Prune:        true,
		Composefiles: opts.CFiles,
		ResolveImage: swarm.ResolveImageAlways,
		Detach:       false,
		Quiet:        false,
	}

	var env composeTypes.Mapping
	if opts.Env != nil {
		env = make(composeTypes.Mapping)
		for _, e := range opts.Env {
			parts := strings.Split(e, "=")
			if len(parts) != 2 {
				return errors.New("invalid environment variable")
			}
			env[parts[0]] = parts[1]
		}
	}
	return s.start(ctx, dOpts, env)
}

func (s *Swarm) start(ctx context.Context, opts options.Deploy, env composeTypes.Mapping) error {

	cfg, err := loader.LoadComposefile(s.dCli, opts)
	if err != nil {
		return err
	}

	if err := common.ExportEnvs(env); err != nil {
		log.Info("Error exporting environment variables")
		return err
	}
	defer func() {
		if err := common.RemoveEnvs(env); err != nil {
			log.Info("Error removing environment variables")
		}
	}()

	if err := s.swarmService.Deploy(ctx, s.dCli, nil, &opts, cfg); err != nil {
		return err
	}

	return nil
}

func (s *Swarm) Stop(ctx context.Context, opts *types.StopOpts) error {
	return s.stop(ctx, opts.ProjectName)
}

func (s *Swarm) stop(ctx context.Context, projectName string) error {
	return s.swarmService.Remove(ctx, s.dCli, options.Remove{Namespaces: []string{projectName}, Detach: true})
}

func (s *Swarm) Install(ctx context.Context) error {
	return nil
}

func (s *Swarm) Remove(ctx context.Context) error {
	return nil
}

func (s *Swarm) Close() error {
	return s.dCli.Client().Close()
}
