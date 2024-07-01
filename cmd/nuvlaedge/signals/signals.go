package signals

import (
	"golang.org/x/sys/unix"
	"os"
)

var TerminationSignal = []os.Signal{unix.SIGINT, unix.SIGTERM}
