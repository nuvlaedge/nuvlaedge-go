package common

import (
	"errors"
	composeTypes "github.com/compose-spec/compose-go/v2/types"
	"os"
	"strings"
)

func ExportEnvs(mapping composeTypes.Mapping) error {
	var errs []error
	for k, v := range mapping {
		if err := os.Setenv(k, v); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func RemoveEnvs(mapping composeTypes.Mapping) error {
	var errs []error
	for k := range mapping {
		if err := os.Unsetenv(k); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func GetEnvironWithPrefix(prefixes ...string) []string {
	// Get all environment variables
	envs := os.Environ()

	var filteredEnvs []string
	for _, env := range envs {
		for _, filter := range prefixes {
			if strings.HasPrefix(env, filter) {
				filteredEnvs = append(filteredEnvs, env)
				break
			}
		}
	}

	return filteredEnvs
}
