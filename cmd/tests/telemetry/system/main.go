package main

import (
	"context"
	"encoding/json"
	"fmt"
	"nuvlaedge-go/telemetry/monitor"
	"nuvlaedge-go/types/metrics"
	"time"
)

func main() {
	ch := make(chan metrics.Metric)
	sysMon := monitor.NewSystemMonitor(10, ch)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(30)*time.Second)
	defer cancel()

	go func(sMon *monitor.SystemMonitor) {
		if err := sMon.Run(ctx); err != nil {
			panic(err)
		}
	}(sysMon)

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
