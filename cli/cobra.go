package cli

import (
	"github.com/spf13/cobra"
	"nuvlaedge-go/cli/commands/register"
	"nuvlaedge-go/cli/commands/run"
	"nuvlaedge-go/cli/commands/update"
)

func SetUpRootCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(
		run.NewRunCommand(),
		register.NewRegisterCommand(),
		update.NewUpdateCommand(),
	)
}
