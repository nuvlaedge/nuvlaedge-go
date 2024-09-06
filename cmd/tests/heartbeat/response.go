package main

import (
	"github.com/nuvla/api-client-go/clients"
	"github.com/nuvla/api-client-go/types"
	"nuvlaedge-go/cmd/tests/util"
	"nuvlaedge-go/common"
)

func main() {
	// Requires a commissioned NuvlaEdge credentials
	neInfo, err := util.GetNuvlaEdgeInfo("")
	if err != nil {
		panic(err)
	}
	creds := &types.ApiKeyLogInParams{
		Key:    neInfo.ApiKey,
		Secret: neInfo.ApiSecret,
	}
	nuvlaedgeUuid := neInfo.NuvlaEdgeUUID

	cli := clients.NewNuvlaEdgeClient(nuvlaedgeUuid, creds)

	res, err := cli.Heartbeat()
	if err != nil {
		panic(err)
	}
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

	common.ProcessResponse(res, jobChan, confChan)
}
