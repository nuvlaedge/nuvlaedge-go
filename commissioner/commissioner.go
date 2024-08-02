package commissioner

import (
	"context"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/common"
	"nuvlaedge-go/types"
	"time"
)

type Commissioner struct {
	types.WorkerBase

	lastCommission types.CommissionAttributes
	currentData    types.CommissionAttributes

	commissionChan chan types.CommissionData

	client types.CommissionClientInterface
}

var DefaultTags = []string{"go", "nuvlaedge"}
var DefaultCapabilities = []string{"NUVLA_JOB_PULL", "NUVLA_HEARTBEAT"}

func NewCommissioner(
	period int,
	client types.CommissionClientInterface,
	ch chan types.CommissionData) *Commissioner {

	// period needs to be aligned with the monitoring period
	c := &Commissioner{
		WorkerBase:     types.NewWorkerBase(period, types.Commissioner),
		client:         client,
		commissionChan: ch,
	}

	c.newDefaultData()
	return c
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
		return node.(string)
	}
	return ""
}

func (c *Commissioner) Run(ctx context.Context) error {
	log.Info("Starting Commissioner")
	ticker := time.NewTicker(time.Duration(c.Period) * time.Second)
	defer c.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("Commissioner stopped")
			return ctx.Err()
		case d := <-c.commissionChan:
			if err := d.WriteToAttrs(&c.currentData); err != nil {
				log.Errorf("Error writing commission data to attributes: %s", err)
			}
		case <-ticker.C:
			if data, ok := c.needsCommissioning(); ok {
				if err := c.commission(data); err != nil {
					log.Errorf("Error commissioning: %s", err)
				} else {
					c.lastCommission = c.currentData
				}
			}
		}
	}
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

func (c *Commissioner) GetType() types.WorkerType {
	return types.Commissioner
}
