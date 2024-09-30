package engine

import (
	"bufio"
	"context"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	"io"
	"nuvlaedge-go/common"
	"nuvlaedge-go/types/jobs"
	"strings"
	"time"
)

const (
	ImagePullTimeout = 120 * time.Second
)

type DockerEngine struct {
	coeType CoeType
	client  *client.Client
}

func NewDockerEngine() *DockerEngine {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil
	}

	return &DockerEngine{
		coeType: DockerType,
		client:  cli,
	}
}

/**************************************** NuvlaEdge Utils *****************************************/

/**************************************** Struct Utils *****************************************/

// String
func (dc *DockerEngine) String() string {
	return "docker"
}

/********************************* Docker container management functions *************************************/

func (dc *DockerEngine) RunContainer(image string, configuration map[string]string) (string, error) {
	//ctx := context.Background()

	return "", nil
}

type ImagePullResponse struct {
	Status         string `json:"status"`
	ProgressDetail struct {
		Current int `json:"current"`
		Total   int `json:"total"`
	} `json:"progressDetail"`
	Progress string `json:"progress"`
	ID       string `json:"id"`
}

func (dc *DockerEngine) RunJobEngineContainer(conf *jobs.LegacyJobConf) (string, error) {
	if conf.Image == "" {
		conf.Image = "nuvlaedge/job-engine:latest"
	}

	ctx := context.Background()
	// Pull image
	if err := dc.pullAndWaitImage(ctx, conf.Image); err != nil {
		return "", err
	}

	if !strings.HasPrefix(conf.Endpoint, "https://") && !strings.HasPrefix(conf.Endpoint, "http://") {
		conf.Endpoint = "https://" + conf.Endpoint
	}

	command := []string{"--", "/app/job_executor.py",
		"--api-url", conf.Endpoint,
		"--api-key", conf.ApiKey,
		"--api-secret", conf.ApiSecret,
		"--nuvlaedge-fs", "/tmp/nuvlaedge-fs",
		"--job-id", conf.JobId}
	if conf.EndpointInsecure {
		command = append(command, "--api-insecure")
	}

	envs := common.GetEnvironWithPrefix("NE_IMAGE_", "JOB_")
	log.Debugf("Passing envs: %v", envs)
	// Create container config
	config := &container.Config{
		Image:        conf.Image,
		Cmd:          command,
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

	resp, err := dc.client.ContainerCreate(
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

	err = dc.client.ContainerStart(ctx, resp.ID, container.StartOptions{})
	if err != nil {
		log.Infof("Error starting container: %s", err)
		return "", err
	}

	return resp.ID, nil
}

func (dc *DockerEngine) pullAndWaitImage(ctx context.Context, imageName string) error {
	ctxTimed, cancel := context.WithTimeout(ctx, ImagePullTimeout)
	defer cancel()

	// Pull image
	r, err := dc.client.ImagePull(ctxTimed, imageName, image.PullOptions{})
	defer func() {
		err := r.Close()
		if err != nil {
			log.Infof("Error closing image pull response: %s", err)
		}
	}()

	if err != nil {
		return err
	}

	// Wait for image pull to complete
	_, err = io.Copy(io.Discard, r)
	if err != nil {
		return err
	}

	log.Infof("Successfully pulled image %s", imageName)
	return nil

}

func (dc *DockerEngine) GetContainerLogs(containerId, since string) (io.ReadCloser, error) {
	logs, err := dc.client.ContainerLogs(
		context.Background(),
		containerId,
		container.LogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Timestamps: true,
			Since:      since,
		})

	if err != nil {
		return nil, err
	}
	return logs, nil
}

// printLogLines reads the log lines from the reader and prints them to the log.
// It returns the timestamp of the last log line read.
func printLogLines(reader io.ReadCloser) string {
	scanner := bufio.NewScanner(reader)
	var sinceTime string
	for scanner.Scan() {
		logLine := scanner.Text()
		log.Infof("Container log: %s", logLine)

		// Update the sinceTime to the timestamp of the current log line
		logParts := strings.SplitN(logLine, " ", 2)
		if len(logParts) > 0 {
			// Remove any non-timestamp characters from the start of the timestamp
			timestamp := strings.TrimLeft(logParts[0], "\x02\x00\x01\x1b")
			// Update the sinceTime to the timestamp of the current log line
			sinceTime = timestamp
		}
	}
	return sinceTime
}

func (dc *DockerEngine) printLogsUntilFinished(containerId string, exitFlag chan interface{}) {
	var sinceTime string

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-exitFlag:
			return
		case <-ticker.C:
			logs, err := dc.GetContainerLogs(containerId, sinceTime)
			if err != nil {
				log.Infof("Error getting logs: %s", err)
				return
			}
			sinceTime = printLogLines(logs)
			log.Infof("Container logs: %s", logs)
		}
	}
}

func (dc *DockerEngine) GetContainerStatus(containerId string) (string, error) {
	info, err := dc.client.ContainerInspect(context.Background(), containerId)
	if err != nil {
		return "", err
	}
	return info.State.Status, nil
}

func (dc *DockerEngine) WaitContainerFinish(containerId string, timeout time.Duration, printLogs bool) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if printLogs {
		exitFlag := make(chan interface{})
		go dc.printLogsUntilFinished(containerId, exitFlag)
		// TODO: Not sure if this is needed, or it is enough to close the channel
		defer func() {
			exitFlag <- struct{}{}
			close(exitFlag)
		}()
	}

	statusCh, errCh := dc.client.ContainerWait(ctx, containerId, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			log.Infof("Error waiting for container %s to finish: %s", containerId, err)
			return -1, err
		}
	case status := <-statusCh:
		log.Infof("Container %s finished with status: %d", containerId, status.StatusCode)
		return status.StatusCode, nil
	}
	return -1, nil
}

func (dc *DockerEngine) StopContainer(containerId string, force bool) (bool, error) {
	return false, nil
}

func (dc *DockerEngine) RemoveContainer(containerId string, containerName string) (bool, error) {
	return false, nil
}

var _ Coe = &DockerEngine{}
