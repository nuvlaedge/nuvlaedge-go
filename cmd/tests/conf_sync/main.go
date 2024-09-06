package main

import (
	"context"
	"fmt"
	"github.com/nuvla/api-client-go/clients"
	cliTypes "github.com/nuvla/api-client-go/types"
	"nuvlaedge-go/cmd/tests/util"
	"nuvlaedge-go/types/worker"
	"nuvlaedge-go/workers"
	"time"
)

func main() {
	// Requires a commissioned NuvlaEdge credentials
	neInfo, err := util.GetNuvlaEdgeInfo("")
	if err != nil {
		panic(err)
	}
	creds := &cliTypes.ApiKeyLogInParams{
		Key:    neInfo.ApiKey,
		Secret: neInfo.ApiSecret,
	}
	nuvlaedgeUuid := neInfo.NuvlaEdgeUUID

	cli := clients.NewNuvlaEdgeClient(nuvlaedgeUuid, creds)

	lastUpdateChan := make(chan string)
	configChan := make(chan *worker.WorkerConfig)
	conf := &worker.WorkerConfig{}
	updater := workers.NewConfUpdaterWorker(cli, conf, lastUpdateChan, []chan *worker.WorkerConfig{configChan})
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	err := updater.Start(*worker.NewDefaultWorkersConfig(), ctx)
	if err != nil {
		panic(err)
	}

	dateTicker := time.NewTicker(5 * time.Second)
	go func() {
		for {
			select {
			case <-dateTicker.C:
				lastUpdateChan <- time.Now().Format(time.RFC3339)
			case <-ctx.Done():
				fmt.Println("Context done in new dates")
				return
			}
		}
	}()

	go func() {
		for {
			select {
			case c := <-configChan:
				fmt.Printf("Received configuration %v\n", c)
			case <-ctx.Done():
				fmt.Println("Context done in receiver")
				return
			}

		}
	}()

	<-ctx.Done()
	time.Sleep(100 * time.Millisecond)
	fmt.Println("Context done, finishing tests")
}
