package monitor

import (
	"context"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/common/constants"
	"nuvlaedge-go/types/metrics"
	"sync"
	"time"
)

type BaseMonitor struct {
	reportChan  chan metrics.Metric
	period      int
	Ticker      *time.Ticker
	running     bool
	runningLock *sync.Mutex
}

func NewBaseMonitor(period int, repChan chan metrics.Metric) BaseMonitor {
	base := BaseMonitor{
		running:     false,
		reportChan:  repChan,
		runningLock: &sync.Mutex{},
	}
	base.SetPeriod(period)
	return base
}

func (bm *BaseMonitor) GetChannel() chan metrics.Metric {
	return bm.reportChan
}

func (bm *BaseMonitor) GetPeriod() int {
	return bm.period
}

func (bm *BaseMonitor) SetPeriod(p int) {
	if p < constants.MinTelemetryPeriod {
		p = constants.MinTelemetryPeriod
		log.Errorf("period must be greater than 0, setting default %d", bm.period)
	}

	bm.period = p

	if bm.Ticker != nil {
		bm.Ticker.Reset(time.Duration(bm.period) * time.Second)
	} else {
		bm.Ticker = time.NewTicker(time.Duration(bm.period) * time.Second)
	}
}

func (bm *BaseMonitor) Running() bool {
	bm.runningLock.Lock()
	defer bm.runningLock.Unlock()
	return bm.running
}

func (bm *BaseMonitor) SetRunning() {
	bm.runningLock.Lock()
	bm.running = true
	bm.runningLock.Unlock()
}

func (bm *BaseMonitor) Close() error {
	bm.runningLock.Lock()
	defer bm.runningLock.Unlock()
	bm.running = false
	bm.Ticker.Stop()
	return nil
}

type NuvlaEdgeMonitor interface {
	Run(ctx context.Context) error
	GetChannel() chan metrics.Metric
	GetPeriod() int
	SetPeriod(int)
	Running() bool
	Close() error
}
