package updater

import (
	"context"
	"errors"
	"fmt"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/orchestrator"
	"nuvlaedge-go/types"
	"nuvlaedge-go/types/options/command"
	"nuvlaedge-go/updater/release"
	"strings"
	"time"
)

type Updater func(opts *command.UpdateCmdOptions) error

func GetUpdater() Updater {
	// This is a comment

	return UpdateWithCompose
}

/*
Three types of updates depending on the NuvlaEdge version:
- Kubernetes/Helm
- Docker
- Host
*/

type UpdaterI interface {
	Update(ctx context.Context, opts *command.UpdateCmdOptions) error
	ValidateOpts() error
}

type DockerUpdater struct {
	// cs is the compose service client
	cs   *orchestrator.Compose
	dCli client.APIClient
	opts *command.UpdateCmdOptions
}

func NewDockerUpdater(opts *command.UpdateCmdOptions) (*DockerUpdater, error) {
	dCli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	o, err := orchestrator.NewComposeOrchestrator(dCli)
	if err != nil {
		return nil, err
	}
	return &DockerUpdater{
		cs:   o,
		opts: opts,
		dCli: dCli,
	}, nil
}

func (du *DockerUpdater) Update(ctx context.Context) error {
	if err := du.ValidateOpts(); err != nil {
		return err
	}

	// Remove invalid and empty environment variables
	du.cleanEnvs()

	// Get compose files
	composeFiles, err := du.getComposeFiles(du.opts.ComposeFiles, du.opts.WorkingDir, du.opts.TargetVersion)
	if err != nil {
		return err
	}

	// Start the deployment
	err = du.cs.Start(ctx, &types.StartOpts{
		CFiles:      composeFiles,
		Env:         du.opts.Environment,
		ProjectName: du.opts.Project,
		WorkingDir:  du.opts.WorkingDir,
	})

	if err != nil {
		return err
	}

	err = du.checkHealth(ctx, 10)
	if err == nil {
		log.Info("NuvlaEdge is healthy")
		return nil
	}

	log.Warn("NuvlaEdge is not healthy, rolling back")

	return nil
}

func (du *DockerUpdater) ValidateOpts() error {
	if du.opts == nil {
		return fmt.Errorf("update options are nil")
	}

	if du.opts.Project == "" {
		return fmt.Errorf("project name is required")
	}

	if du.opts.ComposeFiles == nil {
		log.Warn("No compose files provided. Using default docker-compose.yml")
		du.opts.ComposeFiles = []string{"docker-compose.yml"}
	}

	if du.opts.CurrentVersion == "" {
		return fmt.Errorf("current version is required to allow for rolling back and recovery")
	}

	if du.opts.TargetVersion == "" {
		if !du.opts.Force {
			return fmt.Errorf("target version is required. Use --force to update without a target version to the latest available")
		}
	}
	return nil
}

func (du *DockerUpdater) checkHealth(ctx context.Context, period int) error {

	ticker := time.NewTicker(time.Duration(period) * time.Second)
	defer ticker.Stop()
	initialRestartCount := -1
	for {
		select {
		case <-ctx.Done():
			log.Info("Health check finished, NuvlaEDge is healthy")
			return nil
		case <-ticker.C:
			// Check health
			h, err := du.monitorNuvlaEdge(ctx)
			if err != nil {
				log.Warn("Error monitoring NuvlaEdge: ", err)
			}
			if initialRestartCount == -1 {
				initialRestartCount = h.RestartCount
			}
			if h.RestartCount > initialRestartCount {
				log.Warn("NuvlaEdge restarted, rolling back")
				return errors.New("NuvlaEdge restarted, not healthy")
			}
			if !h.Running {
				log.Warn("NuvlaEdge stopped, rolling back")
				return errors.New("NuvlaEdge stopped, not healthy")
			}
			if h.Status != "running" {
				log.Warn("NuvlaEdge status is not running, rolling back")
				return errors.New("NuvlaEdge status is not running, not healthy")
			}
		}
	}
}

func (du *DockerUpdater) monitorNuvlaEdge(ctx context.Context) (NuvlaEdgeHealth, error) {
	var neHealth NuvlaEdgeHealth
	containers, err := du.cs.GetProjectStatus(ctx, du.opts.Project)
	if err != nil {
		return neHealth, err
	}
	var agent api.ContainerSummary
	// Monitor containers
	for _, c := range containers {
		if du.isAgent(c) {
			agent = c
			log.Info("found agent: ", agent.ID)

			break
		}
	}

	// Monitor the agent
	data, err := du.dCli.ContainerInspect(ctx, agent.ID)
	if err != nil {
		return neHealth, err
	}

	neHealth.RestartCount = data.RestartCount
	neHealth.Status = data.State.Status
	neHealth.Running = data.State.Running

	return neHealth, nil
}

func (du *DockerUpdater) isAgent(container api.ContainerSummary) bool {
	return container.Labels["com.docker.compose.service"] == "agent"
}

func (du *DockerUpdater) cleanEnvs() {
	var newEnv []string
	for _, e := range du.opts.Environment {
		if strings.Contains(e, "=") {
			newEnv = append(newEnv, e)
		} else {
			log.Warn("Invalid environment variable: ", e)
		}
	}
	du.opts.Environment = newEnv
}

func (du *DockerUpdater) getComposeFiles(reqFiles []string, workDir string, version string) ([]string, error) {
	nuvlaReleases, err := release.GetNuvlaRelease(version)
	if err == nil {
		files, err := nuvlaReleases.GetComposeFiles(reqFiles, workDir)
		if err == nil {
			return files, nil
		}
		log.Warn("Error getting compose files from Nuvla release: ", err)
	}

	log.Info("No Nuvla release found, trying GitHub release")
	ghReleases, err := release.GetGitHubRelease(version)
	if err != nil {
		log.Errorf("Error getting GitHub release: %s", err)
		return nil, err
	}

	files, err := ghReleases.GetComposeFiles(reqFiles, workDir)
	if err != nil {
		log.Errorf("Error getting compose files from GitHub release: %s", err)
		return nil, err
	}

	return files, nil
}

// findConfigInEnvironment looks for the configuration in the environment. Particularly the image to update to.
func (du *DockerUpdater) findConfigInEnvironment() string {
	keys := []string{"NE_IMAGE_REGISTRY", "NE_IMAGE_ORGANIZATION", "NE_IMAGE_GO_REPOSITORY", "NE_IMAGE_TAG"}

	envs := make(map[string]string)
	for _, env := range du.opts.Environment {
		if strings.ContainsAny(env, "=") {
			kv := strings.Split(env, "=")
			envs[kv[0]] = kv[1]
		}
	}

	var image imageName
	for _, key := range keys {
		if val, ok := envs[key]; ok {
			switch key {
			case "NE_IMAGE_REGISTRY":
				image.registry = val
			case "NE_IMAGE_ORGANIZATION":
				image.organization = val
			case "NE_IMAGE_GO_REPOSITORY":
				image.repository = val
			case "NE_IMAGE_TAG":
				image.tag = val
			}
		}
	}
	return image.String()
}

type imageName struct {
	registry     string
	organization string
	repository   string
	tag          string
}

func (i *imageName) String() string {
	return fmt.Sprintf("%s/%s/%s:%s", i.registry, i.organization, i.repository, i.tag)
}

type NuvlaEdgeHealth struct {
	RestartCount int
	Running      bool
	Status       string
}
