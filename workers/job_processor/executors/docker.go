package executors

import (
	"context"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	docker "github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
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

	result, err := client.ImagePull(ctx, constants.BaseImage, image.PullOptions{})
	if err != nil {
		log.Errorf("Failed to pull image: %s", err)
		return err
	}

	if err = result.Close(); err != nil {
		log.Warnf("Failed to close image pull response: %s", err)
	}

	// Run a basic common container with the command "-c 'sleep 10 && echo b > /sysrq'"
	// This will reboot the system after 10 seconds
	response, err := client.ContainerCreate(ctx, &container.Config{
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

	if err != nil {
		return err
	}

	log.Infof("Reboot container created: %s", response.ID)
	if len(response.Warnings) > 0 {
		for _, warning := range response.Warnings {
			log.Warnf("Warning creting Reboot container: %s", warning)
		}
	}

	return client.ContainerStart(ctx, response.ID, container.StartOptions{})
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
