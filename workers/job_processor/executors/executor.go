package executors

type ExecutorName string

type Executor interface {
	GetName() ExecutorName
}

type ExecutorBase struct {
	Name ExecutorName
}

func (e *ExecutorBase) GetName() ExecutorName {
	return e.Name
}

const (
	ComposeExecutorName ExecutorName = "compose"
	HostExecutorName    ExecutorName = "host"
	StackExecutorName   ExecutorName = "stack"
	DockerExecutorName  ExecutorName = "docker"
)
