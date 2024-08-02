package common

import (
	composeTypes "github.com/compose-spec/compose-go/v2/types"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func Test_ExportEnvs(t *testing.T) {
	mapping := composeTypes.Mapping{
		"key": "value",
	}
	err := ExportEnvs(mapping)
	assert.NoError(t, err, "ExportEnvs should not return an error")
	assert.Equal(t, "value", os.Getenv("key"))

	mapping = composeTypes.Mapping{
		"": "value",
	}
	err = ExportEnvs(mapping)
	assert.Error(t, err, "ExportEnvs should return an error")
}

func Test_RemoveEnvs(t *testing.T) {
	mapping := composeTypes.Mapping{
		"key": "value",
	}
	err := ExportEnvs(mapping)
	assert.NoError(t, err, "ExportEnvs should not return an error")
	assert.Equal(t, "value", os.Getenv("key"))

	err = RemoveEnvs(mapping)
	assert.NoError(t, err, "RemoveEnvs should not return an error")
	assert.Empty(t, os.Getenv("key"))

}
