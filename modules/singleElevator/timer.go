package singleElevator

import (
	"Driver-go/modules/config"
	"time"
)

// Global variables
var (
	timerEndTime          time.Time
	timerActive           bool
	timerEndTimeAvailable time.Time
	timerActiveAvailable  bool
)

// Start the timer with a given duration in seconds
func TimerStart(duration float64, timerType string) {
	switch timerType {
	case "available":
		timerEndTimeAvailable = time.Now().Add(time.Duration(duration) * 3 * time.Second)
		timerActiveAvailable = true
	case "door":
		timerEndTime = time.Now().Add(time.Duration(duration) * time.Second)
		timerActive = true
	}
}

// Stop the timer
func TimerStop(timerType string) {
	switch timerType {
	case "door":
		timerActive = false
	case "available":
		timerActiveAvailable = false
	}
}

// Check if the timer has timed out
func TimerTimedOut(elev Elevator) bool {
	return timerActive && time.Now().After(timerEndTime) && !elev.ObstructionActive
}
func TimerTimedOutAvailable(elev *Elevator) bool {
	active_requests := false
	for i := 0; i < config.N_floor_const; i++ {
		for j := 0; j < config.N_buttons_const; j++ {
			if elev.Requests[i][j] {
				active_requests = true
				break
			}
		}
	}
	return timerActiveAvailable && time.Now().After(timerEndTimeAvailable) && active_requests
}

func PollAvailableTimeout(receiver chan<- bool, elev *Elevator) {
	prev := false
	for {
		time.Sleep(config.PollRate)
		v := TimerTimedOutAvailable(elev)
		if v != prev {
			TimerStop("available")
			receiver <- v
		}
		prev = v
	}
}

func PollTimeout(receiver chan<- bool, elev Elevator) {
	prev := false
	for {
		time.Sleep(config.PollRate)
		v := TimerTimedOut(elev)
		if v != prev {
			TimerStop("door")
			receiver <- v
		}
		prev = v
	}
}
