package common

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
	"time"
)

func GenericErrorHandler(message string, err error) {
	if err != nil {
		e := fmt.Errorf("%s: %s", message, err)
		fmt.Println(e.Error())
	}
}

func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

func FileExistsAndNotEmpty(filePath string) bool {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return false
	}
	return fileInfo.Size() != 0
}

func ExecutionTime(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Infof("%s took %s\n", name, elapsed)
}

func CleanMap(m map[string]interface{}) {
	// Iterate over the map and remove keys with nil or empty string values
	for k, v := range m {
		switch v := v.(type) {
		case string:
			if v == "" {
				delete(m, k)
			}
		case map[string]interface{}:
			CleanMap(m[k].(map[string]interface{}))
		case nil:
			delete(m, k)
		}
	}
}

func LoadJsonFile(filePath string, data interface{}) error {
	if !FileExists(filePath) {
		return NewFileMissingError(filePath)
	}

	file, err := os.Open(filePath)
	if err != nil {
		log.Errorf("Error opening file %s: %s", filePath, err)
		return NewFileOpenError(filePath)
	}

	defer file.Close()

	decoder := json.NewDecoder(file)
	if err = decoder.Decode(&data); err != nil {
		log.Errorf("Error decoding file %s: %s", filePath, err)
		return err
	}
	return nil
}

func WriteContentToFile(content string, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		log.Errorf("Error creating file %s: %s", filePath, err)
		return err
	}
	_, err = file.Write([]byte(content))
	if err != nil {
		log.Errorf("Error writing to file %s: %s", filePath, err)
		return err
	}
	return nil
}

// SanitiseUUID returns the resource ID. If UUID starts with the resource name, means we already have the full ID.
// Else, we need to add the resource name to the UUID.
func SanitiseUUID(uuid, resourceName string) string {
	if uuid == "" {
		log.Infof("UUID for resource %s is empty", resourceName)
		return uuid
	}

	if strings.HasPrefix(uuid, resourceName) {
		return uuid
	}

	if strings.Contains(uuid, "/") {
		s := strings.Split(uuid, "/")
		if len(s) == 2 {
			log.Infof("UUID (%s) belongs to resource %s, not %s", s[0], s[1], resourceName)
		}
		return ""
	}

	return fmt.Sprintf("%s/%s", resourceName, uuid)
}
