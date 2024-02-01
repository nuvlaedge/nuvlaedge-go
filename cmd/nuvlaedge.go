package main

import (
	log "github.com/sirupsen/logrus"
	"native-nuvlaedge/src/agent"
)

func testme() {
	ne := agent.New(agent.Config{})
	log.Infof("Starting NuvlaEdge agent")
	ne.Start()

	//testMe := coe.NewDockerCoe("")
	//log.Infof(testMe.String())

	log.Infof("Runnign nuvlaEdge agent")
	ne.Run()

}
