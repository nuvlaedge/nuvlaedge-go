package main

import (
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/src/agent"
	"nuvlaedge-go/src/coe"
)

func testme() {
	coeClient := coe.NewDockerCoe()
	ne := agent.NewAgent(agent.Config{}, coeClient)
	log.Infof("Starting NuvlaEdge agent")
	ne.Start()

	//testMe := coe.NewDockerCoe("")
	//log.Infof(testMe.String())

	log.Infof("Runnign nuvlaEdge agent")
	ne.Run()

}
