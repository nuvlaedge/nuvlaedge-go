package connector

import (
	"context"
	"errors"
	"github.com/docker/docker/api/types/container"
	image2 "github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	"io"
	"nuvlaedge-go/common/constants"
	"nuvlaedge-go/types/jobs"
	"nuvlaedge-go/types/options/command"
	"os"
	"strings"
)

type DockerConnector struct {
	dCli *client.Client
}

func NewDockerConnector() (*DockerConnector, error) {
	dCli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	return &DockerConnector{
		dCli: dCli,
	}, nil
}

func (dc *DockerConnector) Stop() {
	err := dc.dCli.Close()
	if err != nil {
		log.Errorf("Error closing Docker client: %s", err)
	}
}

func (dc *DockerConnector) Reboot(ctx context.Context) error {
	// Run a basic common container with the command "-c 'sleep 10 && echo b > /sysrq'"
	// This will reboot the system after 10 seconds
	_, err := dc.dCli.ContainerCreate(ctx, &container.Config{
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
	return err
}

func (dc *DockerConnector) InstallSSHKey(ctx context.Context, sshPub, user string) error {
	return nil
}

func (dc *DockerConnector) RevokeSSKKey(ctx context.Context, sshkey string) error {
	return nil
}

func (dc *DockerConnector) UpdateNuvlaEdge(ctx context.Context, image string, params *command.UpdateCmdOptions) error {
	res, err := dc.dCli.ContainerCreate(ctx, &container.Config{
		Image: image,
		Cmd:   []string{"update", "--working-dir", params.WorkingDir, "--project", params.Project},
		Env:   params.Environment,
	}, &container.HostConfig{
		AutoRemove: false,
	}, nil, nil, "")
	if err != nil {
		return err
	}

	if err := dc.dCli.ContainerStart(ctx, res.ID, container.StartOptions{}); err != nil {
		return err
	}

	return errors.New("not implemented")
}

func (dc *DockerConnector) FetchNuvlaEdgeLogs(ctx context.Context) error {
	return nil
}

func (dc *DockerConnector) RunLegacyJobEngine(ctx context.Context, conf *jobs.LegacyJobConf) (string, error) {

	// Pull image
	if err := dc.pullAndWaitImage(ctx, conf.Image); err != nil {
		return "", err
	}

	cmd := []string{"--", "/app/job_executor.py",
		"--api-url", conf.Endpoint,
		"--api-key", conf.ApiKey,
		"--api-secret", conf.ApiSecret,
		"--nuvlaedge-fs", "/tmp/nuvlaedge-fs",
		"--job-id", conf.JobId}
	if conf.EndpointInsecure {
		cmd = append(cmd, "--api-insecure")
	}

	envs := GetEnvironWithPrefix("NE_IMAGE_", "JOB_")
	log.Debugf("Passing envs: %v", envs)
	// Create container config
	config := &container.Config{
		Image:        conf.Image,
		Cmd:          cmd,
		AttachStderr: false,
		AttachStdout: false,
		AttachStdin:  false,
		Hostname:     conf.JobId,
		Env:          envs,
	}

	hostConf := &container.HostConfig{
		AutoRemove: true,
		Binds: []string{
			"/var/run/docker.sock:/var/run/docker.sock:rw", // Bind mount Docker socket
		},
	}

	resp, err := dc.dCli.ContainerCreate(
		ctx,
		config,
		hostConf,
		nil,
		nil,
		strings.Replace(conf.JobId, "/", "-", -1))
	if err != nil {
		log.Infof("Error creating container: %s", err)
		return "", err
	}
	log.Infof("Created container: %s, %v", resp.ID, resp.Warnings)

	err = dc.dCli.ContainerStart(ctx, resp.ID, container.StartOptions{})
	if err != nil {
		log.Infof("Error starting container: %s", err)
		return "", err
	}

	return resp.ID, nil
}

func (dc *DockerConnector) pullAndWaitImage(ctx context.Context, image string) error {
	// Pull image
	r, err := dc.dCli.ImagePull(ctx, image, image2.PullOptions{})
	defer r.Close()
	if err != nil {
		return err
	}

	// Wait for image pull to complete
	_, err = io.Copy(io.Discard, r)
	if err != nil {
		return err
	}

	log.Infof("Successfully pulled image %s", image)
	return nil
}

func (dc *DockerConnector) WaitContainerFinish(ctx context.Context, contId string, printLogs bool) error {

	return nil
}

func GetEnvironWithPrefix(prefixes ...string) []string {
	// Get all environment variables
	envs := os.Environ()

	var filteredEnvs []string
	for _, env := range envs {
		for _, filter := range prefixes {
			if strings.HasPrefix(env, filter) {
				filteredEnvs = append(filteredEnvs, env)
				break
			}
		}
	}

	return filteredEnvs
}
