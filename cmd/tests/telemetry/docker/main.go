package main

import (
	"context"
	"encoding/json"
	"fmt"
	client "github.com/docker/docker/client"
	"nuvlaedge-go/types"
	"nuvlaedge-go/types/metrics"
	"nuvlaedge-go/workers/telemetry/monitor"
	"time"
)

func main() {
	docker, _ := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	ch := make(chan metrics.Metric)
	comCh := make(chan types.CommissionData)
	mon := monitor.NewDockerMonitor(docker, 10, ch, "https://nuvla.io", comCh)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(30)*time.Second)
	defer cancel()

	go func(dMon *monitor.DockerMonitor) {
		if err := dMon.Run(ctx); err != nil {
			fmt.Printf("Error running Docker monitor: %s\n", err)
		}
	}(mon)
	fmt.Printf("Monitor running: %t\n", mon.Running())
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Context done")
			fmt.Printf("Error: %v\n", ctx.Err())
			return
		case t := <-ch:
			fmt.Println("Got metric")
			printStruct(t)
		}
	}
}

func printStruct(s interface{}) {

	str, _ := json.MarshalIndent(s, "", "  ")
	fmt.Printf("%s\n", str)
}
