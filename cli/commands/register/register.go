package register

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"nuvlaedge-go/cli/flags"
	"nuvlaedge-go/types/options/command"
)

func NewRegisterCommand() *cobra.Command {
	var opts command.RegisterCmdOptions

	cmd := &cobra.Command{
		Use:   "register",
		Short: "Create new NuvlaEdge resource in Nuvla",

		RunE: func(cmd *cobra.Command, args []string) error {
			if err := flags.ParseRegisterFlags(cmd, &opts); err != nil {
				return err
			}

			if err := ValidateOpts(&opts); err != nil {
				return err
			}

			return Register(cmd.Context(), &opts)
		},
	}

	flags.AddRegisterFlags(cmd)

	return cmd
}

func Register(ctx context.Context, opts *command.RegisterCmdOptions) error {
	// Create a new session with the provided opts
	client, err := newUserClient(opts)
	if err != nil {
		return err
	}

	conf, err := newNuvlaEdgeConfig(opts)
	if err != nil {
		return err
	}

	// Create a new NuvlaEdge resource with the provided configuration
	id, err := client.Add(ctx, "nuvlabox", conf)
	if err != nil {
		log.Errorf("Failed to create NuvlaEdge resource: %s", err)
		return err
	}
	log.Infof("NuvlaEdge resource created with ID: %s", id)
	return nil
}
