package main

import (
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"nuvlaedge-go/cli"
	"os"
)

func main() {
	if err := nuvlaEdgeMain(context.Background()); err != nil {
		if !errors.Is(err, context.Canceled) {
			log.Infof("NuvlaEdge has been stopped with error: %s", err)
			os.Exit(0)
		}
		if !errors.Is(err, context.DeadlineExceeded) {
			log.Info("NuvlaEdge has reached a timeout")
			os.Exit(1)
		}

		log.Infof("NuvlaEdge stopped with unknown error %s", err)
		os.Exit(1)
	}
}

func nuvlaEdgeMain(ctx context.Context) error {
	rootCmd := &cobra.Command{
		Use:   "nuvlaedge",
		Short: "NuvlaEdge agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	cli.SetUpRootCommand(rootCmd)

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		return err
	}
	return nil
}
