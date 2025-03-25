package singleElevator

import (
	"Driver-go/modules/config"
	"Driver-go/modules/elevio"
)

// Called every time elev.Request is updated, acts according to how FSM is defined for a single elevator 
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
			TimerStart(elev.Config.DoorOpenDuration_s, "available") //Starting available deadline 
		}
	default:
	}
}

// Called every time elevator arrives at a floor, acts according to how FSM is defined for a single elevator 
func FsmOnFloorArrival(newFloor int, elev *Elevator, requestDone chan<- config.ButtonEvent, MotorDirectionCh chan<- config.MotorDirection, SetDoorCh chan<- bool) {
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
	elev.Available = true
	TimerStart(elev.Config.DoorOpenDuration_s, "available") //Resetting available deadline 
}

// Called every time the door times out, acts according to how FSM is defined for a single elevator 
func FsmOnDoorTimeout(elev *Elevator, requestDoneCh chan<- config.ButtonEvent, MotorDirectionCh chan<- config.MotorDirection, SetDoorCh chan<- bool) {
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
