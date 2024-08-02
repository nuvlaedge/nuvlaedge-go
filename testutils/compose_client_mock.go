package testutils

import (
	"context"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/compose/v2/pkg/api"
)

type ComposeClientMock struct {
	UpCnt int
	UpErr error

	DownCnt int
	DownErr error

	PsCnt    int
	PsErr    error
	PsReturn []api.ContainerSummary

	PullCnt int
	PullErr error

	ListCnt    int
	ListErr    error
	ListReturn []api.Stack
}

func (c *ComposeClientMock) Up(_ context.Context, _ *types.Project, opts api.UpOptions) error {
	c.UpCnt++
	return c.UpErr
}

func (c *ComposeClientMock) Down(_ context.Context, _ string, opts api.DownOptions) error {
	c.DownCnt++
	return c.DownErr
}

func (c *ComposeClientMock) Ps(_ context.Context, _ string, opts api.PsOptions) ([]api.ContainerSummary, error) {
	c.PsCnt++
	return c.PsReturn, c.PsErr
}

func (c *ComposeClientMock) Pull(_ context.Context, _ *types.Project, opts api.PullOptions) error {
	c.PullCnt++
	return c.PullErr
}

func (c *ComposeClientMock) List(_ context.Context, opts api.ListOptions) ([]api.Stack, error) {
	c.ListCnt++
	return c.ListReturn, c.ListErr
}
