package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/nuvlaedge/common"
	"os"
	"runtime"
	"strings"
)

func initLogging() error {
	log.SetLevel(common.LogLevel)

	log.SetReportCaller(true)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:          true,
		TimestampFormat:        "02-01-2006 15:04:05",
		DisableLevelTruncation: true,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			return "", fmt.Sprintf("%s:%d", formatFilePath(f.File), f.Line)
		},
	})
	log.SetOutput(os.Stdout)

	return nil
}

func updateLoggingLevel(loggingSettings *common.LoggingSettings) {
	common.SetGlobalLogLevel(loggingSettings.Level)
	if loggingSettings.Debug {
		common.SetGlobalLogLevel("DEBUG")
	}
	log.SetLevel(common.LogLevel)
}

func formatFilePath(filePath string) string {
	arr := strings.Split(filePath, "/")
	return arr[len(arr)-1]
}
