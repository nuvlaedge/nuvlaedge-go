package monitoring

import (
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/nuvlaedge/common/resources"
	"nuvlaedge-go/nuvlaedge/orchestrator"
	"time"
)

type ContainerStats struct {
	Coe             orchestrator.Coe
	refreshInterval int // in seconds
	stats           chan []resources.ContainerStats
}

func NewContainerStats(coe *orchestrator.Coe, refreshInterval int) *ContainerStats {
	return &ContainerStats{
		Coe:             *coe,
		refreshInterval: refreshInterval,
	}
}

func (cs *ContainerStats) Run() {
	ticker := time.NewTicker(time.Duration(cs.refreshInterval) * time.Second)
	for {
		select {
		case <-ticker.C:
			cs.stats <- cs.getContainerStats()
		}
	}
}

func (cs *ContainerStats) getContainerStats() []resources.ContainerStats {
	containers, err := cs.Coe.GetContainers()
	if err != nil {
		return nil
	}

	var containerStats []resources.ContainerStats
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
		log.Infof("Container %s stat %v", id, containerInfo)
		containerStat := resources.ContainerStats{
			Id:           id,
			CreatedAt:    containerInfo["created-at"].(string),
			Status:       containerInfo["status"].(string),
			State:        containerInfo["state"].(string),
			Name:         containerInfo["name"].(string),
			Image:        containerInfo["image"].(string),
			RestartCount: containerInfo["restart-count"].(int),
			NetIn:        containerInfo["net-in"].(int64),
			NetOut:       containerInfo["net-out"].(int64),
			MemUsage:     containerInfo["mem-usage"].(int64),
			MemLimit:     containerInfo["mem-limit"].(int64),
			CpuUsage:     containerInfo["cpu-usage"].(int64),
			CpuCapacity:  containerInfo["cpu-capacity"].(int),
			DiskIn:       containerInfo["disk-in"].(int64),
			DiskOut:      containerInfo["disk-out"].(int64),
		}
		containerStats = append(containerStats, containerStat)
	}
	return containerStats
}
