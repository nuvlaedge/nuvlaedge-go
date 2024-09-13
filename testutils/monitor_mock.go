package testutils

import (
	"context"
	"nuvlaedge-go/types/metrics"
	"sync"
)

type MonitorMock struct {
	// MonitorMock is a mock for the Monitor interface
	runningLock *sync.Mutex
	running     bool
}

func NewMonitorMock() *MonitorMock {
	return &MonitorMock{
		running:     false,
		runningLock: &sync.Mutex{},
	}
}

// Implement the Monitor interface
func (m *MonitorMock) Run(ctx context.Context) error {
	m.SetRunning()
	return nil
}

func (m *MonitorMock) SetRunning() {
	m.runningLock.Lock()
	defer m.runningLock.Unlock()
	m.running = true
}
func (m *MonitorMock) Running() bool {
	m.runningLock.Lock()
	defer m.runningLock.Unlock()
	return m.running
}

func (m *MonitorMock) Stop() {
	m.runningLock.Lock()
	defer m.runningLock.Unlock()
	m.running = false
}

func (m *MonitorMock) GetChannel() chan metrics.Metric {
	return nil
}

func (m *MonitorMock) GetPeriod() int {
	return 0
}

func (m *MonitorMock) SetPeriod(int) {

}

func (m *MonitorMock) Close() error {
	return nil
}
