package monitoring

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"io"
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
		log.Infof("Container Stats need to be updated")
		cs.stats = []map[string]any{}
		err := cs.getContainerStats()
		cs.updateTime = time.Now()
		if err != nil {
			return nil, err
		}
	}
	log.Debugf("Container Stats Collected")
	return cs.stats, nil
}

func (cs *ContainerStats) getContainerStats() error {
	log.Debugf("Getting Container Stats")
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
			if errors.Is(err, io.EOF) {
				// This could happen if the container is stopped while we are reading the stats
				// or due to network issues.
				log.Errorf("EOF encountered while reading container stats")
				continue
			}
			return nil
		}
		cs.stats = append(cs.stats, containerInfo)
	}
	return nil
}