package workers

import (
	"context"
	"github.com/nuvla/api-client-go/clients"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/common"
	"nuvlaedge-go/types/worker"
)

type Heartbeat struct {
	worker.TimedWorker

	client         *clients.NuvlaEdgeClient
	jobChan        chan string
	confUpdateChan chan string
}

func (h *Heartbeat) Init(opts *worker.WorkerOpts, conf *worker.WorkerConfig) error {
	h.TimedWorker = worker.NewTimedWorker(conf.HeartBeatPeriod, worker.Heartbeat)

	h.client = opts.NuvlaClient
	h.jobChan = opts.JobCh
	h.confUpdateChan = opts.ConfLastUpdateCh

	return nil
}

func (h *Heartbeat) Start(ctx context.Context) error {
	// Maybe check here is NuvlaEdge status is commissioned, else wait
	go h.Run(ctx)
	return nil
}

func (h *Heartbeat) Run(ctx context.Context) error {
	log.Info()
	for {
		select {
		case <-ctx.Done():
			if err := h.Stop(ctx); err != nil {
				log.Error("Failed to stop heartbeat worker: ", err)
				return err
			}
			return ctx.Err()
		case <-h.BaseTicker.C:
			log.Info("Sending heartbeat")
			if err := h.sendHeartbeat(); err != nil {
				log.Error("Failed to send heartbeat: ", err)
			}
		case conf := <-h.ConfChan:
			log.Info("Received configuration in heartbeat: ", conf)
			err := h.Reconfigure(conf)
			if err != nil {
				log.Error("Failed to reconfigure heartbeat worker: ", err)
			}
		}
	}
}

func (h *Heartbeat) Reconfigure(conf *worker.WorkerConfig) error {
	log.Info("Reconfiguring heartbeat worker")
	h.SetPeriod(conf.HeartBeatPeriod)
	return nil
}

func (h *Heartbeat) sendHeartbeat() error {
	res, err := h.client.Heartbeat()
	if err != nil {
		log.Error("Failed to send heartbeat: ", err)
		return err
	}

	if err := common.ProcessResponse(res, h.jobChan, h.confUpdateChan); err != nil {
		log.Error("Failed to process heartbeat response: ", err)
		return err
	}

	return nil
}

func (h *Heartbeat) Stop(ctx context.Context) error {
	h.BaseTicker.Stop()
	return nil
}

var _ worker.Worker = &Heartbeat{}
