package monitoring

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"io"
	"nuvlaedge-go/nuvlaedge/common/resources"
	"nuvlaedge-go/nuvlaedge/orchestrator"
	"time"
)

type ContainerStats struct {
	Coe             orchestrator.Coe
	refreshInterval int // in seconds
	stats           []interface{}
	updateTime      time.Time
	oldVersion      bool
}

func NewContainerStats(coe *orchestrator.Coe, refreshInterval int, old bool) *ContainerStats {
	return &ContainerStats{
		Coe:             *coe,
		refreshInterval: refreshInterval,
		updateTime:      time.Now(),
		stats:           nil,
		oldVersion:      old,
	}
}

func (cs *ContainerStats) getStats() ([]interface{}, error) {
	if time.Since(cs.updateTime) > 10*time.Second || cs.stats == nil {
		log.Infof("Container Stats need to be updated")
		cs.stats = []interface{}{}
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
	containers, err := cs.Coe.GetContainers(cs.oldVersion)
	if err != nil {
		log.Errorf("Got Error while getting containers %s", err)
		return err
	}

	for _, containerInfo := range containers {
		if containerInfo == nil {
			log.Errorf("Error getting container information")
			continue
		}
		log.Debugf("Getting stats for container: %v", containerInfo)
		containerId, err := GetContainerId(&containerInfo)
		if err != nil {
			log.Errorf("Error getting container id: %s", err)
			continue
		}
		err = cs.Coe.GetContainerStats(containerId, &containerInfo)
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
		log.Debugf("Container Stats Got for GetContainerStats: %v", containerInfo)
		cs.stats = append(cs.stats, containerInfo)
	}
	return nil
}

func GetContainerId(containerInfo *interface{}) (string, error) {
	switch info := (*containerInfo).(type) {
	case resources.ContainerStatsOld:
		return info.ContainerId, nil
	case resources.ContainerStatsNew:
		return info.ContainerId, nil
	default:
		return "", errors.New("Unknown container stats type")
	}
}
