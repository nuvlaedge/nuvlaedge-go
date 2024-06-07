package common

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"runtime"
	"strings"
)

// LogLevel Global log level propagator
var LogLevel log.Level

// TODO: Create a per package logger and allow for different log levels per package in NuvlaEdge file configuration

func init() {
	LogLevel = log.InfoLevel
}

func SetGlobalLogLevel(level string) {
	l, err := log.ParseLevel(level)
	if err != nil {
		log.Warnf("Invalid log level: %s. Setting level to default INFO", level)
		LogLevel = log.InfoLevel
	} else {
		LogLevel = l
	}
}

func InitLogging(logOpts LoggingSettings) {
	// Set nuvlaedge global lo level
	if logOpts.Debug {
		SetGlobalLogLevel("DEBUG")
	} else {
		SetGlobalLogLevel(logOpts.Level)
	}
	// Set logrus log level
	log.SetLevel(LogLevel)
	// Print the method calling the logger
	log.SetReportCaller(true)

	// Set formater
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:          true,
		TimestampFormat:        "02-01-2006 15:04:05",
		DisableLevelTruncation: true,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			return "", fmt.Sprintf("%s:%d", formatFilePath(f.File), f.Line)
		},
	})
	log.SetOutput(os.Stdout)
}

// formatFilePath formats the file path to only show the file name
func formatFilePath(filePath string) string {
	arr := strings.Split(filePath, "/")
	return arr[len(arr)-1]
}

type LoggingSettings struct {
	Debug         bool   `mapstructure:"debug" toml:"debug" json:"debug,omitempty"`
	Level         string `mapstructure:"level" toml:"level" json:"level,omitempty"`
	LogFile       string `mapstructure:"log-file" toml:"log-file" json:"log-file,omitempty"`
	LogPath       string `mapstructure:"log-path" toml:"log-path" json:"log-path,omitempty"`
	LogMaxSize    int    `mapstructure:"log-max-size" toml:"log-max-size" json:"log-max-size,omitempty"`
	LogMaxBackups int    `mapstructure:"log-max-backups" toml:"log-max-backups" json:"log-max-backups,omitempty"`
}
