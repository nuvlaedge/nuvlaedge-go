package monitor

import (
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func init() {
	// Set log level to panic to avoid logs during tests
	log.SetLevel(log.PanicLevel)
}
func TestUpdateCPU_SuccessfullyUpdatesCPUData(t *testing.T) {
	resourceMonitor := &ResourceMonitor{}

	err := resourceMonitor.updateCPU()
	l := resourceMonitor.cpuData.Load
	assert.NoError(t, err)

	assert.Equal(t, l, resourceMonitor.cpuData.Load1)
	assert.Equal(t, l, resourceMonitor.cpuData.Load5)
}
