package monitor

import "github.com/shirou/gopsutil/v3/cpu"

func (rm *ResourceMonitor) updateCPU(_ ...string) error {
	cnt, err := cpu.Counts(false)
	if err == nil {
		rm.cpuData.Capacity = cnt
	}

	load, err := cpu.Percent(0, false)
	if err == nil {
		rm.cpuData.Load = load[0]
		rm.cpuData.Load1 = load[0]
		rm.cpuData.Load5 = load[0]
	}
	return err
}
