package installers

import (
	"fmt"
	"nuvlaedge-go/cli/types"
)

type ProcessInstaller struct {
	BaseBinaryInstaller
}

func NewProcessInstaller(uuid string, version string, flags *types.InstallFlags) *ProcessInstaller {
	return &ProcessInstaller{
		BaseBinaryInstaller: NewBaseInstaller(uuid, version, flags),
	}
}

func (pi *ProcessInstaller) Install() error {
	fmt.Println("Installing NuvlaEdge as a process")
	return pi.PrepareCommonFiles()
}

func (pi *ProcessInstaller) Start() error {
	return nil
}

func (pi *ProcessInstaller) Stop() error {
	return nil
}

func (pi *ProcessInstaller) Remove() error {
	return nil
}

func (pi *ProcessInstaller) Status() string {
	return ""
}

func (pi *ProcessInstaller) String() string {
	return ""
}
