package jobProcessor

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

type SudoRequiredError string

func (s SudoRequiredError) Error() string {
	return string(s)
}

func executeCommand(command string, args ...string) (string, error) {
	log.Infof("Executing command: %s with arguments %v", command, args)

	cmd := exec.Command(command, args...)
	output, err := cmd.Output()

	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			log.Errorf("Program %s exited with error code: %d", command, exitError.ExitCode())
			return string(exitError.Stderr), err
		}
		log.Errorf("Error executing command: %s", err)
		return string(output), err
	}

	return string(output), nil

}

// For possible future use
func executeCommandAsSuperUser(command string, args ...string) error {
	cmd := exec.Command("sudo", append([]string{command}, args...)...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return cmd.Run()
}

func isSuperUser() (bool, error) {
	if output, err := executeCommand("id", "-u"); err != nil {
		return false, err
	} else {
		return output == "0", nil
	}
}

func saveFileToDeploymentDir(deploymentUUID string, fileName string, content string) error {
	deploymentUUID = strings.Replace(deploymentUUID, "/", "_", -1)
	dirPath := filepath.Join("/tmp", deploymentUUID)
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		log.Error("Failed to create directory: %v", err)
		return err
	}

	filePath := filepath.Join(dirPath, fileName)
	err = os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		log.Errorf("Failed to write to file: %v", err)
		return err
	}

	log.Printf("Successfully wrote to file: %s", filePath)
	return nil
}
