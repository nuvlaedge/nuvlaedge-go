package common

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/common/constants"
	"os"
	"runtime"
	"strings"
)

var LogLevel = constants.LogLevel

func SetGlobalLogLevel(level string) {
	l, err := log.ParseLevel(level)
	if err != nil {
		log.Warnf("Invalid log level: %s. Setting level to default INFO", level)
		LogLevel = constants.LogLevel
	} else {
		LogLevel = l
	}
}

func InitLogging(level string, debug bool) {
	if debug {
		SetGlobalLogLevel("DEBUG")
	} else {
		SetGlobalLogLevel(level)
	}
	log.SetLevel(LogLevel)
	log.SetReportCaller(true)

	// Set formater
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:          true,
		TimestampFormat:        "02-01-2006 15:04:05",
		DisableLevelTruncation: true,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			return "", fmt.Sprintf("%s:%d:", formatFilePath(f.File), f.Line)
		},
	})
	log.SetOutput(os.Stdout)
}

// formatFilePath formats the file path to only show the file name
func formatFilePath(filePath string) string {
	arr := strings.Split(filePath, "/")
	return arr[len(arr)-1]
}
