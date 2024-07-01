//go:build pprof

package main

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
