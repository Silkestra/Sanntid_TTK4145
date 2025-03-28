package singleElevator

import (
	"Driver-go/modules/config"
	"time"
)

// Global variables --> Should probably be changed to ensure no race conditions
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

	if timerActiveDoor && time.Now().After(timerEndTimeDoor) && !elev.ObstructionActive {
		timerActiveDoor = false
		return true
	}
	return false
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

func PollAvailableTimeout(receiver chan<- bool, updatedLocalElevatorCh <-chan Elevator) {
	prev := false
	for {
		currentElev := <-updatedLocalElevatorCh
		v := TimerTimedOutAvailable(currentElev)
		if v != prev {
			TimerStop("available")
			receiver <- v
		}
		prev = v
	}
}

func PollDoorTimeout(receiver chan<- bool, elev Elevator) {
    prev := false
    timeoutHandled := false 
    for {
        time.Sleep(config.PollRate)
        v := TimerTimedOutDoor(elev)

        if v != prev && !timeoutHandled {  
            timeoutHandled = true
            TimerStop("door")
            receiver <- v
        } else if !v {
            timeoutHandled = false 
        }
        
        prev = v
    }
}
