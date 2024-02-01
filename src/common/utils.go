package common

import (
	"fmt"
	"os"
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

func ExecutionTime(start time.Time, name string) {
	elapsed := time.Since(start)
	fmt.Printf("%s took %s\n", name, elapsed)
}
