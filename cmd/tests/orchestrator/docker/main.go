package main

import (
	"fmt"
	"github.com/docker/docker/client"
	"nuvlaedge-go/orchestrator"
	"time"
)

import (
	log "github.com/sirupsen/logrus"
	"net/http"
	_ "net/http/pprof"
)

func init() {
	log.Info("Starting pprof server on localhost:6060")
	go func() {
		_ = http.ListenAndServe("localhost:6060", nil)
	}()
}

func main() {
	fmt.Printf("Creating Client\n")
	dCli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	time.Sleep(10 * time.Second)
	createSwarm(dCli)
	createCompose(dCli)

	//fmt.Printf("Creating both orchestrators\n")
	//cOrch, err := orchestrator.NewComposeOrchestrator(dCli)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Printf("Compose orchestrator: %v\n", cOrch)
	//
	//sOrch, err := orchestrator.NewSwarmOrchestrator(dCli)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Printf("Swarm orchestrator: %v\n", sOrch)
	//fmt.Printf("CLI client: %v\n", dCli)
	//time.Sleep(240 * time.Second)
}

func createCompose(dCli client.APIClient) {
	fmt.Printf("Creating swarm\n")
	cOrch, err := orchestrator.NewComposeOrchestrator(dCli)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Compose orchestrator: %v\n", cOrch)
	time.Sleep(20 * time.Second)
	fmt.Printf("Exit swarm creation")
}

func createSwarm(dCli client.APIClient) {
	fmt.Printf("Creating compose\n")
	sOrch, err := orchestrator.NewSwarmOrchestrator(dCli)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Swarm orchestrator: %v\n", sOrch)
	time.Sleep(20 * time.Second)
	fmt.Printf("Exit compose creation")
}
