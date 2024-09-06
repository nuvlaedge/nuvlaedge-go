package main

import (
	"context"
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

	jobChan := make(chan string)
	confChan := make(chan string)
	go func() {
		for {
			select {
			case <-jobChan:
			case <-confChan:

			}
		}
	}()

	var hbWorker worker.Worker
	hbWorker = workers.NewHeartbeatWorker(cli, jobChan, confChan, 10)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	err := hbWorker.Start(*worker.NewDefaultWorkersConfig(), ctx)
	if err != nil {
		panic(err)
	}

	// Wait for the worker to finish
	<-ctx.Done()
}
