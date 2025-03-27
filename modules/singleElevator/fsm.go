package singleElevator

import (
	"Driver-go/modules/config"
	"Driver-go/modules/elevio"
	"fmt"
	"time"
)

// Called every time elev.Request is updated, acts according to how FSM is defined for a single elevator
func FsmOnRequest(elev Elevator, setDoorCh chan<- bool, requestDoneCh chan<- config.ButtonEvent, motorDirectionCh chan<- config.MotorDirection, timerDurationDoorCh chan<- time.Duration, timerDurationAvailableCh chan<- time.Duration) Elevator {
	switch elev.Behaviour {
	case config.EB_Idle:
		output := RequestsChooseDirection(elev)
		elev.Dirn = output.Dirn
		elev.Behaviour = output.Behaviour
		switch elev.Behaviour {
		case config.EB_DoorOpen:
			setDoorCh <- true
			fmt.Printf("onrequest")
			timerDurationDoorCh <- time.Duration(elev.DoorOpenDuration_s) * time.Second
		case config.EB_Moving:
			motorDirectionCh <- elev.Dirn
		case config.EB_Idle:
			timerDurationAvailableCh <- time.Duration(elev.AvailableDuration_s) * time.Second
		}
	default:
	}
	return elev
}

// Called every time elevator arrives at a floor, acts according to how FSM is defined for a single elevator
func FsmOnFloorArrival(newFloor int, elev Elevator, requestDoneCh chan<- config.ButtonEvent, motorDirectionCh chan<- config.MotorDirection, setDoorCh chan<- bool, timerDurationDoorCh chan<- time.Duration, timerDurationAvailableCh chan<- time.Duration) Elevator {
	elev.Floor = newFloor
	elevio.SetFloorIndicator(elev.Floor)

	switch elev.Behaviour {
	case config.EB_Moving:
		if RequestsShouldStop(elev) {
			motorDirectionCh <- config.MotorDirection(0)
			setDoorCh <- true
			timerDurationDoorCh <- time.Duration(elev.DoorOpenDuration_s) * time.Second
			elev.Behaviour = config.EB_DoorOpen
		}
	}
	elev.Available = true
	timerDurationAvailableCh <- time.Duration(elev.AvailableDuration_s) * time.Second // Resetting available deadline
	return elev
}

// Called every time the door times out, acts according to how FSM is defined for a single elevator
func FsmOnDoorTimeout(elev Elevator, requestDoneCh chan<- config.ButtonEvent, motorDirectionCh chan<- config.MotorDirection, setDoorCh chan<- bool, timerDurationDoorCh chan<- time.Duration) Elevator {
	switch elev.Behaviour {
	case config.EB_DoorOpen:
		elev = ClearRequestsAtCurrentFloor(elev, requestDoneCh)
		output := RequestsChooseDirection(elev)
		elev.Dirn = output.Dirn
		elev.Behaviour = output.Behaviour
		fmt.Println("behavior: ", elev.Behaviour)
		switch elev.Behaviour {
		case config.EB_DoorOpen:
			fmt.Printf("in door timeout")
			timerDurationDoorCh <- time.Duration(elev.DoorOpenDuration_s) * time.Second
		case config.EB_Moving, config.EB_Idle:
			setDoorCh <- false
			motorDirectionCh <- elev.Dirn
		}
	}
	return elev
}
