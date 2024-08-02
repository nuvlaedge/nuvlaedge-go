package testutils

import (
	"context"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/stack/options"
	composeTypes "github.com/docker/cli/cli/compose/types"
	"github.com/spf13/pflag"
)

type SwarmServiceMock struct {
	DeployCnt int
	DeployErr error

	RemoveCnt int
	RemoveErr error

	PsCnt int
	PsErr error
}

func (s *SwarmServiceMock) Deploy(ctx context.Context, dockerCli command.Cli, flags *pflag.FlagSet, opts *options.Deploy, cfg *composeTypes.Config) error {
	s.DeployCnt++
	return s.DeployErr
}

func (s *SwarmServiceMock) Remove(ctx context.Context, dockerCli command.Cli, opts options.Remove) error {
	s.RemoveCnt++
	return s.RemoveErr
}

func (s *SwarmServiceMock) Ps(ctx context.Context, dockerCli command.Cli, opts options.PS) error {
	s.PsCnt++
	return s.PsErr
}
