package common

import (
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetGlobalLogLevel(t *testing.T) {
	SetGlobalLogLevel("debug")
	assert.Equal(t, LogLevel, log.DebugLevel)

	SetGlobalLogLevel("notALevel")
	assert.Equal(t, LogLevel, log.InfoLevel)
}

func TestInitLogging(t *testing.T) {
	InitLogging("debug", false)
	assert.Equal(t, log.GetLevel(), log.DebugLevel)

	InitLogging("info", true)
	assert.Equal(t, log.GetLevel(), log.DebugLevel)
}
