package common

import (
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
	fmt.Printf("%s took %s\n", name, elapsed)
}
