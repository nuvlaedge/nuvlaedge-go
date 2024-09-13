package types

import (
	"context"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/stack/options"
	"github.com/docker/cli/cli/command/stack/swarm"
	"github.com/docker/cli/cli/compose/types"
	"github.com/spf13/pflag"
)

type SwarmService interface {
	Deploy(ctx context.Context, dockerCli command.Cli, flags *pflag.FlagSet, opts *options.Deploy, cfg *types.Config) error
	Remove(ctx context.Context, dockerCli command.Cli, opts options.Remove) error
	Ps(ctx context.Context, dockerCli command.Cli, opts options.PS) error
}

type CustomSwarmService struct {
}

func (s *CustomSwarmService) Deploy(ctx context.Context, dockerCli command.Cli, flags *pflag.FlagSet, opts *options.Deploy, cfg *types.Config) error {
	return swarm.RunDeploy(ctx, dockerCli, flags, opts, cfg)
}

func (s *CustomSwarmService) Remove(ctx context.Context, dockerCli command.Cli, opts options.Remove) error {
	return swarm.RunRemove(ctx, dockerCli, opts)
}

func (s *CustomSwarmService) Ps(ctx context.Context, dockerCli command.Cli, opts options.PS) error {
	return swarm.RunPS(ctx, dockerCli, opts)
}

// Compile time check for interface implementation
var _ SwarmService = &CustomSwarmService{}
