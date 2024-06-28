package monitoring

import (
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/nuvlaedge/orchestrator"
	"time"
)

type ContainerStats struct {
	Coe             orchestrator.Coe
	refreshInterval int // in seconds
	stats           []map[string]any
	updateTime      time.Time
}

func NewContainerStats(coe *orchestrator.Coe, refreshInterval int) *ContainerStats {
	return &ContainerStats{
		Coe:             *coe,
		refreshInterval: refreshInterval,
		updateTime:      time.Now(),
		stats:           nil,
	}
}

func (cs *ContainerStats) getStats() ([]map[string]any, error) {
	if time.Since(cs.updateTime) > 10*time.Second || cs.stats == nil {
		cs.stats = []map[string]any{}
		err := cs.getContainerStats()
		if err != nil {
			return nil, err
		}
	}
	log.Debugf("Container Stats Collected")
	return cs.stats, nil
}

func (cs *ContainerStats) getContainerStats() error {
	containers, err := cs.Coe.GetContainers()
	if err != nil {
		log.Errorf("Got Error while getting containers %s", err)
		return err
	}

	for _, containerInfo := range containers {
		id, ok := containerInfo["id"].(string)
		if !ok {
			log.Errorf("Error getting container id")
			continue
		}
		err := cs.Coe.GetContainerStats(id, &containerInfo)
		if err != nil {
			log.Errorf("Error getting container stats: %s", err)
			return nil
		}

		cs.stats = append(cs.stats, containerInfo)
	}
	return nil
}
