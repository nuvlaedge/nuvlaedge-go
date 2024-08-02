package types

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_NewWorkerBase(t *testing.T) {
	wb := NewWorkerBase(10, "test")
	assert.NotNil(t, wb, "NewWorkerBase shouldn't be nil")

	assert.Equal(t, WorkerType("test"), wb.wType, "workerType should be test")
	assert.Equal(t, 10, wb.Period, "period should be 10")
	assert.NotNil(t, wb.BaseTicker, "BaseTicker shouldn't be nil")
	assert.Equal(t, NEW, wb.Status, "status should be NEW")
}

func Test_WorkerBase_GetPeriod(t *testing.T) {
	wb := NewWorkerBase(10, "test")
	assert.Equal(t, 10, wb.GetPeriod(), "period should be 10")
}

func Test_WorkerBase_SetPeriod(t *testing.T) {
	wb := NewWorkerBase(10, "test")
	wb.SetPeriod(20)
	assert.Equal(t, 20, wb.Period, "period should be 20")
}

func Test_WorkerBase_GetStatus(t *testing.T) {
	wb := NewWorkerBase(10, "test")
	assert.Equal(t, WorkerStatusReport{Type: "test", Status: NEW}, wb.GetStatus(), "status should be NEW")
}

func Test_WorkerBase_GetType(t *testing.T) {
	wb := NewWorkerBase(10, "test")
	assert.Equal(t, WorkerType("test"), wb.GetType(), "workerType should be test")
}

func Test_WorkerBase_Stop(t *testing.T) {
	wb := NewWorkerBase(10, "test")
	wb.Stop()
	assert.Equal(t, STOPPED, wb.Status, "status should be STOPPED")
}
