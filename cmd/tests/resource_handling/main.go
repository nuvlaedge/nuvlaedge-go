package main

import (
	"context"
	"encoding/json"
	"fmt"
	handler "nuvlaedge-go/workers/job_processor/executors/resource_handler"
	"time"
)

func main() {
	fmt.Println("Resource handling test")

	h, err := handler.NewDockerResourceHandler()
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	actions := []handler.ResourceAction{
		{
			Action:   "pull",
			Resource: "image",
			Id:       "alpine:latest",
		},
		{
			Action:   "remove",
			Resource: "image",
			Id:       "alpine:latest",
		},
	}

	fmt.Println("Handling actions")
	responses, err := h.HandlerActions(ctx, actions)
	if err != nil {
		panic(err)
	}

	fmt.Println("Responses:")
	for _, r := range responses {
		//fmt.Println("Response:", r)
		b, _ := json.MarshalIndent(r, "", "  ")
		fmt.Printf("%s\n", string(b))
	}
}
