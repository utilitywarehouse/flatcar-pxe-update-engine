package main

import (
	"fmt"
)

const (
	updateStatusIdle              = "UPDATE_STATUS_IDLE"
	updateStatusUpdatedNeedReboot = "UPDATE_STATUS_UPDATED_NEED_REBOOT"
)

type status struct {
	lastCheckedTime  int64
	progress         float64
	currentOperation string
	newVersion       string
	newSize          int64
}

func newStatus() *status {
	return &status{
		lastCheckedTime:  0,
		progress:         0,
		currentOperation: updateStatusIdle,
		newVersion:       "0.0.0",
		newSize:          0,
	}
}

func (s *status) String() string {
	return fmt.Sprintf("LastCheckedTime=%v Progress=%v CurrentOperation=%q NewVersion=%v NewSize=%v",
		s.lastCheckedTime,
		s.progress,
		s.currentOperation,
		s.newVersion,
		s.newSize,
	)
}
