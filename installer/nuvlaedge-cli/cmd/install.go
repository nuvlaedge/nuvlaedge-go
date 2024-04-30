// Package cmd
/*
Copyright © 2024 SixSq SA <support@sixsq.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type InstallFlags struct {
	Version    string
	InstallDir string
	ConfigFile string

	Service    bool
	Process    bool
	Docker     bool
	Kubernetes bool

	// Run flags
	Uuid string
}

func (f *InstallFlags) String() string {
	s, _ := json.MarshalIndent(f, "", "  ")
	return string(s)
}

var installFlags InstallFlags

func ValidateInstallFlags() error {
	fmt.Println("Service mode: ", installFlags.Service)
	if installFlags.Service && !hasSudoPermissions() {
		return fmt.Errorf("service mode requires sudo permissions")
	}

	if installFlags.Docker && !isDockerRunning() {
		return fmt.Errorf("docker mode requires Docker running")
	}

	if installFlags.Kubernetes && !isKubernetesRunning() {
		return fmt.Errorf("kubernetes mode requires Kubernetes running")
	}

	return nil
}

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Installs NuvlaEdge on the local machine",
	Long: `Installs NuvlaEdge on the local machine. The installation process will download the binary and the 
configuration file. Then place them in the appropriate directories. If the configuration file is not provided.

The installation process will also create the necessary directories for the binary and the configuration file.
The installation also allows to run the NuvlaEdge after the installation is completed. The start process can be on tree
modes:
 - Run as a service (default)
 - Run as a detached process
 - Run as a foreground process`,

	Run: func(cmd *cobra.Command, args []string) {
		// Validate the flags
		err := ValidateInstallFlags()
		if err != nil {
			cmd.Printf("Error validating flags: %v\n", err)
			os.Exit(1)
		}

		cmd.Println("Running install with configuration: %s", installFlags.String())

		installer := GetInstaller(&installFlags)
		if installer == nil {
			fmt.Println("Error getting installer")
			os.Exit(1)
		}

		fmt.Println("Installing NuvlaEdge...")
		err = installer.Install()
		if err != nil {
			cmd.Printf("Error installing NuvlaEdge: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Installing NuvlaEdge... Done")

		fmt.Println("Starting NuvlaEdge...")
		err = installer.Start()
		if err != nil {
			cmd.Printf("Error starting NuvlaEdge: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Starting NuvlaEdge... Done")
	},
}

func init() {
	rootCmd.AddCommand(installCmd)

	installFlags = InstallFlags{}

	installCmd.Flags().StringVarP(&installFlags.Version, "version", "v", "latest", "Version of the NuvlaEdge to install")
	installCmd.Flags().StringVarP(&installFlags.InstallDir, "dir", "d", "", "Directory where to install NuvlaEdge")
	installCmd.Flags().StringVarP(&installFlags.ConfigFile, "config-file", "c", "", "File to use as configuration for NuvlaEdge")

	installCmd.Flags().BoolVar(&installFlags.Service, "service", false, "Run NuvlaEdge as a service")
	installCmd.Flags().BoolVar(&installFlags.Process, "process", false, "Run NuvlaEdge as a process")
	installCmd.Flags().BoolVar(&installFlags.Docker, "docker", false, "Run NuvlaEdge as a Docker container")
	installCmd.Flags().BoolVar(&installFlags.Kubernetes, "kubernetes", false, "Run NuvlaEdge as a Kubernetes pod")

	// Mutually exclusive flags
	installCmd.MarkFlagsMutuallyExclusive("service", "process", "docker", "kubernetes")
	installCmd.MarkFlagsOneRequired("service", "process", "docker", "kubernetes")

	// To run the NuvlaEdge after installing, we need the UUID of the NuvlaEdge
	installCmd.Flags().StringVar(&installFlags.Uuid, "uuid", "", "UUID to assign to the NuvlaEdge instance")
	err := installCmd.MarkFlagRequired("uuid")
	if err != nil {
		fmt.Println("Error marking uuid as required")
	}
}

func getVersion(v string) string {
	// 1. Get all available versions
	releases, err := getAllReleases("nuvlaedge", "nuvlaedge-go")
	if err != nil {
		fmt.Println("Error getting releases")
		return ""
	}
	fmt.Println("Releases: ", releases)

	// 2.1 If v is "latest", return the latest version
	if v == "latest" {
		return releases[0].TagName
	}

	// 2.2 If v is not "latest", return check availability and return the version
	for _, r := range releases {
		cleanRemote := strings.Replace(r.TagName, "v", "", 1)
		cleanLocal := strings.Replace(v, "v", "", 1)
		if cleanLocal == cleanRemote {
			return v
		}
	}
	fmt.Printf("Requested version %s not found, installing latest\n", v)
	return releases[0].TagName
}

type Installer interface {
	Install() error
	Start() error
	Stop() error
	Remove() error
	Status() string
	String() string
}

func GetInstaller(flags *InstallFlags) Installer {
	if flags.Process {
		return NewProcessInstaller(flags.Uuid, flags.Version, flags)
	}
	if flags.Service {
		return NewServiceInstaller(flags.Uuid, flags.Version, flags)
	}
	rootCmd.Println("Invalid run mode, select proper mode...")
	return nil
}

type BaseBinaryInstaller struct {
	Uuid    string
	Version string
	Paths   InstallPaths
	Flags   InstallFlags

	// Host Configuration
	Os   string
	Arch string
}

func (bi *BaseBinaryInstaller) GetHostConfig() {
	// 1. Get OS
	bi.Os = runtime.GOOS
	fmt.Println("Installing NuvlaEdge on OS: ", bi.Os)

	// 2. Get ARCH
	bi.Arch = runtime.GOARCH
	fmt.Println("Installing NuvlaEdge on ARCH: ", bi.Arch)
}

func (bi *BaseBinaryInstaller) CleanVersion(v string) {
	fmt.Println("Requested NuvlaEdge version: ", v)
	bi.Version = getVersion(v)
	fmt.Println("Installing NuvlaEdge version: ", bi.Version)
}

func (bi *BaseBinaryInstaller) AssertInstallationPaths() {
	if bi.Flags.InstallDir == "" {
		bi.Paths = NewDefaultInstallPaths(hasSudoPermissions())
	} else {
		bi.Paths = NewFromBasePath(bi.Flags.InstallDir)
	}
	//
	b, _ := json.MarshalIndent(&bi.Paths, "", "  ")
	fmt.Printf("Installing NuvlaEdge in:\n%s\n ", string(b))
	bi.Paths.MakePaths()
}

func (bi *BaseBinaryInstaller) FillConfigFile() error {
	// 1. Read the configuration file
	file, err := os.OpenFile(filepath.Join(bi.Paths.ConfigPath, "template.toml"), os.O_RDWR, 0644)
	if err != nil {
		fmt.Printf("Error opening configuration file: %s\n", bi.Paths.ConfigPath)
		return err
	}
	defer file.Close()

	var conf map[string]interface{}
	if err = toml.NewDecoder(file).Decode(&conf); err != nil {
		fmt.Println("Error decoding configuration file")
		return err
	}

	conf["data-location"] = bi.Paths.DatabasePath
	conf["config-location"] = bi.Paths.ConfigPath

	var logConf map[string]interface{}
	logConf = conf["logging"].(map[string]interface{})
	logConf["log-file"] = filepath.Join(bi.Paths.LogsPath, "nuvlaedge.log")
	conf["logging"] = logConf

	var agentConf map[string]interface{}
	agentConf = conf["agent"].(map[string]interface{})
	agentConf["nuvlaedge-uuid"] = bi.Uuid
	conf["agent"] = agentConf

	// Write the modified configuration back to the file
	_, _ = file.Seek(0, 0) // Reset the file pointer to the beginning
	_ = file.Truncate(0)   // Clear the file content

	if err := toml.NewEncoder(file).Encode(conf); err != nil {
		fmt.Println("Error writing TOML file:", err)
		return err
	}

	return nil
}

func (bi *BaseBinaryInstaller) PrepareCommonFiles() error {
	bi.GetHostConfig()
	bi.CleanVersion(bi.Flags.Version)
	bi.AssertInstallationPaths()
	// Common installation section
	// 1. Create temporal directory
	if _, err := os.Stat(TempDir); os.IsNotExist(err) {
		err = os.Mkdir(TempDir, os.ModePerm)
		if err != nil {
			fmt.Printf("Error creating temporal directory: %s\n", TempDir)
			os.Exit(1)
		}
	}

	// 2. Compose download URLS both for the config file and the binary
	bUrl := fmt.Sprintf(NuvlaEdgeBinaryURL, bi.Version, bi.Os, bi.Arch, bi.Version)
	fmt.Println("Downloading NuvlaEdge binary from: ", bUrl)
	err := InstallFileWithTmpDir(bUrl, "nuvlaedge", bi.Paths.BinaryPath)
	if err != nil {
		fmt.Printf("Error installing NuvlaEdge binary to %s\n", bi.Paths.BinaryPath)
		return err
	}
	// Make binary file executable
	err = os.Chmod(filepath.Join(bi.Paths.BinaryPath, "nuvlaedge"), 0755)
	if err != nil {
		fmt.Printf("Error making binary executable: %s\n", err)
		return err
	}

	confUrl := fmt.Sprintf(NuvlaEdgeConfigTemplateURL, bi.Version)
	fmt.Println("Downloading NuvlaEdge configuration template from: ", confUrl)
	err = InstallFileWithTmpDir(confUrl, "template.toml", bi.Paths.ConfigPath)
	if err != nil {
		fmt.Printf("Error installing NuvlaEdge configuration to %s\n", bi.Paths.ConfigPath)
	}
	// We need to edit the config file to include, the uuid, the version and the paths
	err = bi.FillConfigFile()
	if err != nil {
		fmt.Println("Error filling configuration file")
		return err
	}
	return nil
}

func NewBaseInstaller(uuid, version string, installFlags *InstallFlags) BaseBinaryInstaller {
	return BaseBinaryInstaller{
		Uuid:    uuid,
		Version: version,
		// Remove the pointer so the initial configuration persists in the original struct
		Flags: *installFlags,
	}
}

type ServiceInstaller struct {
	BaseBinaryInstaller
}

func NewServiceInstaller(uuid string, version string, installFlags *InstallFlags) *ServiceInstaller {
	return &ServiceInstaller{
		BaseBinaryInstaller: NewBaseInstaller(uuid, version, installFlags),
	}
}

func (si *ServiceInstaller) FillServiceFile() error {
	serviceFile := filepath.Join(TempDir, "nuvlaedge.service") // filepath.Join(DefaultServicePath, "nuvlaedge.service")
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

	err = os.WriteFile(filepath.Join(DefaultServicePath, "nuvlaedge.service"), []byte(content), 0644)
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

	serviceUrl := fmt.Sprintf(NuvlaEdgeServiceURL, si.Version)
	fmt.Printf("Downloading service file from %s\n", serviceUrl)
	_ = downloadFile(serviceUrl, filepath.Join(TempDir, "nuvlaedge.service"))

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
	cmd := exec.Command("systemctl", "start", "nuvlaedge")
	err := cmd.Run()
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

type ProcessInstaller struct {
	BaseBinaryInstaller
}

func NewProcessInstaller(uuid string, version string, flags *InstallFlags) *ProcessInstaller {
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
