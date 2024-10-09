package main

import (
	"context"
	"encoding/json"
	"nuvlaedge-go/workers/job_processor/executors/resource_handler"
	"time"
)

func main() {
	h, err := resource_handler.NewDockerResourceHandler(nil)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	actions := []resource_handler.ResourceAction{
		{
			Action:   "remove",
			Resource: "image",
			Id:       "image-id",
		},
		{
			Action:   "remove",
			Resource: "container",
			Id:       "container-id",
		},
		{
			Action:   "remove",
			Resource: "volume",
			Id:       "volume-id",
		},
		{
			Action:   "remove",
			Resource: "network",
			Id:       "network-id",
		},
		{
			Action:   "pull",
			Resource: "image",
			Id:       "image-id",
		},
	}

	responses := h.HandleActions(ctx, actions)
	b, _ := json.MarshalIndent(responses, "", "  ")

	println(string(b))
}
