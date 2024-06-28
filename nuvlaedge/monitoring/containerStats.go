package monitoring

import (
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/nuvlaedge/orchestrator"
	"time"
)

const StatsChannelSize int = 5

type ContainerStats struct {
	Coe             orchestrator.Coe
	refreshInterval int // in seconds
	stats           chan []map[string]any
	updateTime      time.Time
}

func NewContainerStats(coe *orchestrator.Coe, refreshInterval int) *ContainerStats {
	return &ContainerStats{
		Coe:             *coe,
		refreshInterval: refreshInterval,
		updateTime:      time.Now(),
		stats:           make(chan []map[string]any, StatsChannelSize),
	}
}

func (cs *ContainerStats) getStats() ([]map[string]any, error) {
	if time.Since(cs.updateTime) < 10*time.Second {
		log.Debugf("Getting the CS stats ")
		if len(cs.stats) == 0 {
			cs.stats <- cs.getContainerStats()
		}
	} else {
		if len(cs.stats) < StatsChannelSize {
			cs.stats <- cs.getContainerStats()
		}
	}
	log.Debugf("Got the CS stats ")
	return <-cs.stats, nil
}

func (cs *ContainerStats) getContainerStats() []map[string]any {
	containers, err := cs.Coe.GetContainers()
	if err != nil {
		log.Errorf("Got Error while getting containers %s", err)
		return nil
	}

	var containerStats []map[string]any
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
	}
	return containerStats
}
