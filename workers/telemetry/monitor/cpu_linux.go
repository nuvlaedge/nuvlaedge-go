package monitor

import (
	"errors"
	"github.com/shirou/gopsutil/v3/cpu"
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"
	"strings"
)

func (rm *ResourceMonitor) updateCPU(files ...string) error {
	cpuCount, err := cpu.Counts(false)
	if err != nil {
		log.Errorf("Error getting CPU capacity: %s", err)
		rm.cpuData.Capacity = 1
	} else {
		rm.cpuData.Capacity = cpuCount
	}

	var file string
	if len(file) <= 0 {
		file = "/proc/loadavg"
	} else {
		file = files[0]
	}
	// #nosec
	raw, err := os.ReadFile(file)
	if err != nil {
		return err
	} else {
		l, _ := cpu.Percent(0, false)
		rm.cpuData.Load = l[0]
	}

	values := strings.Split(string(raw), " ")
	if len(values) != 5 {
		return errors.New("/proc/loadavg field count mismatch")
	}

	n, err := strconv.ParseFloat(values[0], 64)
	if err != nil {
		return errors.New("LoadAverage1 parse error")
	}
	rm.cpuData.Load1 = n

	n, err = strconv.ParseFloat(values[1], 64)
	if err != nil {
		return errors.New("LoadAverage5 parse error")
	}
	rm.cpuData.Load5 = n

	n, err = strconv.ParseFloat(values[2], 64)
	if err != nil {
		return errors.New("LoadAverage10 parse error")
	}
	rm.cpuData.Load = n
	return nil
}
