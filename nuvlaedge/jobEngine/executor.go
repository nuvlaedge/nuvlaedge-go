package jobEngine

import (
	log "github.com/sirupsen/logrus"
	"os/exec"
)

type ExecutorType string

const (
	DockerComposeExecutorType   ExecutorType = "docker-compose"
	DockerSwarmExecutorType     ExecutorType = "docker-swarm"
	DockerServiceExecutorType   ExecutorType = "docker-service"
	DockerContainerExecutorType ExecutorType = "docker-container"
	KubernetesExecutorType      ExecutorType = "kubernetes"
	UnixServiceExecutorType     ExecutorType = "unix-service"
	ShellCommandExecutorType    ExecutorType = "shell-command"
)

type Executor interface {
	RunAction() error
	GetExecutorType() ExecutorType
	Init() error
}

type ShellCommandExecutor struct {
	action Action

	command    string
	parameters []string
}

func NewShellCommandExecutor(action Action, command string, parameters []string) *ShellCommandExecutor {
	return &ShellCommandExecutor{
		action:     action,
		command:    command,
		parameters: parameters,
	}
}

func (sce *ShellCommandExecutor) RunAction() error {
	cmd := exec.Command(sce.command, sce.parameters...)
	output, err := cmd.Output()
	log.Infof("Executing command: %s", cmd.String())
	log.Infof("Output: %s", string(output))
	if err != nil {
		return err
	}
	log.Infof("Command %s executed successfully. Output: %s", sce.command, output)
	return nil
}

func (sce *ShellCommandExecutor) GetExecutorType() ExecutorType {
	return ShellCommandExecutorType
}

func (sce *ShellCommandExecutor) Init() error {
	return nil
}
