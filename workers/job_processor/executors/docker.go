package executors

import (
	"context"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	docker "github.com/docker/docker/client"
	"nuvlaedge-go/common/constants"
)

// Docker is the executor to use when running in a docker container and replaces host executor.
type Docker struct {
	ExecutorBase
}

func (d *Docker) Reboot() error {
	ctx := context.Background()
	client, err := docker.NewClientWithOpts(docker.FromEnv, docker.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	// Run a basic common container with the command "-c 'sleep 10 && echo b > /sysrq'"
	// This will reboot the system after 10 seconds
	_, err = client.ContainerCreate(ctx, &container.Config{
		Image: constants.BaseImage,
		Cmd:   []string{"sh", "-c", "sleep 10 && echo b > /sysrq"},
	}, &container.HostConfig{
		AutoRemove: true,
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: "/proc/sysrq-trigger",
				Target: "/sysrq",
			},
		},
	}, nil, nil, "")
	return nil
}

func (d *Docker) InstallSSHKey(sshPub, user string) error {
	return nil
}

func (d *Docker) RevokeSSKKey(sshkey string) error {
	return nil
}

func (d *Docker) UpdateNuvlaEdge() error {

	return nil
}
