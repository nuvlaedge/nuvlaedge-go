package agent

import (
	"encoding/json"
	nuvla "github.com/nuvla/api-client-go/clients"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/nuvlaedge/common"
	"nuvlaedge-go/nuvlaedge/orchestrator"
	"reflect"

	"nuvlaedge-go/nuvlaedge/types"
	"time"
)

type Commissioner struct {
	nuvlaClient *nuvla.NuvlaEdgeClient
	coeClient   orchestrator.Coe

	lastPayload *types.CommissioningAttributes
	currentData *types.CommissioningAttributes
}

func NewCommissioner(nuvlaClient *nuvla.NuvlaEdgeClient, coeClient orchestrator.Coe) *Commissioner {
	return &Commissioner{
		nuvlaClient: nuvlaClient,
		coeClient:   coeClient,
	}
}

func (c *Commissioner) commission() error {
	var mapData map[string]interface{}
	d, err := json.Marshal(c.currentData)
	if err != nil {
		log.Errorf("Error marshaling commissioning data: %s", err)
		return err
	}
	if err = json.Unmarshal(d, &mapData); err != nil {
		log.Errorf("Error generating map commissioning data: %s", err)
		return err
	}
	common.CleanMap(mapData)
	if err = c.nuvlaClient.Commission(mapData); err != nil {
		log.Errorf("Error commissioning: %s", err)
		return err
	}
	b, _ := json.MarshalIndent(mapData, "", "  ")
	log.Infof("Commissioning successful with data %s", string(b))
	return nil
}

func (c *Commissioner) getClusterIdFromStatus() string {
	if c.nuvlaClient.NuvlaEdgeStatusId == nil {
		log.Infof("NuvlaEdge status not available, cannot get cluster id")
		return ""
	}
	resource, err := c.nuvlaClient.Get(c.nuvlaClient.NuvlaEdgeStatusId.String(), nil)
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

func (c *Commissioner) diffCommissioningAttributes() bool {
	return true
}

func (c *Commissioner) needsCommission() bool {
	if c.lastPayload == nil {
		return true
	}
	return !reflect.DeepEqual(c.currentData, c.lastPayload)
}

func (c *Commissioner) Run() {
	log.Infof("Commissioner running...")
	for {
		startTime := time.Now()
		c.updateData()

		if c.needsCommission() {
			log.Infof("Commissioning %s", c.currentData)
			if err := c.commission(); err != nil {
				log.Errorf("Error commissining with data %v: %s", c.currentData, err)
			} else {
				copied := *c.currentData
				c.lastPayload = &copied
			}
		}

		err := common.WaitPeriodicAction(startTime, 60, "Commissioner")
		if err != nil {
			log.Errorf("Error waiting for commissioner: %s", err)
		}
	}
}
