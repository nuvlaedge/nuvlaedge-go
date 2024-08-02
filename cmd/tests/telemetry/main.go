package main

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/nuvla/api-client-go/clients"
	"github.com/nuvla/api-client-go/types"
	//_ "net/http/pprof"
	"nuvlaedge-go/telemetry"
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
	creds := &types.ApiKeyLogInParams{
		Key:    "credential/8718ba5e-7000-4862-b27d-8791c63d73e1",
		Secret: "aRDCvP.Nrzp2k.vpU3sv.SmcEnR.SDmGLS",
	}
	nuvlaedgeUuid := "nuvlabox/2cf4bff8-3e1b-411e-be50-f01310d8f884"
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
