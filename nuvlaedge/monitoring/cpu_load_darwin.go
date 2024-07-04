package monitoring

import (
	"github.com/shirou/gopsutil/v3/cpu"
	log "github.com/sirupsen/logrus"
	"time"
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

	log.Info("Starting CPU metrics reading")
	go c.load1Updater()
	go c.load5Updater()
	go c.load15Updater()
}

func (c *CPUMetrics) load1Updater() {
	for {
		loads1, err := cpu.Percent(time.Minute, false)
		if err != nil {
			log.Errorf("Error getting CPU load-1: %s", err)
		}
		c.Load1 = loads1[0]
	}
}

func (c *CPUMetrics) load5Updater() {
	for {
		loads5, err := cpu.Percent(5*time.Minute, false)
		if err != nil {
			log.Errorf("Error getting CPU load-5: %s", err)
		}
		c.Load5 = loads5[0]
	}
}

func (c *CPUMetrics) load15Updater() {
	for {
		load15, err := cpu.Percent(15*time.Minute, false)
		if err != nil {
			log.Errorf("Error getting CPU load-15: %s", err)
		}
		c.Load = load15[0]
	}
}

func (c *CPUMetrics) Update() error {
	return nil
}
