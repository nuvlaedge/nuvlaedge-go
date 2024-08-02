package monitor

import (
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func init() {
	// Set log level to panic to avoid logs during tests
	log.SetLevel(log.PanicLevel)
}

func TestUpdateCPU(t *testing.T) {
	// Setup
	mockData := "0.10 0.20 0.30 1/234 12345\n"
	tempFile, err := os.CreateTemp("", "loadavg")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	_, err = tempFile.WriteString(mockData)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()

	resourceMonitor := &ResourceMonitor{}

	// Execute
	err = resourceMonitor.updateCPU(tempFile.Name())

	// Verify
	assert.NoError(t, err)
	assert.Equal(t, 0.10, resourceMonitor.cpuData.Load1)
	assert.Equal(t, 0.20, resourceMonitor.cpuData.Load5)
	assert.Equal(t, 0.30, resourceMonitor.cpuData.Load)
}
