package singleElevator

import (
	"Driver-go/modules/config"
	"fmt"
	"time"
)

// Global variables
var (
	timerEndTimeDoor      time.Time
	timerActiveDoor       bool
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
		timerEndTimeDoor = time.Now().Add(time.Duration(duration) * time.Second)
		timerActiveDoor = true
		fmt.Printf("timer start")
	}
}

// Stop the timer
func TimerStop(timerType string) {
	switch timerType {
	case "door":
		timerActiveDoor = false
	case "available":
		timerActiveAvailable = false
	}
}

// Check if the door timer has timed out
func TimerTimedOutDoor(elev Elevator) bool {
	return timerActiveDoor && time.Now().After(timerEndTimeDoor) && !elev.ObstructionActive
}

// Check if the elevator available timer has timed out
func TimerTimedOutAvailable(elev Elevator) bool {
	activeRequests := false
	for i := 0; i < config.N_floor_const; i++ {
		for j := 0; j < config.N_buttons_const; j++ {
			if elev.Requests[i][j] {
				activeRequests = true
				break
			}
		}
	}
	return timerActiveAvailable && time.Now().After(timerEndTimeAvailable) && activeRequests
}

func PollAvailableTimeout(receiver chan<- bool, elev *Elevator) {
	prev := false
	for {
		time.Sleep(config.PollRate)
		v := TimerTimedOutAvailable(*elev)
		if v != prev {
			TimerStop("available")
			receiver <- v
		}
		prev = v
	}
}

func PollDoorTimeout(receiver chan<- bool, elev Elevator) {
	prev := false
	for {
		time.Sleep(config.PollRate)
		v := TimerTimedOutDoor(elev)
		if v != prev {
			TimerStop("door")
			receiver <- v
		}
		prev = v
	}
}
