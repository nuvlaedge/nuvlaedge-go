//go:build pprof

package nuvlaedge

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	_ "net/http/pprof"
	"os"
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
	log.Infof("Starting pprof server on %s", listenAddr)

	go func() {
		_ = http.ListenAndServe(listenAddr, nil)
	}()
}
