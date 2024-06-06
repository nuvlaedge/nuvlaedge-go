package types

type Installer interface {
	Install() error
	Start() error
	Stop() error
	Remove() error
	Status() string
	String() string
}
