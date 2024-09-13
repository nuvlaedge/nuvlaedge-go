package main

import (
	"context"
	"fmt"
	"github.com/spf13/pflag"
	"nuvlaedge-go/cli/flags"
	"nuvlaedge-go/nuvlaedge"
	"nuvlaedge-go/types/settings"
)

func configureSettings() *settings.NuvlaEdgeSettings {
	fls := &pflag.FlagSet{}

	conf := &settings.NuvlaEdgeSettings{}

	err := flags.ParseSettings(fls, conf)
	if err != nil {
		panic(err)
	}

	conf.DBPPath = "/tmp/nuvlaedge"
	conf.NuvlaEdgeUUID = "cc02a9de-e799-4c50-a301-8b943aa66362"
	return conf
}
func main() {
	ctx, cancel := context.WithCancel(context.Background()) //, 360*time.Second)
	defer cancel()

	conf := configureSettings()

	ne, err := nuvlaedge.NewNuvlaEdge(ctx, conf)
	if err != nil {
		panic(err)
	}
	fmt.Printf("NuvlaEdge: %v\n", ne)

	if err = ne.Start(); err != nil {
		panic(err)
	}

	<-ctx.Done()
}
