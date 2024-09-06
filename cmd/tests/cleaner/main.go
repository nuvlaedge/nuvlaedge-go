package main

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"nuvlaedge-go/workers"
	"time"
)

func main() {

	dCli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	// Create a new DockerCleaner
	dc := workers.NewDockerCleaner(dCli)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	fmt.Printf("Starting cleaner\n")
	if err := dc.CleanContainers(ctx); err != nil {
		panic(err)
	}
	fmt.Printf("Finished cleaning containers\n")
}
