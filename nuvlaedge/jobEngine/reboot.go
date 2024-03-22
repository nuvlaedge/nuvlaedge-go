package jobEngine

import (
	log "github.com/sirupsen/logrus"
	"os/exec"
	"strings"
)

type SudoRequiredError string

func (sre SudoRequiredError) Error() string {
	return string(sre)

}

type RebootAction struct {
	ActionBase

	executor Executor
}

func NewRebootAction(opts *ActionBaseOpts) *RebootAction {
	return &RebootAction{
		ActionBase: *NewActionBase(opts),
		executor:   &ShellCommandExecutor{},
	}
}

func isSuperUser() (bool, error) {
	cmd := exec.Command("id", "-u")
	output, err := cmd.Output()

	if err != nil {
		return false, err
	}

	// The root user's ID is 0
	return strings.TrimSpace(string(output)) == "0", nil
}

func (ra *RebootAction) Execute() error {
	log.Infof("Executing reboot action...")
	if superUser, _ := isSuperUser(); !superUser {
		log.Warn("Reboot action requires super user privileges")
		return SudoRequiredError("Reboot action requires super user privileges")
	}
	log.Infof("Triggering reboot...")
	ra.executor = NewShellCommandExecutor(ra, "reboot", nil)
	if err := ra.executor.RunAction(); err != nil {
		return err
	}

	return nil
}

func (ra *RebootAction) GetActionType() ActionType {
	return RebootActionType
}

func (ra *RebootAction) Init(opts *ActionBaseOpts) error {
	return nil
}

func (ra *RebootAction) GetExecutor() Executor {
	return ra.executor
}
