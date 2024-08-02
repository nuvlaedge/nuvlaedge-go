package run

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	runFlags "nuvlaedge-go/cli/flags"
	"nuvlaedge-go/types/settings"
)

func NewRunCommand() *cobra.Command {
	var opts settings.NuvlaEdgeSettings
	cmd := &cobra.Command{
		Use:   "run [OPTIONS]",
		Short: "Run a NuvlaEdge",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := runFlags.ParseSettings(cmd.Flags(), &opts); err != nil {
				log.Errorf("Error parsing settings: %s", err)
				return err
			}

			log.Infof("Running NuvlaEdge with settings: %v", opts)
			return nil
		},
	}

	flags := cmd.Flags()
	runFlags.AddRunFlags(flags)

	return cmd
}
