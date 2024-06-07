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
	"fmt"
	"github.com/spf13/cobra"
	"nuvlaedge-go/cli/common"
	"nuvlaedge-go/cli/installers"
	"nuvlaedge-go/cli/types"
	"os"
)

var installFlags types.InstallFlags

func ValidateInstallFlags() error {
	fmt.Println("Service mode: ", installFlags.Service)
	if installFlags.Service && !common.HasSudoPermissions() {
		return fmt.Errorf("service mode requires sudo permissions")
	}

	if installFlags.Docker && !common.IsDockerRunning() {
		return fmt.Errorf("docker mode requires Docker running")
	}

	if installFlags.Kubernetes && !common.IsKubernetesRunning() {
		return fmt.Errorf("kubernetes mode requires Kubernetes running")
	}

	return nil
}

// installCmd represents the installation command
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

	installFlags = types.InstallFlags{}

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

func GetInstaller(flags *types.InstallFlags) types.Installer {
	if flags.Process {
		return installers.NewProcessInstaller(flags.Uuid, flags.Version, flags)
	}
	if flags.Service {
		return installers.NewServiceInstaller(flags.Uuid, flags.Version, flags)
	}
	if flags.Kubernetes {
		return installers.NewHelmInstaller(flags.Uuid, flags.Version, flags)
	}

	rootCmd.Println("Invalid run mode, select proper mode...")
	return nil
}
