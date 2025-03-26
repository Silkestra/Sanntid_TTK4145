package singleElevator

import (
	"Driver-go/modules/config"
	"Driver-go/modules/elevio"
	"fmt"
)

// Called every time elev.Request is updated, acts according to how FSM is defined for a single elevator
func FsmOnRequest(elev Elevator, SetDoorCh chan<- bool, requestDone chan<- config.ButtonEvent, MotorDirectionCh chan<- config.MotorDirection) Elevator {
	switch elev.Behaviour {
	case config.EB_Idle:
		output := RequestsChooseDirection(elev)
		elev.Dirn = output.Dirn
		elev.Behaviour = output.Behaviour

		switch elev.Behaviour {
		case config.EB_DoorOpen:
			SetDoorCh <- true
			fmt.Printf("onrequest")
			TimerStart(elev.DoorOpenDuration_s, "door")
			//elev = ClearRequestsAtCurrentFloor(elev, requestDone)
		case config.EB_Moving:
			MotorDirectionCh <- elev.Dirn
		case config.EB_Idle:
			//TimerStart(elev.DoorOpenDuration_s, "available") //Starting available deadline
			//fmt.Printf("available timer started")
		}
	default:
	}
	return elev
}

// Called every time elevator arrives at a floor, acts according to how FSM is defined for a single elevator
func FsmOnFloorArrival(newFloor int, elev Elevator, requestDone chan<- config.ButtonEvent, MotorDirectionCh chan<- config.MotorDirection, SetDoorCh chan<- bool) Elevator {
	elev.Floor = newFloor
	elevio.SetFloorIndicator(elev.Floor)

	switch elev.Behaviour {
	case config.EB_Moving:
		if RequestsShouldStop(elev) {
			MotorDirectionCh <- config.MotorDirection(0)
			SetDoorCh <- true
			//elev = ClearRequestsAtCurrentFloor(elev, requestDone)
			fmt.Printf("floorarrival")
			TimerStart(elev.DoorOpenDuration_s, "door")
			elev.Behaviour = config.EB_DoorOpen
		}
	}
	elev.Available = true
	TimerStart(elev.DoorOpenDuration_s, "available") //Resetting available deadline
	return elev
}

// Called every time the door times out, acts according to how FSM is defined for a single elevator
func FsmOnDoorTimeout(elev Elevator, requestDoneCh chan<- config.ButtonEvent, MotorDirectionCh chan<- config.MotorDirection, SetDoorCh chan<- bool) Elevator {
	elev = ClearRequestsAtCurrentFloor(elev, requestDoneCh)
	output := RequestsChooseDirection(elev)
	elev.Dirn = output.Dirn
	elev.Behaviour = output.Behaviour
	switch elev.Behaviour {
	case config.EB_DoorOpen:
		fmt.Printf("doortimout")
		TimerStart(elev.DoorOpenDuration_s, "door")
	case config.EB_Moving, config.EB_Idle:
		SetDoorCh <- false
		MotorDirectionCh <- elev.Dirn
	}
	return elev
}
