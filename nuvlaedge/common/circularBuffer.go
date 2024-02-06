package common

import (
	"fmt"
	"sync"
)

type CircularBuffer struct {
	data      []float64
	start     int
	end       int
	size      int
	lockMutex *sync.Mutex
}

func NewCircularBuffer(size int) *CircularBuffer {
	return &CircularBuffer{
		data:      make([]float64, size),
		size:      size,
		lockMutex: &sync.Mutex{},
	}
}

func (cb *CircularBuffer) Add(value float64) {
	cb.lockMutex.Lock()
	defer cb.lockMutex.Unlock()
	cb.data[cb.end] = value
	cb.end = (cb.end + 1) % cb.size

	if cb.end == cb.start {
		cb.start = (cb.start + 1) % cb.size
	}
}

func (cb *CircularBuffer) Get(index int) (float64, error) {
	if index < 0 || index >= cb.size {
		return 0, fmt.Errorf("index out of range")
	}
	cb.lockMutex.Lock()
	defer cb.lockMutex.Unlock()
	return cb.data[(cb.start+index)%cb.size], nil
}

func (cb *CircularBuffer) GetLatest(X int) ([]float64, error) {
	if X < 0 || X > cb.size {
		return nil, fmt.Errorf("requested number of elements is out of range")
	}
	cb.lockMutex.Lock()
	defer cb.lockMutex.Unlock()

	result := make([]float64, X)
	for i := 0; i < X; i++ {
		index := (cb.end - 1 - i + cb.size) % cb.size
		result[i] = cb.data[index]
	}

	return result, nil
}

func (cb *CircularBuffer) GetLatestAvg(X int) (float64, error) {
	if X < 0 || X > cb.size {
		return 0, fmt.Errorf("requested number of elements is out of range")
	}
	cb.lockMutex.Lock()
	defer cb.lockMutex.Unlock()

	sum := 0.0
	for i := 0; i < X; i++ {
		index := (cb.end - 1 - i + cb.size) % cb.size
		sum += cb.data[index]
	}

	return sum / float64(X), nil
}
