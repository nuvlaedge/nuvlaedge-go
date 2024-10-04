package workers

import (
	"context"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/common"
	"nuvlaedge-go/types"
	"nuvlaedge-go/types/worker"
	"time"
)

type Heartbeat struct {
	worker.TimedWorker

	client         types.HeartbeatClient
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
	go func() {
		err := h.Run(ctx)
		if err != nil {
			log.Errorf("Error running Commissioner: %s", err)
		}
	}()
	return nil
}

func (h *Heartbeat) Run(ctx context.Context) error {
	log.Debug("Running heartbeat...")
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
			if err := h.sendHeartbeat(ctx); err != nil {
				log.Error("Failed to send heartbeat: ", err)
			}

		case conf := <-h.ConfChan:
			log.Debug("Received configuration in heartbeat: ", conf)
			err := h.Reconfigure(conf)
			if err != nil {
				log.Error("Failed to reconfigure heartbeat worker: ", err)
			}
		}
	}
}

func (h *Heartbeat) Reconfigure(conf *worker.WorkerConfig) error {
	h.SetPeriod(conf.HeartBeatPeriod)
	return nil
}

func (h *Heartbeat) sendHeartbeat(ctx context.Context) error {
	ctxTimed, cancel := context.WithTimeout(ctx, time.Duration(h.GetPeriod())*time.Second)
	defer cancel()

	res, err := h.client.Heartbeat(ctxTimed)
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
