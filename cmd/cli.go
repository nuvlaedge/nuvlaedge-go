package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"nuvlaedge-go/cli/cmd"
)

func main() {
	go func() {
		fmt.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	cmd.Execute()
}
