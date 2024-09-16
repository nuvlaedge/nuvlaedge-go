package workers

import (
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/types"
	"nuvlaedge-go/types/worker"
	"sync"
	"time"
)

type ConfUpdater struct {
	worker.WorkerBase
	client types.ConfUpdaterClient

	lastUpdate time.Time

	confChan       chan string
	config         *worker.WorkerConfig
	configChannels []chan *worker.WorkerConfig
}

func (c *ConfUpdater) Init(opts *worker.WorkerOpts, conf *worker.WorkerConfig) error {
	c.WorkerBase = worker.NewWorkerBase(worker.ConfUpdater)
	c.client = opts.NuvlaClient
	c.confChan = opts.ConfLastUpdateCh
	c.config = conf
	c.configChannels = opts.ConfigChannels

	return nil
}

func (c *ConfUpdater) Start(ctx context.Context) error {
	go func() {
		err := c.Run(ctx)
		if err != nil {
			log.Errorf("Error running Commissioner: %s", err)
		}
	}()
	return nil
}

func (c *ConfUpdater) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			err := c.Stop(ctx)
			if err != nil {
				log.Error("Failed to stop ConfUpdater: ", err)
				return err
			}
			return ctx.Err()
		case lastUpdate := <-c.confChan:
			if err := c.updateConfigIfNeeded(lastUpdate); err != nil {
				log.Error("Failed to update config: ", err)
			}
		case <-c.ConfChan:
			// We need to listen to configuration changes even if we don't use them to prevent the channel from blocking
		}
	}
}

func (c *ConfUpdater) Reconfigure(conf *worker.WorkerConfig) error {
	return nil
}

func (c *ConfUpdater) Stop(_ context.Context) error {
	log.Info("Stopping ConfUpdater")
	return nil
}

func (c *ConfUpdater) needsUpdate(remoteUpdateTime string) (bool, *time.Time) {

	remoteTime, err := time.Parse(time.RFC3339, remoteUpdateTime)
	if err != nil {
		log.Error("Failed to parse last update date: ", err)
		return false, nil
	}

	if remoteTime.After(c.lastUpdate) {
		return true, &remoteTime
	}

	return false, nil
}

func (c *ConfUpdater) updateConfigIfNeeded(lastUpdateDate string) error {
	// Check if new update is needed
	ok, remoteTime := c.needsUpdate(lastUpdateDate)
	if !ok {
		log.Debugf("Local configuration is up to date")
		return nil
	}

	err := c.client.UpdateResourceSelect([]string{"refresh-interval", "heartbeat-interval"})
	if err != nil {
		log.Error("Failed to update resource: ", err)
		return err
	}

	nuvlaEdgeRes := c.client.GetNuvlaEdgeResource()
	c.config.UpdateFromResource(&nuvlaEdgeRes)

	if err := c.distributeConfig(c.config); err != nil {
		log.Error("Failed to distribute new config")
		return err
	}

	c.lastUpdate = *remoteTime
	return nil
}

func (c *ConfUpdater) distributeConfig(conf *worker.WorkerConfig) error {
	var wg sync.WaitGroup
	log.Infof("Distributing new config to %d channels", len(c.configChannels))
	wg.Add(len(c.configChannels))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, confChan := range c.configChannels {
		go func(confChan chan *worker.WorkerConfig, wg *sync.WaitGroup) {
			confChan <- conf
			wg.Done()
		}(confChan, &wg)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Info("Config distributed")
	case <-ctx.Done():
		log.Error("Failed to distribute config: ", ctx.Err())
		return errors.New("failed to distribute config")
	}
	return nil
}

// Compile time check
var _ worker.Worker = &ConfUpdater{}
