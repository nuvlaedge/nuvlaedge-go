package main

import (
	"context"
	"encoding/json"
	"fmt"
	"nuvlaedge-go/types/metrics"
	"nuvlaedge-go/workers/telemetry/monitor"
	"time"
)

func main() {
	ch := make(chan metrics.Metric)
	resMon := monitor.NewResourceMonitor(10, ch)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(30)*time.Second)
	defer cancel()

	go func(sMon *monitor.ResourceMonitor) {
		if err := sMon.Run(ctx); err != nil {
			panic(err)
		}
	}(resMon)

	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ch:
			printStruct(t)
		}

	}
}

func printStruct(s interface{}) {

	str, _ := json.MarshalIndent(s, "", "  ")
	fmt.Printf("%s\n", str)
}
