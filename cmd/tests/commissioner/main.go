package main

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/nuvla/api-client-go/clients"
	nuvlaTypes "github.com/nuvla/api-client-go/types"
	"nuvlaedge-go/commissioner"
	"nuvlaedge-go/telemetry"
	"nuvlaedge-go/types"
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
	creds := &nuvlaTypes.ApiKeyLogInParams{
		Key:    "credential/8718ba5e-7000-4862-b27d-8791c63d73e1",
		Secret: "aRDCvP.Nrzp2k.vpU3sv.SmcEnR.SDmGLS",
	}
	nuvlaedgeUuid := "nuvlabox/2cf4bff8-3e1b-411e-be50-f01310d8f884"
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

	com := commissioner.NewCommissioner(15, &comClient, commissionerCh)
	go com.Run(ctx)
	<-ctx.Done()
}
