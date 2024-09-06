package jobs

import (
	"context"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/common/constants"
)

type DockerExecutor struct {
	dCli client.APIClient
}

func NewDockerExecutor(dCli client.APIClient) *DockerExecutor {
	return &DockerExecutor{
		dCli: dCli,
	}
}

func (d *DockerExecutor) Reboot(ctx context.Context) error {
	// Run a basic common container with the command "-c 'sleep 10 && echo b > /sysrq'"
	// This will reboot the system after 10 seconds
	containerConf := &container.Config{
		Image: constants.BaseImage,
		Cmd:   []string{"sh", "-c", "sleep 10 && echo b > /sysrq"},
	}

	hostConf := &container.HostConfig{
		AutoRemove: true,
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: "/proc/sysrq-trigger",
				Target: "/sysrq",
			},
		},
	}
	// Pull Image
	b, err := d.dCli.ImagePull(ctx, constants.BaseImage, image.PullOptions{})

	if err != nil {
		return err
	}

	err = b.Close()
	if err != nil {
		log.Warn("Error closing image pull response")
		err = nil
	}

	res, err := d.dCli.ContainerCreate(ctx, containerConf, hostConf, nil, nil, "")
	if err != nil {
		return err
	}

	err = d.dCli.ContainerStart(ctx, res.ID, container.StartOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (d *DockerExecutor) AddSSHKey(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (d *DockerExecutor) RevokeSSHKey(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (d *DockerExecutor) Update(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (d *DockerExecutor) LogFetch(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

type HostExecutor struct{}

func (h *HostExecutor) Reboot(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (h *HostExecutor) AddSSHKey(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (h *HostExecutor) RevokeSSHKey(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (h *HostExecutor) Update(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (h *HostExecutor) LogFetch(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}
