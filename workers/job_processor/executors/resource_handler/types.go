package resource_handler

import (
	"context"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"io"
)

//go:generate mockery --name ResourceHandlerDockerClient
type ResourceHandlerDockerClient interface {
	ImagePull(ctx context.Context, ref string, options image.PullOptions) (io.ReadCloser, error)
	ImageRemove(ctx context.Context, image string, options image.RemoveOptions) ([]image.DeleteResponse, error)
	ContainerRemove(ctx context.Context, container string, options container.RemoveOptions) error
	VolumeRemove(ctx context.Context, volumeID string, force bool) error
	NetworkRemove(ctx context.Context, network string) error
}
