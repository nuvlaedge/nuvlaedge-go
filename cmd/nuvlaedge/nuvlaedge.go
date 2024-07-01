package main

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"nuvlaedge-go/cmd/nuvlaedge/signals"
	"nuvlaedge-go/nuvlaedge"
	"nuvlaedge-go/nuvlaedge/version"
	"os"
	"os/signal"
)

func main() {
	onError := func(err error) {
		log.Errorf("Error: %s", err)
		os.Exit(1)
	}

	if err := initLogging(); err != nil {
		onError(err)
	}

	// TODO: If we want to add a requirement check, it should go Here
	// if err := checkRequirements(); err != nil {
	//	onError(err)
	// }

	cmd, err := newNuvlaEdgeCommand()
	if err != nil {
		onError(err)
	}

	if err = cmd.Execute(); err != nil {
		onError(err)
	}
}

func nuvlaEdgeMain(settings *nuvlaedge.Settings) error {
	ctx, cancelNotify := signal.NotifyContext(context.Background(), signals.TerminationSignal...)
	defer cancelNotify()
	log.Infof("Starting nuvlaedge with ID %s", settings.Agent.NuvlaEdgeUUID)
	nuvlaEdge := nuvlaedge.NewNuvlaEdge(ctx, settings)

	if err := nuvlaEdge.Start(); err != nil {
		return err
	}

	// Run starts all the NuvlaEdge components routines
	errChan := make(chan error)
	go nuvlaEdge.Run(errChan)

	select {
	case <-ctx.Done():
		log.Info("NuvlaEdge has been stopped")
		return nil
	case err := <-errChan:
		log.Errorf("NuvlaEdge has been stopped with error: %s", err)
		return err
	}
}

func newNuvlaEdgeCommand() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "nuvlaedge",
		Short: "NuvlaEdge is the Nuvla agent for edge devices",
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Infof("Starting NuvlaEdge on version: %s", version.GetVersion())
			opts := GetSettings()

			updateLoggingLevel(&opts.Logging)

			return nuvlaEdgeMain(opts)
		},
	}

	SetupRootCommand(cmd)

	return cmd, nil
}
