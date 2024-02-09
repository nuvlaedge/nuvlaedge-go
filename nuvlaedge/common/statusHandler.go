package common

import (
	"fmt"
	"sync"
	"time"
)

type Status int

const (
	RUNNING Status = iota
	STARTING
	STARTED
	STOPPED
	FAILING
	FAILED
	UNKNOWN
)

func (s *Status) String() string {
	switch *s {
	case RUNNING:
		return "RUNNING"
	case STARTING:
		return "STARTING"
	case STARTED:
		return "STARTED"
	case STOPPED:
		return "STOPPED"
	case FAILED:
		return "FAILED"
	case FAILING:
		return "FAILING"
	default:
		return "UNKNOWN"
	}
}

type StatusReport struct {
	origin       string
	message      string
	moduleStatus Status
	date         time.Time
}

type StatusHandler struct {
	status      string
	statusNotes []string

	statusChan    chan *StatusReport
	exitChan      chan bool
	moduleReports map[string]*StatusReport

	reportsLock sync.Mutex
}

func NewStatusHandler(statusChan chan *StatusReport, exitChan chan bool) *StatusHandler {
	return &StatusHandler{
		statusNotes:   make([]string, 0),
		statusChan:    statusChan,
		exitChan:      exitChan,
		moduleReports: make(map[string]*StatusReport),
		reportsLock:   sync.Mutex{},
	}
}

func buildStatusMessage(report *StatusReport) string {
	return fmt.Sprintf("%s - %s: %s-%s",
		report.origin,
		time.Since(report.date),
		report.moduleStatus.String(),
		report.message)
}

func (s *StatusHandler) processStatus() {
	s.reportsLock.Lock()
	defer s.reportsLock.Unlock()

	tempStatus := "UNKNOWN"
	tempNotes := make([]string, 0)

	for mod, report := range s.moduleReports {
		log.Infof("Processing status report from %s", mod)

		if report.moduleStatus > STARTED {
			tempStatus = "DEGRADED"
		}

		if report.moduleStatus <= STARTED && tempStatus != "DEGRADED" {
			tempStatus = "OPERATIONAL"
		}
		tempNotes = append(tempNotes, buildStatusMessage(report))
	}
	s.status = tempStatus
	s.statusNotes = tempNotes
}

func (s *StatusHandler) GetStatus() (string, []string) {
	s.reportsLock.Lock()
	defer s.reportsLock.Unlock()

	return s.status, s.statusNotes
}

func (s *StatusHandler) Run() {
	for {
		select {
		case statusReport := <-s.statusChan:
			s.moduleReports[statusReport.origin] = statusReport

		case <-s.exitChan:
			log.Infof("Exiting status handler")
			return
		}
		s.processStatus()
	}
}
