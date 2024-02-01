package commissioner

import (
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/src/coe"
	"nuvlaedge-go/src/common"
	"nuvlaedge-go/src/nuvlaClient"
	"time"
)

type Commissioner struct {
	nuvlaClient *nuvlaClient.NuvlaEdgeClient
	coeClient   coe.Coe

	lastPayload *CommissioningAttributes
	currentData *CommissioningAttributes
}

func NewCommissioner(nuvlaClient *nuvlaClient.NuvlaEdgeClient, coeClient coe.Coe) *Commissioner {
	return &Commissioner{
		nuvlaClient: nuvlaClient,
		coeClient:   coeClient,
	}
}

func (c *Commissioner) commission() {
	commission, err := c.nuvlaClient.Commission(c.currentData)
	if err != nil {
		return
	}

	if commission {
		*c.lastPayload = *c.currentData
	}
}

func (c *Commissioner) updateData() {
	// Updating cluster data from COE
	clusterData, err := c.coeClient.GetClusterData()
	if err != nil {
		log.Errorf("Error retrieving cluster data: %s", err)
		return
	}
	c.currentData.ClusterID = clusterData.ClusterId
	c.currentData.ClusterManagers = clusterData.ClusterManagers
	c.currentData.ClusterWorkers = clusterData.ClusterWorkers
	c.currentData.ClusterOrchestrator = string(c.coeClient.GetCoeType())

	// Update Orchestrator data

}

func (c *Commissioner) diffCommissioningAttributes() bool {
	return true
}

func (c *Commissioner) needsCommission() bool {

	return true
}

func (c *Commissioner) Run() {
	for {
		startTime := time.Now()
		log.Infof("Commissioner started at %s", startTime)
		c.updateData()

		if c.needsCommission() {
			log.Infof("Commissioning %s", c.currentData)
			c.commission()
		}

		err := common.WaitPeriodicAction(startTime, 10, "Commissioner")
		if err != nil {
			log.Errorf("Error waiting for commissioner: %s", err)
		}
	}
}
