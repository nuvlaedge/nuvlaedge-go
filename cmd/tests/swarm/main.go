package main

import (
	"context"
	"fmt"
	docker "github.com/docker/docker/client"
	"strings"
)

func main() {
	dCli, err := docker.NewClientWithOpts(docker.FromEnv, docker.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	s, err := dCli.SwarmInspect(context.Background())

	if err != nil {
		if strings.Contains(err.Error(), "This node is not a swarm manager") {
			fmt.Println("This node is not a swarm manager or swarm is not initialized")
			return
		}
		panic(err)
	}

	fmt.Printf("Swarm ID: %s\n", s.ID)

}
