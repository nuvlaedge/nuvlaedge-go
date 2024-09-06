package register

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/nuvla/api-client-go/clients"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/common/constants"
	"nuvlaedge-go/types/options/command"
	"time"
)

// newUserClient creates a new user client with the provided options and tries to login with the provided API keys
// Returns the user client or an error if login fails
func newUserClient(opts *command.RegisterCmdOptions) (*clients.UserClient, error) {
	if opts.Key == "" || opts.Secret == "" {
		log.Errorf("Nuvla API key and are required")
		return nil, nil
	}

	c := clients.NewUserClient(opts.Endpoint, opts.Insecure, false)
	err := c.LoginApiKeys(opts.Key, opts.Secret)
	if err != nil {
		log.Errorf("Failed to login with API key: %s", err)
		return nil, err
	}

	return c, nil
}

// newNuvlaEdgeConfig creates a new NuvlaEdge configuration based on the provided options
func newNuvlaEdgeConfig(opts *command.RegisterCmdOptions) (map[string]interface{}, error) {
	var res map[string]interface{}

	b, err := json.Marshal(opts)
	if err != nil {
		log.Errorf("Error marshaling NuvlaEdge configuration: %s", err)
		return nil, err
	}

	err = json.Unmarshal(b, &res)
	if err != nil {
		log.Errorf("Error unmarshaling NuvlaEdge configuration: %s", err)
		return nil, err
	}

	return res, nil
}

func ValidateOpts(opts *command.RegisterCmdOptions) error {
	if opts.Key == "" {
		return errors.New("Nuvla API key is required")
	}

	if opts.Secret == "" {
		return errors.New("Nuvla API secret is required")
	}

	if opts.VPNEnabled {
		opts.VPNServerId = constants.DefaultVPNId
	}

	if opts.Name == "" {
		opts.Name = generateDefaultName(opts.NamePrefix)
	}
	return nil
}

func generateDefaultName(prefix string) string {

	if prefix != "" {
		prefix = constants.DefaultPrefix
	}
	// Generate a short UUID
	fullUUID := uuid.New().String()
	shortUUID := fullUUID[:8] // Taking first 8 characters for brevity

	// Get current timestamp
	currentTime := time.Now()

	// Format timestamp
	formattedTime := currentTime.Format("02/01/2006|15:04:05")

	// Concatenate prefix, formatted timestamp, and short UUID
	return fmt.Sprintf("%s-%s|%s", prefix, formattedTime, shortUUID)
}
