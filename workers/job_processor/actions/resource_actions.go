package actions

import (
	"context"
	"encoding/json"
	"errors"
	"nuvlaedge-go/workers/job_processor/executors"
	"nuvlaedge-go/workers/job_processor/executors/resource_handler"
)

type COEResourceActions struct {
	ActionBase

	actions       ResourceActionsPayload
	results       []resource_handler.ResourceActionResponse
	dockerHandler *resource_handler.DockerResourceHandler
}

func (c *COEResourceActions) Init(ctx context.Context, optsFn ...ActionOptsFn) error {
	opts := GetActionOpts(optsFn...)

	if opts.JobResource == nil || opts.Client == nil {
		return errors.New("jobs resource or client not available")
	}

	// Convert payload into bytes
	b, err := json.Marshal(opts.JobResource)
	if err != nil {
		return err
	}

	// Unmarshal bytes into map
	err = json.Unmarshal(b, &c.actions)
	if err != nil {
		return err
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
		c.results = c.dockerHandler.HandleActions(ctx, c.actions.Docker)
	}

	return nil
}

func (c *COEResourceActions) GetOutput() string {
	resultString := ""

	for _, result := range c.results {
		resultString += result.Message + "\n"
	}

	return resultString
}

type ResourceActionsPayload struct {
	Docker     []resource_handler.ResourceAction `json:"docker"`
	Kubernetes []resource_handler.ResourceAction `json:"kubernetes"`
}
