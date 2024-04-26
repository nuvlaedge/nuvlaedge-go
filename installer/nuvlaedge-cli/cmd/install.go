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
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type RunMode string

const (
	ProcessMode RunMode = "process"
	ServiceMode RunMode = "service"
)

type InstallFlags struct {
	Version     string
	InstallDir  string
	ConfigFile  string
	InstallMode *strEnum
	// Run flags
	Run      bool
	Uuid     string
	Detached bool
}

func (f *InstallFlags) String() string {
	s, _ := json.MarshalIndent(f, "", "  ")
	return string(s)
}

var installFlags InstallFlags

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
		cmd.Println("Running install with configuration: %s", installFlags.String())
		paths, err := installNuvlaEdge(installFlags)
		if err != nil {
			cmd.Printf("Error installing NuvlaEdge: %v\n", err)
			os.Exit(1)
		}

		if installFlags.Run {
			err = startNuvlaEdge(installFlags, paths)
			if err != nil {
				cmd.Printf("Error starting NuvlaEdge: %v\n", err)
				os.Exit(1)
			}
		}
	},
}

func init() {
	installFlags = InstallFlags{}

	rootCmd.AddCommand(installCmd)
	installFlags.InstallMode = newStrEnum(string(ProcessMode), string(ProcessMode), string(ServiceMode))

	installCmd.Flags().StringVarP(&installFlags.Version, "version", "v", "latest", "Version of the NuvlaEdge to install")
	installCmd.Flags().StringVarP(&installFlags.InstallDir, "dir", "d", "", "Directory where to install NuvlaEdge")
	installCmd.Flags().StringVarP(&installFlags.ConfigFile, "config-file", "c", "", "File to use as configuration for NuvlaEdge")
	installCmd.Flags().VarP(installFlags.InstallMode, "mode", "m", fmt.Sprintf("Run mode for NuvlaEdge. Allowed values are: %v", installFlags.InstallMode.Allowed))

	// To run the NuvlaEdge after installing, we need the UUID of the NuvlaEdge
	installCmd.Flags().BoolVarP(&installFlags.Run, "start", "s", false, "Run NuvlaEdge after installation")
	installCmd.Flags().StringVar(&installFlags.Uuid, "uuid", "", "Run NuvlaEdge after installation")
	installCmd.MarkFlagsRequiredTogether("start", "uuid")

	installCmd.Flags().BoolVar(&installFlags.Detached, "detached", false, "Run NuvlaEdge in detached mode if the installation mode is process")
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

func installNuvlaEdge(flags InstallFlags) (*InstallPaths, error) {
	// 1. Get OS
	localOS := runtime.GOOS
	fmt.Println("Installing NuvlaEdge on OS: ", localOS)

	// 2. Get ARCH
	localArch := runtime.GOARCH
	fmt.Println("Installing NuvlaEdge on ARCH: ", localArch)

	// 3. Assert version to install
	version := getVersion(flags.Version)
	fmt.Println("Installing NuvlaEdge version: ", version)

	// 4. Build folder structure
	// If directory is provided, override default path and construct under the directory two folders:
	// - bin: for the binary
	// - config: for the configuration file
	// If directory is not provided, use default paths depending on the user permissions
	var paths InstallPaths
	if flags.InstallDir == "" {
		paths = NewDefaultInstallPaths(hasSudoPermissions())
	} else {
		// Here we assume that the user has permissions to write in the provided directory
		paths = NewFromBasePath(flags.InstallDir)
	}
	fmt.Println("Installing NuvlaEdge in: ", paths)
	paths.MakePaths()

	// 5. Compose download URLS both for the config file and the binary
	bUrl := fmt.Sprintf(NuvlaEdgeBinaryURL, version, localOS, localArch, version)
	fmt.Println("Downloading NuvlaEdge binary from: ", bUrl)
	// TODO: Need to be downloaded from the same tag version as the binary
	cUrl := NuvlaEdgeLatestConfTemplateURL
	fmt.Println("Downloading NuvlaEdge configuration template from: ", cUrl)

	// 6. Download files to temporal directory
	tempDir := "/tmp/nuvlaedge"
	_ = os.Mkdir(tempDir, os.ModePerm)
	err := downloadFile(bUrl, tempDir+"/nuvlaedge")
	if err != nil {
		fmt.Println("Error downloading binary")
		panic(err)
	}

	err = downloadFile(cUrl, tempDir+"/template.toml")
	if err != nil {
		fmt.Println("Error downloading configuration template")
		panic(err)
	}

	// 7. Move files to the installation directory
	err = os.Rename(tempDir+"/nuvlaedge", paths.BinaryPath+"/nuvlaedge")
	if err != nil {
		fmt.Println("Error moving binary")
		panic(err)
	}

	err = os.Rename(tempDir+"/template.toml", paths.ConfigPath+"/template.toml")
	if err != nil {
		fmt.Println("Error moving configuration template")
		panic(err)
	}

	return &paths, nil
}

func startNuvlaEdge(flags InstallFlags, paths *InstallPaths) error {
	// 1. Get the run mode
	mode := flags.InstallMode.Value
	fmt.Println("Starting NuvlaEdge as", mode)

	switch mode {
	case string(ProcessMode):
		return startNuvlaEdgeProcess(flags, paths)
	case string(ServiceMode):
		return startNuvlaEdgeService(flags, paths)
	default:
		return fmt.Errorf("invalid run mode: %s", mode)
	}
}

func startNuvlaEdgeProcess(flags InstallFlags, paths *InstallPaths) error {
	cmdPath := paths.BinaryPath + "/nuvlaedge"
	attr := []string{"-c", paths.ConfigPath + "/template.toml"}
	fmt.Println("Starting NuvlaEdge process with command: ", cmdPath, attr)

	cmd := exec.Command(cmdPath, attr...)
	var err error
	if flags.Detached {
		err = cmd.Run()
	} else {
		err = cmd.Start()
	}

	return err
}

func startNuvlaEdgeService(flags InstallFlags, paths *InstallPaths) error {
	return nil
}
