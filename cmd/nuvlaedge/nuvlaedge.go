package main

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"nuvlaedge-go/cli"
	"nuvlaedge-go/cmd/nuvlaedge/signals"
	"os"
	"os/signal"
)

func main() {
	onError := func(err error) {
		log.Errorf("Error: %s", err)
		os.Exit(1)
	}

	cmd := getNuvlaEdgeRootCommand()

	ctx, cancelNotify := signal.NotifyContext(context.Background(), signals.TerminationSignal...)
	defer cancelNotify()

	cmd.SetContext(ctx)

	if err := cmd.Execute(); err != nil {
		onError(err)
	}
}

func getNuvlaEdgeRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nuvlaedge",
		Short: "NuvlaEdge CLI",
		Long:  "NuvlaEdge CLi is a command line interface for NuvlaEdge",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	cli.SetUpRootCommand(cmd)

	return cmd
}
