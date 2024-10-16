package actions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/workers/job_processor/executors"
	"nuvlaedge-go/workers/job_processor/executors/resource_handler"
	"strings"
)

type COEResourceActions struct {
	ActionBase

	actions       ResourceActionsPayload
	results       ResourceActionsResult
	dockerHandler *resource_handler.DockerResourceHandler
}

func (c *COEResourceActions) Init(_ context.Context, optsFn ...ActionOptsFn) error {
	opts := GetActionOpts(optsFn...)

	if opts.JobResource == nil || opts.Client == nil {
		return errors.New("jobs resource or client not available")
	}

	if opts.JobResource.Payload == "" {
		return errors.New("job resource payload not available")
	}

	data := strings.ReplaceAll(opts.JobResource.Payload, "\\", "")

	c.actions = ResourceActionsPayload{}
	// Unmarshal bytes into map
	err := json.Unmarshal([]byte(data), &c.actions)
	if err != nil {
		log.Errorf("Error unmarshaling payload: %s", opts.JobResource.Payload)
		return err
	}

	if err := c.assertExecutor(); err != nil {
		return fmt.Errorf("error asserting executor: %w", err)
	}

	return nil
}

func (c *COEResourceActions) GetExecutorName() executors.ExecutorName {
	return "resource_handler"
}

func (c *COEResourceActions) assertExecutor() error {
	var err error
	if len(c.actions.Docker) > 0 {
		c.dockerHandler, err = resource_handler.NewDockerResourceHandler(nil)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *COEResourceActions) ExecuteAction(ctx context.Context) error {
	if len(c.actions.Docker) > 0 {
		c.results.Docker = c.dockerHandler.HandleActions(ctx, c.actions.Docker)
	}

	// Kubernetes actions not supported
	if len(c.actions.Kubernetes) > 0 {
		c.results.Kubernetes = []resource_handler.ResourceActionResponse{
			{
				Success:    false,
				Message:    "Kubernetes actions not supported",
				ReturnCode: 500,
			},
		}
	}

	return nil
}

func (c *COEResourceActions) GetOutput() string {

	b, err := json.MarshalIndent(c.results, "", "    ")
	if err != nil {
		log.Errorf("Error marshaling results: %s", err)
		return "Error unmarshalling the results"
	}

	return string(b)
}

type ResourceActionsPayload struct {
	Docker     []resource_handler.ResourceAction `json:"docker,omitempty"`
	Kubernetes []resource_handler.ResourceAction `json:"kubernetes,omitempty"`
}

type ResourceActionsResult struct {
	Docker     []resource_handler.ResourceActionResponse `json:"docker,omitempty"`
	Kubernetes []resource_handler.ResourceActionResponse `json:"kubernetes,omitempty"`
}
