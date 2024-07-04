package monitoring

import (
	"errors"
	"github.com/msaf1980/go-stringutils"
	"github.com/shirou/gopsutil/v3/cpu"
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"
	"strings"
)

type CPUMetrics struct {
	Load  float64 `json:"load"`
	Load1 float64 `json:"load-1"`
	Load5 float64 `json:"load-5"`
	//SystemCalls        int64   `json:"system-calls"`
	Capacity int `json:"capacity"`
	//Interrupts         int64   `json:"interrupts"`
	Topic string `json:"topic"`
	//SoftwareInterrupts int64   `json:"software-interrupts"`
	//ContextSwitches    int64   `json:"context-switches"`
}

func NewCPUMetrics() *CPUMetrics {
	c := &CPUMetrics{
		Topic: "cpu",
	}
	return c
}

func (c *CPUMetrics) Run() {
	cpuCount, err := cpu.Counts(false)
	if err != nil {
		log.Errorf("Error getting CPU capacity: %s", err)
	}
	c.Capacity = cpuCount
}

func (c *CPUMetrics) Update() error {
	raw, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return err
	}

	s := strings.TrimRight(stringutils.UnsafeString(raw), "\n")

	values := strings.Split(s, " ")
	if len(values) != 5 {
		return errors.New("/proc/loadavg field count mismatch")
	}

	n, err := strconv.ParseFloat(values[0], 64)
	if err != nil {
		return errors.New("LoadAverage1 parse error")
	}
	c.Load1 = n

	n, err = strconv.ParseFloat(values[1], 64)
	if err != nil {
		return errors.New("LoadAverage5 parse error")
	}
	c.Load5 = n

	n, err = strconv.ParseFloat(values[2], 64)
	if err != nil {
		return errors.New("LoadAverage10 parse error")
	}
	c.Load = n
	return nil
}
