package executors

import (
	log "github.com/sirupsen/logrus"
)

type Host struct {
	ExecutorBase
}

func (h *Host) Reboot() error {
	// If we get to this point, we are running in the host and we can reboot. SUDO is already checked
	if stdout, err := ExecuteCommand("reboot"); err != nil {
		log.Errorf("Error executing reboot command: %s", stdout)
		return err
	}
	return nil
}

func (h *Host) InstallSSHKey(sshPub, user string) error {
	return nil
}

func (h *Host) RevokeSSKKey(sshkey string) error {
	return nil
}
