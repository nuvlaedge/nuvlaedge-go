package installers

import (
	"encoding/json"
	"fmt"
	"github.com/pelletier/go-toml/v2"
	"nuvlaedge-go/cli/common"
	"nuvlaedge-go/cli/types"
	"os"
	"path/filepath"
	"runtime"
)

type BaseBinaryInstaller struct {
	Uuid    string
	Version string
	Paths   types.InstallPaths
	Flags   types.InstallFlags

	// Host Configuration
	Os   string
	Arch string
}

func NewBaseInstaller(uuid, version string, installFlags *types.InstallFlags) BaseBinaryInstaller {
	return BaseBinaryInstaller{
		Uuid:    uuid,
		Version: version,
		// Remove the pointer so the initial configuration persists in the original struct
		Flags: *installFlags,
	}
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
	bi.Version = common.GetVersion(v, "nuvlaedge", "nuvlaedge-go")
	fmt.Println("Installing NuvlaEdge version: ", bi.Version)
}

func (bi *BaseBinaryInstaller) AssertInstallationPaths() {
	if bi.Flags.InstallDir == "" {
		bi.Paths = types.NewDefaultInstallPaths(common.HasSudoPermissions())
	} else {
		bi.Paths = types.NewFromBasePath(bi.Flags.InstallDir)
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
	if _, err := os.Stat(common.TempDir); os.IsNotExist(err) {
		err = os.Mkdir(common.TempDir, os.ModePerm)
		if err != nil {
			fmt.Printf("Error creating temporal directory: %s\n", common.TempDir)
			os.Exit(1)
		}
	}

	// 2. Compose download URLS both for the config file and the binary
	bUrl := fmt.Sprintf(common.NuvlaEdgeBinaryURL, bi.Version, bi.Os, bi.Arch, bi.Version)
	fmt.Println("Downloading NuvlaEdge binary from: ", bUrl)
	err := common.InstallFileWithTmpDir(bUrl, "nuvlaedge", bi.Paths.BinaryPath)
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

	confUrl := fmt.Sprintf(common.NuvlaEdgeConfigTemplateURL, bi.Version)
	fmt.Println("Downloading NuvlaEdge configuration template from: ", confUrl)
	err = common.InstallFileWithTmpDir(confUrl, "template.toml", bi.Paths.ConfigPath)
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
