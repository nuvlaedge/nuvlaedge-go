package nuvlaedge

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/nuvla/api-client-go/clients"
	"github.com/nuvla/api-client-go/clients/resources"
	types2 "github.com/nuvla/api-client-go/types"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/common"
	"nuvlaedge-go/common/constants"
	"nuvlaedge-go/types"
	"nuvlaedge-go/types/jobs"
	"nuvlaedge-go/types/settings"
	"nuvlaedge-go/types/worker"
	"nuvlaedge-go/workers"
	"path"
)

type Workers map[worker.WorkerType]worker.Worker

type NuvlaEdge struct {
	ctx  context.Context             // Parent context
	conf *settings.NuvlaEdgeSettings // NuvlaEdge settings

	// Channels
	commissionerCh   chan types.CommissionData // Connects Telemetry/EngineMonitor with Commissioner
	jobCh            chan string               // Connects Agent and Telemetry with Job Processor
	deploymentCh     chan jobs.Job             // Connects Job Processor with Deployment handler
	confLastUpdateCh chan string               // Connects Heartbeat and Telemetry responses with Configuration handler

	nuvla        *clients.NuvlaEdgeClient
	dockerClient client.APIClient

	workerOpts *worker.WorkerOpts
	workerConf *worker.WorkerConfig

	workers Workers
}

func NewNuvlaEdge(ctx context.Context, conf *settings.NuvlaEdgeSettings) (*NuvlaEdge, error) {

	nuvla, err := ValidateSettings(conf)
	if err != nil {
		return nil, err
	}
	b, _ := json.MarshalIndent(conf, "", "  ")
	log.Infof("Starting NuvlaEdge with settings: %s", string(b))

	// To add K8s, this will need to be converted into an interface
	dockerCli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	wConf := worker.NewDefaultWorkersConfig()
	wConf.EnableJobLegacy = conf.EnableJobLegacySupport
	wConf.LegacyJobImage = conf.JobEngineImage
	wConf.CleanUpPeriod = conf.CleanUpPeriod
	wConf.RemoveObjects = conf.Resources

	ne := &NuvlaEdge{
		ctx:          ctx,
		nuvla:        nuvla,
		dockerClient: dockerCli,
		conf:         conf,
		workerConf:   wConf,

		// Channels
		commissionerCh:   make(chan types.CommissionData),
		jobCh:            make(chan string),
		deploymentCh:     make(chan jobs.Job),
		confLastUpdateCh: make(chan string),
	}

	jobRegistry := jobs.NewRunningJobs()
	ne.workerOpts = &worker.WorkerOpts{
		NuvlaClient:      nuvla,
		DockerClient:     dockerCli,
		CommissionCh:     ne.commissionerCh,
		JobCh:            ne.jobCh,
		DeploymentCh:     ne.deploymentCh,
		ConfLastUpdateCh: ne.confLastUpdateCh,
		Jobs:             &jobRegistry,
	}

	ne.workers, err = WorkerGenerator(ne.workerOpts, ne.workerConf)
	if err != nil {
		return nil, err
	}

	return ne, nil
}

func (ne *NuvlaEdge) Start(ctx context.Context) error {

	// NuvlaEdge startup process...

	if err := ne.startUpProcess(ctx); err != nil {
		return err
	}

	if err := ne.startWorkers(); err != nil {
		return err
	}

	return nil
}

func (ne *NuvlaEdge) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (ne *NuvlaEdge) startWorkers() error {
	var errList []error
	for _, w := range ne.workers {
		if err := w.Start(ne.ctx); err != nil {
			errList = append(errList, err)
		}
	}
	return errors.Join(errList...)
}

func (ne *NuvlaEdge) startUpProcess(ctx context.Context) error {
	// Start up process
	// Get remote nuvlaedge state

	if (ne.conf.ApiKey == "" || ne.conf.ApiSecret == "") && ne.conf.Irs == "" {

		// We need to assume that NuvlaEdge is new
		creds, err := ne.nuvla.Activate(ctx)
		if err != nil {
			return err
		}

		irs, err := common.GetIrsV2(creds, ne.nuvla.NuvlaEdgeId.String())
		if err != nil {
			return err
		}

		ne.conf.Irs = irs
		ne.nuvla.Irs = irs

		err = ne.nuvla.Freeze(path.Join(ne.conf.DBPPath, constants.NuvlaEdgeSessionFile))
		if err != nil {
			return err
		}
	}

	var c types2.ApiKeyLogInParams
	var err error

	if ne.conf.Irs != "" {
		ne.nuvla.Irs = ne.conf.Irs
		c, err = common.FromIrsV2(ne.conf.Irs, ne.nuvla.NuvlaEdgeId.String())

		if err != nil {
			return fmt.Errorf("error decrypting credentials: %s", err)
		}

	} else {
		c = *ne.nuvla.Credentials.(*types2.ApiKeyLogInParams)
	}

	if err := ne.nuvla.LogIn(c); err != nil {
		return err
	}

	ctxCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	// else, check state
	err = ne.nuvla.UpdateResourceSelect(ctxCancel, []string{"state"})
	if err != nil {
		return err
	}

	res := ne.nuvla.GetNuvlaEdgeResource()

	if res.State == resources.NuvlaEdgeStateActivated {
		// Trigger commission once
		err := workers.TriggerBaseCommissioning(ctxCancel, ne.workers[worker.Commissioner], ne.nuvla)
		if err != nil {
			return err
		}
	}

	// else, check state
	err = ne.nuvla.UpdateResourceSelect(ctxCancel, []string{"state"})
	if err != nil {
		return err
	}

	res = ne.nuvla.GetNuvlaEdgeResource()

	if res.State != resources.NuvlaEdgeStateCommissioned {
		return fmt.Errorf("can't start a NuvlaEDge from state: %s", res.State)
	}

	log.Info("Start Up process completed, NuvlaEdge is ready")
	return nil
}
