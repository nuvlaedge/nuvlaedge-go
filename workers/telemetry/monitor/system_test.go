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
func TestNewSystemMonitor(t *testing.T) {
	period := 10
	reportChan := make(chan metrics.Metric, 1)

	systemMonitor := NewSystemMonitor(period, reportChan)

	assert.NotNil(t, systemMonitor)
	assert.Equal(t, period, systemMonitor.GetPeriod())
	assert.Equal(t, reportChan, systemMonitor.reportChan)
}

func TestRun_ContextCancelled_StopsRunning(t *testing.T) {
	reportChan := make(chan metrics.Metric, 1)
	systemMonitor := NewSystemMonitor(10, reportChan)
	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := systemMonitor.Run(ctx)
		assert.Equal(t, context.Canceled, err)
	}()

	cancel()
	wg.Wait()
}

func TestRun_UpdatesMetricsAndSendsToChannel(t *testing.T) {
	reportChan := make(chan metrics.Metric, 1)
	systemMonitor := NewSystemMonitor(10, reportChan)
	systemMonitor.Ticker = time.NewTicker(1 * time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	go func() {
		err := systemMonitor.Run(ctx)
		assert.Equal(t, context.Canceled, err)
	}()

	select {
	case <-reportChan:
		// Expected to receive a metric update
	case <-time.After(4 * time.Second):
		t.Error("Expected to receive a metric update but did not")
	}

	cancel()
}
