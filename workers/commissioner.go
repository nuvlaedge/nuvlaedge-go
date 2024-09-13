package workers

import (
	"context"
	"errors"
	"github.com/nuvla/api-client-go/clients"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/common"
	"nuvlaedge-go/types"
	"nuvlaedge-go/types/worker"
)

type Commissioner struct {
	worker.TimedWorker

	lastCommission types.CommissionAttributes
	currentData    types.CommissionAttributes

	commissionChan chan types.CommissionData

	client types.CommissionClientInterface
}

var DefaultTags = []string{"go", "nuvlaedge"}
var DefaultCapabilities = []string{"NUVLA_JOB_PULL", "NUVLA_HEARTBEAT"}

func (c *Commissioner) Init(opts *worker.WorkerOpts, conf *worker.WorkerConfig) error {
	c.TimedWorker = worker.NewTimedWorker(conf.CommissionPeriod, worker.Commissioner)
	c.client = &types.CommissionClient{NuvlaEdgeClient: opts.NuvlaClient}
	c.commissionChan = opts.CommissionCh
	c.newDefaultData()
	return nil
}

func (c *Commissioner) newDefaultData() {
	c.currentData.Tags = DefaultTags
	c.currentData.Capabilities = DefaultCapabilities
}

func (c *Commissioner) getNodeIdFromStatus() string {
	if c.client.GetStatusId() == "" {
		log.Infof("NuvlaEdge status not available, cannot get node id")
		return ""
	}

	resource, err := c.client.Get(c.client.GetStatusId(), []string{"node-id"})
	if err != nil {
		log.Errorf("Error getting NuvlaEdge status: %s", err)
		return ""
	}

	node, ok := resource.Data["node-id"]
	if ok {
		nodeStr, ok := node.(string)
		if ok {
			return nodeStr
		}

		return ""
	}
	return ""
}

func (c *Commissioner) Start(ctx context.Context) error {
	log.Info("Starting Commissioner")

	if c.client.GetStatusId() == "" {
		log.Warnf("NuvlaEdge status not available, cannot start Commissioner")
		return nil
	}

	go c.Run(ctx)
	return nil
}

func (c *Commissioner) Run(ctx context.Context) error {
	log.Info("Starting Commissioner")

	for {
		select {
		case <-ctx.Done():
			log.Info("Commissioner stopped")
			if err := c.Stop(ctx); err != nil {
				log.Errorf("Failed to stop Commissioner: %s", err)
				return err
			}
			return ctx.Err()

		case d := <-c.commissionChan:
			if err := d.WriteToAttrs(&c.currentData); err != nil {
				log.Errorf("Error writing commission data to attributes: %s", err)
			}

		case <-c.BaseTicker.C:
			if data, ok := c.needsCommissioning(); ok {
				if err := c.commission(data); err != nil {
					log.Errorf("Error commissioning: %s", err)
				} else {
					c.lastCommission = c.currentData
				}
			}

		case conf := <-c.ConfChan:
			log.Info("Received configuration in commissioner: ", conf)
			if err := c.Reconfigure(conf); err != nil {
				log.Errorf("Error reconfiguring Commissioner: %s", err)
			}
		}
	}
}

func (c *Commissioner) Reconfigure(conf *worker.WorkerConfig) error {
	log.Infof("Reconfiguring Commissioner")
	if conf.CommissionPeriod != c.GetPeriod() {
		c.SetPeriod(conf.CommissionPeriod)
	}
	return nil
}

func (c *Commissioner) Stop(_ context.Context) error {
	log.Info("Stopping Commissioner")
	c.BaseTicker.Stop()
	return nil
}

func (c *Commissioner) needsCommissioning() (map[string]interface{}, bool) {
	data, del := common.GetStructDiff(c.lastCommission, c.currentData)
	if len(del) > 0 {
		data["removed"] = del
	}
	if len(data) == 0 {
		log.Infof("No new data to commission")
		return nil, false
	}
	return data, true
}

func (c *Commissioner) commission(data map[string]interface{}) error {
	log.Infof("Commissioning with data: %v", data)
	if err := c.client.Commission(data); err != nil {
		log.Errorf("Error commissioning: %s", err)
		return err
	}
	log.Infof("Commissioned successfully")
	return nil
}

// TriggerBaseCommissioning triggers the base commissioning of the node manually
func TriggerBaseCommissioning(w worker.Worker, nuvla *clients.NuvlaEdgeClient) error {
	c, ok := w.(*Commissioner)
	if !ok {
		return errors.New("worker is not a Commissioner")
	}

	comData, ok := c.needsCommissioning()
	if !ok {
		log.Errorf("This is the first commissioning and data should be available, something went wrong")
	}

	return nuvla.Commission(comData)
}

var _ worker.Worker = &Commissioner{}
