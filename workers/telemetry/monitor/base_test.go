package monitor

import (
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"nuvlaedge-go/common/constants"
	"testing"
)

func init() {
	// Set log level to panic to avoid logs during tests
	log.SetLevel(log.PanicLevel)
}

func TestBaseMonitor_SetPeriod_UpdatesPeriodWhenGreaterThanMinimum(t *testing.T) {
	baseMonitor := NewBaseMonitor(1, mockChan) // Assuming 1 is less than the minimum allowed period
	newPeriod := 30                            // A valid period value greater than the minimum
	baseMonitor.SetPeriod(newPeriod)
	assert.Equal(t, newPeriod, baseMonitor.GetPeriod(), "Expected period to be updated to the new value")
}

func TestBaseMonitor_SetPeriod_SetsToMinimumWhenBelowAllowed(t *testing.T) {
	baseMonitor := NewBaseMonitor(1, mockChan) // Assuming 1 is less than the minimum allowed period
	tooLowPeriod := -5                         // An invalid period value
	baseMonitor.SetPeriod(tooLowPeriod)
	assert.Equal(t, constants.MinTelemetryPeriod, baseMonitor.GetPeriod(), "Expected period to be set to the minimum allowed value")
}

func TestBaseMonitor_Running_ReturnsTrueWhenRunning(t *testing.T) {
	baseMonitor := NewBaseMonitor(1, mockChan)
	baseMonitor.running = true
	assert.True(t, baseMonitor.Running(), "Expected Running() to return true when monitor is running")
}
