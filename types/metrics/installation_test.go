package metrics

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInstallationParameters_WriteToStatus(t *testing.T) {
	p := InstallationParameters{
		ProjectName: "project-name",
		Environment: []string{"environment"},
		WorkingDir:  "working-dir",
		ConfigFiles: []string{"config-files"},
	}
	status := &NuvlaEdgeStatus{}
	err := p.WriteToStatus(status)
	assert.NoErrorf(t, err, "error writing installation parameters to status")
	assert.Equal(t, p, *status.InstallationParameters, "installation parameters not written to status")

}
