package metrics

type InstallationParameters struct {
	ProjectName string   `json:"project-name,omitempty"`
	Environment []string `json:"environment,omitempty"`
	WorkingDir  string   `json:"working-dir,omitempty"`
	ConfigFiles []string `json:"config-files,omitempty"`
}

func (i InstallationParameters) WriteToStatus(status *NuvlaEdgeStatus) error {
	status.InstallationParameters = &i
	return nil
}
