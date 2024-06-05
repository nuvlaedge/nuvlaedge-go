package common

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

func HasSudoPermissions() bool {
	fmt.Printf("Checking sudo permissions... %d\n", os.Geteuid())
	return os.Geteuid() == 0
}

func IsDockerRunning() bool {
	_, err := os.Stat("/.dockerenv")
	return !os.IsNotExist(err)
}

func AmIInKubernetesRunning() bool {
	_, err := os.Stat("/var/run/secrets/kubernetes.io")
	return !os.IsNotExist(err)
}

func IsKubernetesRunning() bool {
	cmd := exec.Command("kubectl", "--kubeconfig=/etc/rancher/k3s/k3s.yaml", "cluster-info")

	o, err := cmd.Output()
	log.Infof("Output: %s", o)
	log.Infof("Error: %v", err)
	return err == nil
}

func DownloadFile(url string, path string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func InstallFileWithTmpDir(url string, filename string, targetDir string) error {
	// Expects the temporal directory to be already created
	tmpFile := filepath.Join(TempDir, filename)
	err := DownloadFile(url, tmpFile)
	if err != nil {
		fmt.Printf("Error downloading file %s from %s: %v\n", filename, url, err)
		return err
	}

	targetFile := filepath.Join(targetDir, filename)
	err = os.Rename(tmpFile, targetFile)
	if err != nil {
		fmt.Printf("Error moving file %s to %s: %v\n", tmpFile, targetFile, err)
		return err
	}

	return nil
}
