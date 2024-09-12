package actions

import (
	"context"
	"encoding/json"
	"errors"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/common"
	"nuvlaedge-go/job_engine/connector"
	"nuvlaedge-go/types/jobs"
	"nuvlaedge-go/types/options/command"
	"strings"
)

type UpdateAction struct {
	job *jobs.JobBase
}

func (u *UpdateAction) Update(ctx context.Context, targetRelease string) error {
	///TODO implement me
	switch common.WhereAmI() {
	case common.Host:
		return u.updateHost(ctx, targetRelease)
	case common.Docker:
		log.Info("Updating with docker image")
		return u.updateDocker(ctx, targetRelease)
	case common.Kubernetes:
		return u.updateHelm(ctx, targetRelease)
	default:
		log.Error("Unknown environment")
		return errors.New("cannot update in unknown environment")
	}
}

func (u *UpdateAction) updateHost(ctx context.Context, targetRelease string) error {
	//TODO implement me
	panic("implement")
}

func (u *UpdateAction) updateDocker(ctx context.Context, targetRelease string) error {
	u.job.Client.SetInitialState()

	dockerCon, err := connector.NewDockerConnector()

	params, err := u.getUpdateParams()
	if err != nil {
		u.job.Client.SetFailedState(err.Error())
		return err
	}

	if err = dockerCon.UpdateNuvlaEdge(ctx, "local/nuvlaedge:refactor", params); err != nil {
		u.job.Client.SetFailedState(err.Error())
		return err
	}

	u.job.Client.SetSuccessState()
	return nil
}

func (u *UpdateAction) updateHelm(ctx context.Context, targetRelease string) error {
	//TODO implement me
	panic("implement")
}

type UpdatePayload struct {
	ProjectName    string   `json:"project-name"`
	WorkingDir     string   `json:"working-dir"`
	Environment    []string `json:"environment"`
	ForceRestart   bool     `json:"force-restart"`
	CurrentVersion string   `json:"current-version"`
	ConfigFiles    []string `json:"config-files"`
}

func (u *UpdateAction) getUpdateParams() (*command.UpdateCmdOptions, error) {
	var updateCmdOptions command.UpdateCmdOptions
	var updatePayload UpdatePayload

	if err := json.Unmarshal([]byte(u.job.Resource.Payload), &updatePayload); err != nil {
		return nil, err
	}

	updateCmdOptions.Project = updatePayload.ProjectName
	updateCmdOptions.WorkingDir = updatePayload.WorkingDir
	updateCmdOptions.Environment = updatePayload.Environment

	b, _ := json.MarshalIndent(updatePayload, "", "  ")
	log.Infof("Update payload: %s", string(b))

	return &updateCmdOptions, nil
}

func (u *UpdateAction) Execute(ctx context.Context) error {
	affectedRes := u.job.Resource.AffectedResources

	var releaseId string
	for _, res := range affectedRes {
		if strings.HasPrefix(res.Href, "nuvlabox-release/") {
			releaseId = res.Href
			break
		}
	}
	if releaseId == "" {
		return errors.New("no nuvlabox-release resource found")
	}

	return u.Update(ctx, releaseId)
}

func (u *UpdateAction) Init(opts *ActionOpts) error {
	u.job = opts.Job
	return nil
}
