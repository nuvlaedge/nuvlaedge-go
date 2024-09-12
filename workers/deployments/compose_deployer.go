package deployments

import (
	"bytes"
	"context"
	"github.com/docker/compose/v2/cmd/formatter"
	"github.com/nuvla/api-client-go/clients/resources"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/orchestrator"
	"nuvlaedge-go/types"
	"nuvlaedge-go/types/jobs"
	"os"
	"path/filepath"
)

type ComposeDeployer struct {
	*orchestrator.Compose
}

func NewComposeDeployer(compose *orchestrator.Compose) *ComposeDeployer {
	return &ComposeDeployer{compose}
}

func (c *ComposeDeployer) StartDeployment(ctx context.Context, resource *resources.DeploymentResource) error {
	tempDir, err := setUpComposeFiles(resource)
	if err != nil {
		return err
	}

	var envs []string
	for _, v := range resource.Module.Content.EnvironmentVariables {
		envs = append(envs, v.GetAsString())
	}

	opts := &types.StartOpts{
		CFiles:      []string{filepath.Join(tempDir, "docker-compose.yml")},
		ProjectName: GetDeploymentProjectName(resource.Id),
		WorkingDir:  tempDir,
		Env:         envs,
	}

	log.Infof("Starting deployment with opts: %v", opts)

	return c.Start(ctx, opts)
}

func (c *ComposeDeployer) StopDeployment(ctx context.Context, deploymentId string) error {
	opts := &types.StopOpts{
		ProjectName: GetDeploymentProjectName(deploymentId),
	}

	return c.Stop(ctx, opts)
}

func (c *ComposeDeployer) UpdateDeployment(ctx context.Context, resource *resources.DeploymentResource) error {
	return c.StartDeployment(ctx, resource)
}

func (c *ComposeDeployer) GetDeploymentLogs(ctx context.Context, deploymentId string) (*jobs.Logs, error) {
	// Need to receive the client probably (Or return the logs maybe better)
	opts := &types.LogOpts{
		ProjectName: GetDeploymentProjectName(deploymentId),
	}
	stdOut := bytes.Buffer{}
	stdErr := bytes.Buffer{}
	opts.LogConsumer = formatter.NewLogConsumer(ctx, &stdOut, &stdErr, false, false, true)

	err := c.Logs(ctx, opts)
	if err != nil {
		return nil, err
	}
	log.Infof("StdOut: %s", stdOut.String())
	log.Infof("StdErr: %s", stdErr.String())
	// Return the logs
	return nil, nil
}

func (c *ComposeDeployer) GetDeploymentState(ctx context.Context, deploymentId string) error {
	return nil
}

func setUpComposeFiles(deployment *resources.DeploymentResource) (string, error) {
	wDir := GetDeploymentProjectName(deployment.Id)
	// Create temporary directory
	tempDir := filepath.Join("/tmp/", wDir)
	err := os.MkdirAll(tempDir, 0755)
	if err != nil {
		log.Errorf("Error creating temporary directory: %s", err)
		return "", err
	}

	filename := "docker-compose.yml"

	// Write f.FileContent into tempDir + f.FileName
	filePath := filepath.Join(tempDir, filename)
	log.Infof("Writing file: %s", filePath)

	if err := os.WriteFile(filePath, []byte(deployment.Module.Content.DockerCompose), 0644); err != nil {
		return "", err

	}
	return tempDir, nil
}
