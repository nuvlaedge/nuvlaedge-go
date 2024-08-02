package cli

import (
	"github.com/spf13/cobra"
	"nuvlaedge-go/cli/commands/run"
)

func SetUpRootCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(
		run.NewRunCommand(),
	)
}
