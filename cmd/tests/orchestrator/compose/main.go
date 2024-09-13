package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/client"
	"net/http"
	_ "net/http/pprof"
	"nuvlaedge-go/orchestrator"
	"nuvlaedge-go/types"
	"os"
	"path/filepath"
	"time"
)

func init() {
	p := os.Getenv("PPROF_LISTEN_PORT")
	if p == "" {
		p = "6060"
	}

	a := os.Getenv("PPROF_LISTEN_ADDR")
	if a == "" {
		a = "localhost"
	}

	listenAddr := fmt.Sprintf("%s:%s", a, p)
	fmt.Printf("Starting pprof server on %s", listenAddr)

	go func() {
		_ = http.ListenAndServe(listenAddr, nil)
	}()
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	go func() {
		doStuff()
		<-ctx.Done()
	}()
	<-ctx.Done()
	fmt.Println("Sleeping after doing stuff")
	time.Sleep(1000 * time.Second)
}

func doStuff() {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	composeClient, err := orchestrator.NewComposeOrchestrator(dockerClient)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	wd := filepath.Join("/root/refactor/cmd/tests/orchestrator/compose/", "testdata")
	configFile := filepath.Join(wd, "docker-compose.yml")
	err = composeClient.Start(ctx, &types.StartOpts{
		CFiles:      []string{configFile},
		Env:         []string{"HTTP_BIND=0.0.0.0", "HTTP_PORT=8080", "EXAMPLE=example"},
		ProjectName: "project",
		WorkingDir:  wd})

	if err != nil {
		panic(err)
	}

	for {
		l, err := composeClient.List(ctx)
		if err != nil {
			panic(err)
		}
		for _, s := range l {
			b, _ := json.MarshalIndent(s, "", "  ")
			fmt.Println(string(b))
		}
		time.Sleep(5 * time.Second)
		if ctx.Err() != nil {
			break
		}
	}
}
