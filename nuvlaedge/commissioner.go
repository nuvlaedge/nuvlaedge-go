package nuvlaedge

import (
	"context"
	nuvla "github.com/nuvla/api-client-go/clients"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/nuvlaedge/common"
	"nuvlaedge-go/nuvlaedge/orchestrator"
	"nuvlaedge-go/nuvlaedge/types"
	"time"
)

type Commissioner struct {
	ctx         context.Context
	nuvlaClient *nuvla.NuvlaEdgeClient
	coeClient   orchestrator.Coe

	lastPayload *types.CommissioningAttributes
	currentData *types.CommissioningAttributes
}

func NewCommissioner(ctx context.Context, nuvlaClient *nuvla.NuvlaEdgeClient, coeClient orchestrator.Coe) *Commissioner {
	return &Commissioner{
		ctx:         ctx,
		nuvlaClient: nuvlaClient,
		coeClient:   coeClient,
	}
}

func (c *Commissioner) commission(commissionData map[string]interface{}) error {
	log.Infof("Commissioning with data: %v", commissionData)
	if err := c.nuvlaClient.Commission(commissionData); err != nil {
		log.Errorf("Error commissioning: %s", err)
		return err
	}
	log.Infof("Commissioned successfully")
	return nil
}

func (c *Commissioner) getClusterIdFromStatus() string {
	if c.nuvlaClient.NuvlaEdgeStatusId == nil {
		log.Infof("NuvlaEdge status not available, cannot get cluster id")
		return ""
	}
	resource, err := c.nuvlaClient.Get(c.nuvlaClient.NuvlaEdgeStatusId.String(), []string{"node-id"})

	if err != nil {
		log.Errorf("Error getting NuvlaEdge status: %s", err)
	}
	cluster, ok := resource.Data["node-id"]
	if ok {
		return cluster.(string)
	} else {
		return ""
	}
}
func (c *Commissioner) updateData() {
	if c.currentData == nil {
		c.currentData = &types.CommissioningAttributes{}
	}

	// Updating cluster data from COE
	clusterData, err := c.coeClient.GetClusterData()
	if err != nil {
		log.Errorf("Error retrieving cluster data: %s", err)
		return
	}

	if clusterData != nil && c.getClusterIdFromStatus() != "" {
		// FIXME: This needs improvement and refactor
		c.currentData.ClusterID = clusterData.ClusterId
		c.currentData.ClusterManagers = clusterData.ClusterManagers
		c.currentData.ClusterWorkers = clusterData.ClusterWorkers
		c.currentData.ClusterOrchestrator = string(c.coeClient.GetCoeType())
	} else {
		log.Infof("Cluster not available")
	}

	// Update Orchestrator data
	if err := c.coeClient.GetOrchestratorCredentials(c.currentData); err != nil {
		log.Errorf("Error retrieving orchestrator credentials: %s", err)
	}

	c.currentData.Tags = []string{"test", "go", "nuvlaedge"}
	c.currentData.Capabilities = []string{"NUVLA_JOB_PULL", "NUVLA_HEARTBEAT"}
}

func (c *Commissioner) needsCommission() (map[string]interface{}, bool) {
	if c.lastPayload == nil {
		c.lastPayload = &types.CommissioningAttributes{}
	}

	diff, del := common.GetStructDiff(*c.lastPayload, *c.currentData)
	if len(del) > 0 {
		diff["removed"] = del
	}
	if len(diff) == 0 {
		return nil, false
	}
	return diff, true
}

func (c *Commissioner) SingleIteration() {
	c.updateData()

	if data, ok := c.needsCommission(); ok {
		if err := c.commission(data); err != nil {
			log.Errorf("Error commissining with data %v: %s", c.currentData, err)
		} else {
			copied := *c.currentData
			c.lastPayload = &copied
		}
	} else {
		log.Infof("No need to commission")
	}
}

func (c *Commissioner) Run() {
	log.Infof("Commissioner started...")
	tick := time.NewTicker(60 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			c.SingleIteration()
		case <-c.ctx.Done():
			log.Infof("Exiting Commissioner...")
			return
		}
	}
}
