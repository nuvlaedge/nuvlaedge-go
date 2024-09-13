package update

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"nuvlaedge-go/cli/flags"
	"nuvlaedge-go/types/options/command"
	"nuvlaedge-go/updater"
)

func NewUpdateCommand() *cobra.Command {
	var opts command.UpdateCmdOptions

	cmd := &cobra.Command{
		Use:   "update",
		Short: "NuvlaEdge Update management command",

		RunE: func(cmd *cobra.Command, args []string) error {
			if err := flags.ParseUpdateFlags(cmd.Flags(), &opts); err != nil {
				log.Errorf("Error parsing update flags: %s", err)
				return err
			}
			log.Info("Triggering update")
			return updateMain(&opts)
		},
	}

	flags.AddUpdateFlags(cmd)

	return cmd
}

func updateMain(opts *command.UpdateCmdOptions) error {
	log.Infof("Triggering update")
	updaterFunc := updater.GetUpdater()

	if err := updaterFunc(opts); err != nil {
		log.Errorf("Error updating NuvlaEdge: %s", err)
		return err
	}

	return nil
}
