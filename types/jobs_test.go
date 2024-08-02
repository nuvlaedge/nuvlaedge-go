package types

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_NewRunningJobs(t *testing.T) {
	// Test code here
	js := NewRunningJobs()
	assert.NotNil(t, js, "NewRunningJobs shouldn't be nil")
	assert.NotNil(t, js.jobs, "jobs map shouldn't be nil")
	assert.NotNil(t, js.lock, "lock shouldn't be nil")
}

func Test_JobRegistry_Add(t *testing.T) {
	// Test code here
	js := NewRunningJobs()
	j := &RunningJob{jobId: "1", jobType: "test", running: true}
	assert.True(t, js.Add(j), "Job should be added")
	assert.False(t, js.Add(j), "Job should not be added")

	_, ok := js.jobs["1"]
	assert.True(t, ok, "Job should be in the map")
}

func Test_JobRegistry_Remove(t *testing.T) {
	// Test code here
	js := NewRunningJobs()
	j := &RunningJob{jobId: "1", jobType: "test", running: true}
	js.Add(j)

	assert.True(t, js.Remove("1"), "Job should be removed")
	assert.False(t, js.Remove("1"), "Job should not be removed")

	_, ok := js.jobs["1"]
	assert.False(t, ok, "Job should not be in the map")
}

func Test_JobRegistry_Get(t *testing.T) {
	// Test code here
	js := NewRunningJobs()
	j := &RunningJob{jobId: "1", jobType: "test", running: true}
	js.Add(j)

	j2, ok := js.Get("1")
	assert.True(t, ok, "Job should be in the map")
	assert.Equal(t, j, j2, "Jobs should be equal")

	j3, ok := js.Get("2")
	assert.False(t, ok, "Job should not be in the map")
	assert.Nil(t, j3, "Job should be nil")
}

func Test_JobRegistry_Exists(t *testing.T) {
	// Test code here
	js := NewRunningJobs()
	j := &RunningJob{jobId: "1", jobType: "test", running: true}
	js.Add(j)

	assert.True(t, js.Exists("1"), "Job should exist")
	assert.False(t, js.Exists("2"), "Job should not exist")
}

func Test_JobRegistry_String(t *testing.T) {
	// Test code here
	js := NewRunningJobs()
	j := &RunningJob{jobId: "1", jobType: "test", running: true}
	js.Add(j)

	assert.NotEmpty(t, js.String(), "String should not be empty")
}
