package telemetry

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"native-nuvlaedge/src/coe"
	"native-nuvlaedge/src/common"
	"os"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/host"
)

type SysInfoData struct {
	Hostname          string `json:"hostname"`
	OperatingSystem   string `json:"operating-system"`
	Architecture      string `json:"architecture"`
	LastBoot          string `json:"last-boot"`
	DockerVersion     string `json:"docker-server-version"`
	KubernetesVersion string `json:"kubernetes-version"`
}

type SystemInfo struct {
	data       SysInfoData
	coeClient  coe.Coe
	period     int
	reportChan chan SysInfoData
}

func NewSystemInfo(period int, reportChan chan SysInfoData, coeClient coe.Coe) SystemInfo {
	sysInfo := SystemInfo{
		data:       SysInfoData{},
		coeClient:  coeClient,
		period:     period,
		reportChan: reportChan,
	}

	return sysInfo
}

func (s *SystemInfo) Update() error {
	defer common.ExecutionTime(time.Now(), "SystemInfo Update")
	err := s.updateArchitecture()
	common.GenericErrorHandler("Error retrieving System architecture", err)

	err = s.updateOperatingSystem()
	common.GenericErrorHandler("Error retrieving system OperatingSystem", err)

	err = s.updateHostname()
	common.GenericErrorHandler("Error retrieving hostname", err)

	err = s.updateLastBoot()
	common.GenericErrorHandler("Error retrieving last boot", err)

	err = s.updateCoeVersions()
	common.GenericErrorHandler("Error retrieving COE client version", err)
	return err
}

func (s *SystemInfo) updateHostname() error {
	hostname, err := os.Hostname()
	s.data.Hostname = hostname
	return err
}

func (s *SystemInfo) updateOperatingSystem() error {
	s.data.OperatingSystem = runtime.GOOS
	if s.data.OperatingSystem == "" {
		return fmt.Errorf("error retrieving OperatingSystem from runtime package")
	}
	return nil
}

func (s *SystemInfo) updateArchitecture() error {
	s.data.Architecture = runtime.GOARCH
	if s.data.Architecture == "" {
		return fmt.Errorf("error retrieving architecture from runtime package")
	}
	return nil
}

func (s *SystemInfo) updateLastBoot() error {
	epochTime, err := host.BootTime()
	if err != nil {
		return err
	}

	tTime := time.Unix(int64(epochTime), 0)
	s.data.LastBoot = tTime.Format(common.DatetimeFormat)
	log.Infof("Last boot time found %s", s.data.LastBoot)
	return nil
}

func (s *SystemInfo) updateCoeVersions() error {
	v, _ := s.coeClient.GetCoeVersion()
	switch s.coeClient.GetCoeType() {
	case coe.DockerType:
		s.data.DockerVersion = v
	case coe.KubernetesType:
		s.data.KubernetesVersion = v
	}
	return nil
}

func (s *SystemInfo) SetPeriod(newPeriod int) {
	s.period = newPeriod
}

func (s *SystemInfo) GetPeriod() int {
	return s.period
}

func (s *SystemInfo) String() string {
	return fmt.Sprintf(
		"Hostname: %s,\n "+
			"OperatingSystem: %s,\n "+
			"Architecture: %s,\n "+
			"LastBoot: %s,\n "+
			"DockerVersion: %s,\n "+
			"KubernetesVersion: %s",
		s.data.Hostname, s.data.OperatingSystem, s.data.Architecture,
		s.data.LastBoot, s.data.DockerVersion, s.data.KubernetesVersion,
	)
}

func (s *SystemInfo) Run() {
	log.Info("Starting system info update")
	for {
		startTime := time.Now()
		err := s.Update()

		if err != nil {
			log.Errorf("error %s updating system information", err)
		} else {
			s.reportChan <- s.data
			err = common.WaitPeriodicAction(startTime, s.period, "SystemInfo Update")
			if err != nil {
				panic(err)
			}
		}
	}
}
