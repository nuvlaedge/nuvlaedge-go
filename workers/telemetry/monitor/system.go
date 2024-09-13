package monitor

import (
	"context"
	"github.com/shirou/gopsutil/v3/host"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/common/constants"
	"nuvlaedge-go/types/metrics"
	"os"
	"runtime"
	"time"
)

type SystemMonitor struct {
	BaseMonitor

	systemData metrics.SystemMetrics
}

func NewSystemMonitor(period int, repChan chan metrics.Metric) *SystemMonitor {

	return &SystemMonitor{
		BaseMonitor: NewBaseMonitor(period, repChan),
	}
}

func (sm *SystemMonitor) Run(ctx context.Context) error {
	sm.SetRunning()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-sm.Ticker.C:
			// Send metric to channel
			sm.updateMetrics()
			sm.reportChan <- sm.systemData
		}
	}
}

func (sm *SystemMonitor) updateMetrics() {
	h, err := os.Hostname()
	if err == nil {
		sm.systemData.Hostname = h
	} else {
		log.Warnf("Error retrieving hostname: %v", err)
	}
	sm.systemData.OperatingSystem = runtime.GOOS
	sm.systemData.Architecture = runtime.GOARCH

	unixTime, err := host.BootTime()
	if err == nil {
		// #nosec
		tTime := time.Unix(int64(unixTime), 0)
		sm.systemData.LastBoot = tTime.Format(constants.DatetimeFormat)
	} else {
		log.Warnf("Error retrieving last boot time: %v", err)
	}
}
