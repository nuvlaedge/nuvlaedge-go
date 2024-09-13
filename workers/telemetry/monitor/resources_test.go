package monitor

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"nuvlaedge-go/types/metrics"
	"sync"
	"testing"
	"time"
)

func init() {
	// Set log level to panic to avoid logs during tests
	log.SetLevel(log.PanicLevel)
}

func Test_NewResourceMonitor_WithPositivePeriod_ReturnsInitializedMonitor(t *testing.T) {
	period := 15
	reportChan := make(chan metrics.Metric, 1)

	resourceMonitor := NewResourceMonitor(period, reportChan)

	assert.NotNil(t, resourceMonitor)
	assert.Equal(t, period, resourceMonitor.GetPeriod())
	assert.Equal(t, reportChan, resourceMonitor.reportChan)
}

func Test_Run_ContextCancelledBeforeUpdate_StopsWithoutSendingMetrics(t *testing.T) {
	reportChan := make(chan metrics.Metric, 10)
	resourceMonitor := NewResourceMonitor(1, reportChan)
	ctx, cancel := context.WithCancel(context.Background())

	cancel() // Cancel context before running

	err := resourceMonitor.Run(ctx)

	assert.Equal(t, context.Canceled, err, "Expected context cancelled error")
	assert.Empty(t, reportChan, "Expected no metrics to be sent after context is cancelled")
}

func TestRun_SendsMultipleMetricTypesToChannel(t *testing.T) {
	reportChan := make(chan metrics.Metric, 10)
	resourceMonitor := NewResourceMonitor(15, reportChan)
	resourceMonitor.Ticker = time.NewTicker(1 * time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go func() {
		resourceMonitor.sendMetrics()
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	var ramMetrics, cpuMetrics, diskMetrics, ifaceMetrics, networkMetrics bool
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case metric := <-reportChan:
				switch metric.(type) {
				case metrics.RamMetrics:
					ramMetrics = true
				case metrics.CPUMetrics:
					cpuMetrics = true
				case metrics.DiskMetrics:
					diskMetrics = true
				case metrics.IfacesMetrics:
					ifaceMetrics = true
				case metrics.NetworkMetrics:
					networkMetrics = true
				}
				if ramMetrics && cpuMetrics && diskMetrics && ifaceMetrics && networkMetrics {
					return
				}
			}
		}
	}()
	wg.Wait()
	assert.True(t, ramMetrics, "Expected RAM metrics to be sent")
	assert.True(t, cpuMetrics, "Expected CPU metrics to be sent")
	assert.True(t, diskMetrics, "Expected Disk metrics to be sent")
	assert.True(t, ifaceMetrics, "Expected Interface metrics to be sent")
	assert.True(t, networkMetrics, "Expected Network metrics to be sent")
}

func Test_ResourcesMonitor_UpdateMetrics(t *testing.T) {
	reportChan := make(chan metrics.Metric, 1)
	resourceMonitor := NewResourceMonitor(15, reportChan)
	err := resourceMonitor.updateMetrics()
	assert.Nil(t, err)
}
