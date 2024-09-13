package common

import (
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func SaveFile(fileName, workDir, content string) error {
	// Create the full file path
	filePath := filepath.Join(workDir, fileName)
	log.Info("Saving file: ", filePath)

	// Open the file for writing, creating it if it does not exist
	// #nosec
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write the content to the file
	_, err = file.WriteString(content)
	if err != nil {
		return err
	}

	return nil
}

func DownloadFile(url string, dest string) error {
	// #nosec
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return SaveFile(dest, "", string(b))
}
