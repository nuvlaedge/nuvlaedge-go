package deployments

import (
	"strings"
)

func GetDeploymentProjectName(deploymentId string) string {
	return strings.Split(deploymentId, "/")[1]
}
