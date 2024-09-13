package main

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/nuvla/api-client-go/clients"
	nuvlaTypes "github.com/nuvla/api-client-go/types"
	"nuvlaedge-go/cmd/tests/util"
	"nuvlaedge-go/types"
	"nuvlaedge-go/workers"
	"nuvlaedge-go/workers/telemetry"
	"time"
)

type CustomNuvlaClient struct {
	*clients.NuvlaEdgeClient
}

func (c *CustomNuvlaClient) GetEndpoint() string {
	return c.NuvlaEdgeClient.SessionOpts.Endpoint
}

func main() {
	fmt.Println("Running commissioning tests")
	// Requires a commissioned NuvlaEdge credentials
	neInfo, err := util.GetNuvlaEdgeInfo("")
	if err != nil {
		panic(err)
	}
	creds := &nuvlaTypes.ApiKeyLogInParams{
		Key:    neInfo.ApiKey,
		Secret: neInfo.ApiSecret,
	}
	nuvlaedgeUuid := neInfo.NuvlaEdgeUUID
	nuvlaClient := clients.NewNuvlaEdgeClient(nuvlaedgeUuid, creds)
	nuvla := &CustomNuvlaClient{
		NuvlaEdgeClient: nuvlaClient,
	}

	if err := nuvla.UpdateResource(); err != nil {
		fmt.Printf("Error updating NuvlaEdge client: %v\n", err)
		return
	}

	nuvla.NuvlaEdgeStatusId = nuvlaTypes.NewNuvlaIDFromId(nuvla.GetNuvlaEdgeResource().NuvlaBoxStatus)
	dockerClient, _ := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	commissionerCh := make(chan types.CommissionData)
	tel := telemetry.NewTelemetry(10, nuvla, dockerClient, commissionerCh, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	go tel.Run(ctx)
	comClient := types.CommissionClient{
		NuvlaEdgeClient: nuvlaClient,
	}

	com := workers.NewCommissioner(15, &comClient, commissionerCh)
	go com.Run(ctx)
	<-ctx.Done()
}
