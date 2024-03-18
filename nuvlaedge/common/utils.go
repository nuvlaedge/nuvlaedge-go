package common

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"time"
)

var log = logrus.New()

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
