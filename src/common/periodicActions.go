package common

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"time"
)

func RunTimedAction(action func() error, actionName string, period int) error {
	startTime := time.Now()
	err := action()

	if err != nil {
		return err
	}
	return WaitPeriodicAction(startTime, period, actionName)
}

func WaitPeriodicAction(startTime time.Time, expectedPeriod int, actionName string) error {
	processTime := time.Since(startTime)

	remainingTime := time.Duration(expectedPeriod)*time.Second - processTime

	if remainingTime.Seconds() < 0 {
		log.Warnf("%s update running behind schedule for %f", actionName, remainingTime.Seconds())
		err := fmt.Errorf("%s falling behind %f", actionName, remainingTime.Seconds())
		remainingTime = time.Duration(0)
		return err
	}

	log.Infof(
		"%s processing time: %v. Sleeping for: %f",
		actionName,
		processTime,
		remainingTime.Seconds())

	time.Sleep(remainingTime)
	return nil
}
