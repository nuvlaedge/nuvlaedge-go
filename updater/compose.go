package updater

import (
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/orchestrator"
	"nuvlaedge-go/types"
	"nuvlaedge-go/types/options/command"
	"nuvlaedge-go/updater/release"
	"os"
	"strings"
	"time"
)

func parseEnv(env []string) []string {
	var cleanEnvs []string

	for _, e := range env {
		if strings.Contains(e, "=") {
			cleanEnvs = append(cleanEnvs, e)
		} else {
			log.Warn("Invalid environment variable: ", e)
		}
	}

	return cleanEnvs
}

func UpdateWithCompose(opts *command.UpdateCmdOptions) error {
	// Assert compose orchestrator
	compose, err := orchestrator.NewComposeOrchestrator(nil)
	if err != nil {
		return err
	}

	if opts.Project == "" {
		return errors.New("project name is required")
	}

	err = os.Setenv("COMPOSE_PROJECT_NAME", opts.Project)
	if err != nil {
		return err
	}

	// Parse environment variables
	env := parseEnv(opts.Environment)

	// Download compose files composing GitHubRelease with compose files
	composeFiles, err := getComposeFiles(opts.ComposeFiles, opts.WorkingDir, opts.TargetVersion)
	if err != nil {
		return err
	}

	startTimeout := 120

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(startTimeout)*time.Second)
	defer cancel()

	err = compose.Start(ctx, &types.StartOpts{
		CFiles:      composeFiles,
		Env:         env,
		ProjectName: opts.Project,
		WorkingDir:  opts.WorkingDir,
	})
	if err != nil {
		return err
	}

	return nil
}

func getComposeFiles(composeFiles []string, workDir string, version string) ([]string, error) {
	nuvlaReleases, err := release.GetNuvlaRelease(version)
	if err == nil {
		files, err := nuvlaReleases.GetComposeFiles(composeFiles, workDir)
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

	files, err := ghReleases.GetComposeFiles(composeFiles, workDir)
	if err != nil {
		log.Errorf("Error getting compose files from GitHub release: %s", err)

		return nil, err
	}

	return files, nil
}
