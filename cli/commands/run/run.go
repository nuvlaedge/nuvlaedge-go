package run

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	runFlags "nuvlaedge-go/cli/flags"
	"nuvlaedge-go/nuvlaedge"
	"nuvlaedge-go/types/settings"
)

func NewRunCommand() *cobra.Command {
	var opts settings.NuvlaEdgeSettings
	cmd := &cobra.Command{
		Use:   "run [OPTIONS]",
		Short: "Run a NuvlaEdge",
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Infof("Running NuvlaEdge")
			if err := runFlags.ParseSettings(cmd.Flags(), &opts); err != nil {
				log.Errorf("Error parsing settings: %s", err)
				return err
			}
			log.Infof("Running NuvlaEdge with settings: %v", opts)
			return nuvlaEdgeMain(cmd.Context(), &opts)
		},
	}

	flags := cmd.Flags()
	runFlags.AddRunFlags(flags)

	return cmd
}

func nuvlaEdgeMain(ctx context.Context, settings *settings.NuvlaEdgeSettings) error {
	log.Infof("Running NuvlaEdge with settings: %v", settings)

	ne, err := nuvlaedge.NewNuvlaEdge(ctx, settings)
	if err != nil {
		log.Errorf("Failed to create NuvlaEdge: %s", err)
		return err
	}

	if err := ne.Start(); err != nil {
		log.Errorf("Failed to start NuvlaEdge: %s", err)
		return err
	}

	<-ctx.Done()

	return nil
}
