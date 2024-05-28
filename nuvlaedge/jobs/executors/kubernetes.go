package executors

type Kubernetes struct {
	ExecutorBase
}

func (k *Kubernetes) Reboot() error {
	return nil
}

func (k *Kubernetes) InstallSSHKey(sshPub, user string) error {
	return nil
}

func (k *Kubernetes) RevokeSSKKey(sshkey string) error {
	return nil
}
