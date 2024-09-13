package main

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/nuvla/api-client-go/clients"
	"github.com/nuvla/api-client-go/types"
	"nuvlaedge-go/cmd/tests/util"
	"nuvlaedge-go/workers/telemetry"
)

//func init() {
//	p := os.Getenv("PPROF_LISTEN_PORT")
//	if p == "" {
//		p = "6060"
//	}
//
//	a := os.Getenv("PPROF_LISTEN_ADDR")
//	if a == "" {
//		a = "localhost"
//	}
//
//	listenAddr := fmt.Sprintf("%s:%s", a, p)
//	fmt.Printf("Starting pprof server on %s", listenAddr)
//
//	go func() {
//		_ = http.ListenAndServe(listenAddr, nil)
//	}()
//}

type CustomNuvlaClient struct {
	*clients.NuvlaEdgeClient
}

func (c *CustomNuvlaClient) GetEndpoint() string {
	return c.NuvlaEdgeClient.SessionOpts.Endpoint
}

func main() {
	fmt.Println("Running telemetry tests...")
	// Requires a commissioned NuvlaEdge credentials
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

	nuvla := &CustomNuvlaClient{
		NuvlaEdgeClient: clients.NewNuvlaEdgeClient(nuvlaedgeUuid, creds),
	}

	if err := nuvla.UpdateResource(); err != nil {
		fmt.Printf("Error updating NuvlaEdge client: %v\n", err)
		return
	}
	nuvla.NuvlaEdgeStatusId = types.NewNuvlaIDFromId(nuvla.GetNuvlaEdgeResource().NuvlaBoxStatus)
	dockerClient, _ := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	tel := telemetry.NewTelemetry(10, nuvla, dockerClient, nil, nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		err := tel.Run(ctx)
		if err != nil {
			fmt.Printf("Error running telemetry: %v\n", err)
		}
	}()

	<-ctx.Done()
	fmt.Printf("Context done %v\n", ctx.Err())
}
