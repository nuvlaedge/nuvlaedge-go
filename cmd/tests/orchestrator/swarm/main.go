package main

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"nuvlaedge-go/orchestrator"
	"nuvlaedge-go/types"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

var composeFilePath = "/Users/nacho/GolandProjects/refactor/nuvlaedge-go/cmd/tests/orchestrator/swarm/testdata/"

func main() {
	startTest()
	time.Sleep(3 * time.Second)
	stopTest()
}

func stopTest() {
	// Stop swarm service
	dClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	// Create orchestrator
	orch, err := orchestrator.NewSwarmOrchestrator(dClient)
	if err != nil {
		panic(err)
	}
	opts := &types.StopOpts{
		ProjectName: "TestSwarm",
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = orch.Stop(ctx, opts)
	if err != nil {
		panic(err)
	}
}

func startTest() {
	env := map[string]string{"EXPOSED_PORT": "8080", "DESTINATION": "127.0.0.1"}
	wd, _ := os.MkdirTemp("/tmp/", "swarm")
	defer os.RemoveAll(wd)

	fmt.Println("Working dir: ", wd)

	f, err := os.ReadFile(filepath.Join(composeFilePath, "docker-compose.yml"))
	if err != nil {
		panic(err)
	}
	res := ExpandEnvMapWithDefaults(string(f), env)

	err = os.WriteFile(filepath.Join(wd, "docker-compose.yml"), []byte(res), 0644)
	if err != nil {
		panic(err)
	}

	fmt.Println("File written")

	// Create swarm service
	dClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	// Create orchestrator
	orch, err := orchestrator.NewSwarmOrchestrator(dClient)
	if err != nil {
		panic(err)
	}
	opts := &types.StartOpts{
		CFiles:      []string{filepath.Join(wd, "docker-compose.yml")},
		ProjectName: "TestSwarm",
		WorkingDir:  wd,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = orch.Start(ctx, opts)
	if err != nil {
		panic(err)
	}

}

var envVarRegex = regexp.MustCompile(`\$\{(.+?)(?::-([^}]*))?}`)

func ExpandEnvMapWithDefaults(s string, envMap map[string]string) string {
	return envVarRegex.ReplaceAllStringFunc(s, func(m string) string {
		match := envVarRegex.FindStringSubmatch(m)
		if val, ok := envMap[match[1]]; ok {
			return val
		}
		if len(match) == 3 {
			return match[2] // return default value
		}
		return ""
	})
}
