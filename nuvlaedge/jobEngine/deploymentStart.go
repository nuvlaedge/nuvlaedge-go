package jobEngine

import (
	"fmt"
	"github.com/nuvla/api-client-go/clients"
	"github.com/nuvla/api-client-go/types"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"path/filepath"
)

type DeploymentStartAction struct {
	ActionBase
	client *clients.NuvlaDeploymentClient

	DeploymentId *types.NuvlaID
}

func NewDeploymentStartAction(opts *ActionBaseOpts) *DeploymentStartAction {
	c := clients.NewNuvlaDeploymentClient(opts.jobResource.TargetResource.Href, opts.NuvlaClient)
	ds := &DeploymentStartAction{
		ActionBase: *NewActionBase(opts),
		client:     c,
	}

	err := c.UpdateResource()
	if err != nil {
		log.Errorf("Error updating resource: %s", err)
		return nil
	}
	c.PrintResource()
	ds.DeploymentId = types.NewNuvlaIDFromId(c.GetId())
	comp, ok := c.GetResource().Module["compatibility"]
	if !ok {
		log.Errorf("Error getting executor description from deployment")
		return nil
	}
	// TODO: Handler module field as a struct
	log.Infof("Starting deployment using: %s", comp.(string))

	return ds
}

func saveFileToJobDir(deploymentUUID string, fileName string, content string) error {
	dirPath := filepath.Join("/tmp", deploymentUUID)
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}

	filePath := filepath.Join(dirPath, fileName)
	err = os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		log.Fatalf("Failed to write to file: %v", err)
	}

	log.Printf("Successfully wrote to file: %s", filePath)
	return nil
}

func (ds *DeploymentStartAction) Execute() error {
	log.Infof("Executing deployment start action...")
	// Assuming docker compose for the moment
	module := ds.client.GetResource().Module
	content := module["content"].(map[string]interface{})
	compose, ok := content["docker-compose"]
	if !ok {
		log.Errorf("Error getting docker-compose file from deployment")
		return nil
	}
	log.Infof("Starting deployment using: %s", compose.(string))
	// Write string to file
	err := saveFileToJobDir(ds.DeploymentId.Uuid, "docker-compose.yml", compose.(string))
	if err != nil {
		log.Errorf("Error writing docker-compose file: %s", err)
		return err
	}

	// Start deployment

	//command := fmt.Sprintf("docker compose -f /tmp/%s/docker-compose.yml up -d", ds.DeploymentId.Uuid)
	command := []string{"compose", "-f", fmt.Sprintf("/tmp/%s/docker-compose.yml", ds.DeploymentId.Uuid), "up", "-d"}
	log.Infof("Executing command: docker %s", command)

	cmd := exec.Command("docker", command...)
	output, err := cmd.Output()
	log.Infof("Executing command: %s", cmd.String())
	log.Infof("Output: %s", string(output))
	if err != nil {
		return err
	}
	log.Infof("Command %s executed successfully. Output: %s", command, output)
	return nil

}

func (ds *DeploymentStartAction) GetActionType() ActionType {
	return DeploymentStartActionType
}

func (ds *DeploymentStartAction) Init(opts *ActionBaseOpts) error {
	return nil
}

func (ds *DeploymentStartAction) GetExecutors() []ExecutorType {
	return nil
}
