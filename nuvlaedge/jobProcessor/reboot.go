package jobProcessor

import log "github.com/sirupsen/logrus"

type RebootAction struct {
}

func (r *RebootAction) ExecuteAction() error {
	if superUser, _ := isSuperUser(); !superUser {
		log.Warn("Reboot action requires super user privileges")
		return SudoRequiredError("Reboot action requires super user privileges")
	}
	log.Infof("Triggering reboot...")
	if stdout, err := executeCommand("reboot"); err != nil {
		log.Errorf("Error executing reboot command: %s", stdout)
		return err
	}
	return nil
}

func (r *RebootAction) GetActionType() ActionType {
	return RebootActionType
}

func (r *RebootAction) Init(opts ...ActionOptsFunc) error {
	return nil
}
