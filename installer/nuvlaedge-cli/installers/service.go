package installers

import (
	"fmt"
	"nuvlaedge-cli/common"
	"nuvlaedge-cli/types"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type ServiceInstaller struct {
	BaseBinaryInstaller
}

func NewServiceInstaller(uuid string, version string, installFlags *types.InstallFlags) *ServiceInstaller {
	return &ServiceInstaller{
		BaseBinaryInstaller: NewBaseInstaller(uuid, version, installFlags),
	}
}

func (si *ServiceInstaller) FillServiceFile() error {
	serviceFile := filepath.Join(common.TempDir, "nuvlaedge.service") // filepath.Join(DefaultServicePath, "nuvlaedge.service")
	fmt.Println("Filling service file: ", serviceFile)

	file, err := os.ReadFile(serviceFile)
	if err != nil {
		fmt.Printf("Error opening configuration file: %s\n", si.Paths.ConfigPath)
		return err
	}

	// Replace the placeholders
	content := string(file)
	content = strings.ReplaceAll(content, "EXEC_PATH_PLACEHOLDER", filepath.Join(si.Paths.BinaryPath, "nuvlaedge"))
	content = strings.ReplaceAll(content, "SETTINGS_PATH_PLACEHOLDER", filepath.Join(si.Paths.ConfigPath, "template.toml"))
	content = strings.ReplaceAll(content, "USER_PLACEHOLDER", os.Getenv("USER"))
	content = strings.ReplaceAll(content, "GROUP_PLACEHOLDER", os.Getenv("USER"))
	content = strings.ReplaceAll(content, "WORKING_DIR_PLACEHOLDER", si.Paths.DatabasePath)

	err = os.WriteFile(filepath.Join(common.DefaultServicePath, "nuvlaedge.service"), []byte(content), 0644)
	if err != nil {
		fmt.Printf("Error writing service file: %s\n", err)
		return err
	}

	return nil
}

func (si *ServiceInstaller) Install() error {
	fmt.Println("Installing NuvlaEdge as a service")
	err := si.PrepareCommonFiles()
	if err != nil {
		fmt.Println("Error preparing common files, binary and configuration")
		return err
	}

	serviceUrl := fmt.Sprintf(common.NuvlaEdgeServiceURL, si.Version)
	fmt.Printf("Downloading service file from %s\n", serviceUrl)
	_ = common.DownloadFile(serviceUrl, filepath.Join(common.TempDir, "nuvlaedge.service"))

	//err = InstallFileWithTmpDir(serviceUrl, "nuvlaedge.service", DefaultServicePath)
	//if err != nil {
	//	fmt.Printf("Error installing NuvlaEdge service file to %s\n", DefaultServicePath)
	//	return err
	//}

	err = si.FillServiceFile()
	if err != nil {
		fmt.Println("Error filling service file")
		return err
	}

	return nil
}

func (si *ServiceInstaller) Start() error {
	fmt.Println("Starting NuvlaEdge as a service...")
	// 1. Start the service
	cmd := exec.Command("systemctl", "enable", "nuvlaedge")
	err := cmd.Run()
	if err != nil {
		return err
	}

	cmd = exec.Command("systemctl", "start", "nuvlaedge")
	err = cmd.Run()
	if err != nil {
		return err
	}
	fmt.Println("Starting NuvlaEdge as a service... Done")

	// 2. Check the status
	fmt.Println("Check NuvlaEdge health status after 5 seconds...")
	cmd = exec.Command("systemctl", "is-active", "nuvlaedge")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error checking NuvlaEdge status: ", err)
		fmt.Println("NuvlaEdge might or might not be running")
		return err
	}

	status := strings.TrimSpace(string(output))
	if status != "active" {
		fmt.Println("NuvlaEdge is not running")
		return fmt.Errorf("NuvlaEdge is not running")
	}

	fmt.Printf("Check NuvlaEdge health status after 5 seconds... %s\n", status)

	return nil
}

func (si *ServiceInstaller) Stop() error {
	return nil
}

func (si *ServiceInstaller) Remove() error {
	return nil
}

func (si *ServiceInstaller) Status() string {
	return ""
}

func (si *ServiceInstaller) String() string {
	return ""
}
