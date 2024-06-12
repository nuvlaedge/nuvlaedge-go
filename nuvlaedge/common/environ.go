package common

import (
	"os"
	"strings"
)

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
