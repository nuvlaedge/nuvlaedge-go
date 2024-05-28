package executors

// ComposeNotAvailableError is raised when trying to deploy an application but the compose file is not present in
// the deployment resource
type ComposeNotAvailableError struct {
	deploymentId string
	appType      string
}

func (e ComposeNotAvailableError) Error() string {
	return "Compose not available for deployment " + e.deploymentId + " with app type " + e.appType
}

func NewComposeNotAvailableError(deploymentId string, appType string) ComposeNotAvailableError {
	return ComposeNotAvailableError{deploymentId: deploymentId, appType: appType}
}
