package singleElevator

import (
	"Driver-go/modules/config"
	"Driver-go/modules/elevio"
	"fmt"
)

func FsmOnRequest(elev *Elevator, SetDoorCh chan<- bool, requestDone chan<- config.ButtonEvent, MotorDirectionCh chan<- config.MotorDirection) {
	switch elev.Behaviour {
	case config.EB_Idle:
		output := RequestsChooseDirection(elev)

		elev.Dirn = output.Dirn
		elev.Behaviour = output.Behaviour

		switch elev.Behaviour {
		case config.EB_DoorOpen:
			SetDoorCh <- true
			TimerStart(elev.Config.DoorOpenDuration_s, "door")
			elev = ClearRequestsAtCurrentFloor(elev, requestDone)

		case config.EB_Moving:
			MotorDirectionCh <- elev.Dirn
		case config.EB_Idle:
		}
	default:

	}
}

func FsmOnFloorArrival(newFloor int, elev *Elevator, requestDone chan<- config.ButtonEvent, MotorDirectionCh chan<- config.MotorDirection, SetDoorCh chan<- bool) {
	fmt.Printf("\n\nFloor arrival: %d\n", newFloor)
	elev.Floor = newFloor
	elevio.SetFloorIndicator(elev.Floor)

	switch elev.Behaviour {
	case config.EB_Moving:
		if RequestsShouldStop(elev) {
			MotorDirectionCh <- config.MotorDirection(0)
			SetDoorCh <- true
			elev = ClearRequestsAtCurrentFloor(elev, requestDone)
			TimerStart(elev.Config.DoorOpenDuration_s, "door")
			elev.Behaviour = config.EB_DoorOpen
		}
	}
}

func FsmOnDoorTimeout(elev *Elevator, requestDoneCh chan<- config.ButtonEvent, MotorDirectionCh chan<- config.MotorDirection, SetDoorCh chan<- bool) {
	fmt.Println("\n\nDoor timeout")
	output := RequestsChooseDirection(elev)
	elev.Dirn = output.Dirn
	elev.Behaviour = output.Behaviour
	switch elev.Behaviour {
	case config.EB_DoorOpen:
		TimerStart(elev.Config.DoorOpenDuration_s, "door")
		elev = ClearRequestsAtCurrentFloor(elev, requestDoneCh)
	case config.EB_Moving, config.EB_Idle:
		SetDoorCh <- false
		MotorDirectionCh <- elev.Dirn
	}
}
