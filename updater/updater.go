package updater

import (
	"nuvlaedge-go/types/options/command"
)

type Updater func(opts *command.UpdateCmdOptions) error

func GetUpdater() Updater {
	// This is a comment

	return UpdateWithCompose
}
