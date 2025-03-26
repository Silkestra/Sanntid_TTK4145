package singleElevator

import (
	"Driver-go/modules/config"
	"fmt"
)

type Button = config.ButtonType

type DirnBehaviourPair struct {
	Dirn      config.MotorDirection
	Behaviour config.ElevatorBehaviour
}

func requestsAbove(e Elevator) bool {
	for f := e.Floor + 1; f < config.N_floor_const; f++ {
		for btn := 0; btn < config.N_buttons_const; btn++ {
			if e.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func requestsBelow(e Elevator) bool {
	for f := 0; f < e.Floor; f++ {
		for btn := 0; btn < config.N_buttons_const; btn++ {
			if e.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func requestsHere(e Elevator) bool {
	for btn := 0; btn < config.N_buttons_const; btn++ {
		if e.Requests[e.Floor][btn] {
			return true
		}
	}
	return false
}

func RequestsChooseDirection(e Elevator) DirnBehaviourPair {
	switch e.Dirn {
	case config.MD_Up:
		if requestsAbove(e) {
			return DirnBehaviourPair{config.MD_Up, config.EB_Moving}
		} else if requestsHere(e) {
			return DirnBehaviourPair{config.MD_Down, config.EB_DoorOpen}
		} else if requestsBelow(e) {
			return DirnBehaviourPair{config.MD_Down, config.EB_Moving}
		} else {
			return DirnBehaviourPair{config.MD_Stop, config.EB_Idle}
		}

	case config.MD_Down:
		if requestsBelow(e) {
			return DirnBehaviourPair{config.MD_Down, config.EB_Moving}
		} else if requestsHere(e) {
			return DirnBehaviourPair{config.MD_Up, config.EB_DoorOpen}
		} else if requestsAbove(e) {
			return DirnBehaviourPair{config.MD_Up, config.EB_Moving}
		} else {
			return DirnBehaviourPair{config.MD_Stop, config.EB_Idle}
		}

	case config.MD_Stop:
		if requestsHere(e) {
			return DirnBehaviourPair{config.MD_Stop, config.EB_DoorOpen}
		} else if requestsAbove(e) {
			return DirnBehaviourPair{config.MD_Up, config.EB_Moving}
		} else if requestsBelow(e) {
			return DirnBehaviourPair{config.MD_Down, config.EB_Moving}
		} else {
			return DirnBehaviourPair{config.MD_Stop, config.EB_Idle}
		}

	default:
		return DirnBehaviourPair{config.MD_Stop, config.EB_Idle}
	}
}

func RequestsShouldStop(e Elevator) bool {
	switch e.Dirn {
	case config.MD_Down:
		return (e.Requests[e.Floor][config.BT_HallDown]) ||
			(e.Requests[e.Floor][config.BT_Cab]) ||
			!requestsBelow(e)
	case config.MD_Up:
		return (e.Requests[e.Floor][config.BT_HallUp]) ||
			(e.Requests[e.Floor][config.BT_Cab]) ||
			!requestsAbove(e)

	case config.MD_Stop:
		fallthrough

	default:
		return true
	}
}

func RequestsShouldClearImmediately(e Elevator, btn_floor int, btn_type Button) bool {
	return e.Floor == btn_floor &&
		(e.Dirn == config.MD_Up && btn_type == config.BT_HallUp ||
			e.Dirn == config.MD_Down && btn_type == config.BT_HallDown ||
			e.Dirn == config.MD_Stop ||
			btn_type == config.BT_Cab)
}

func ClearRequestsAtCurrentFloor(e Elevator, requestDone chan<- config.ButtonEvent) Elevator {
	e.Requests[e.Floor][config.BT_Cab] = false
	fmt.Println(config.ButtonEvent{Floor: e.Floor, Button: config.BT_Cab})
	requestDone <- config.ButtonEvent{Floor: e.Floor, Button: config.BT_Cab}

	switch e.Dirn {
	case config.MD_Up:
		// if !requestsAbove(e) && !e.Requests[e.Floor][config.BT_HallUp] {
		// 	e.Requests[e.Floor][config.BT_HallDown] = false
		// 	requestDone <- config.ButtonEvent{Floor: e.Floor, Button: config.BT_HallDown}

		// }

		e.Requests[e.Floor][config.BT_HallUp] = false
		requestDone <- config.ButtonEvent{Floor: e.Floor, Button: config.BT_HallUp}

	case config.MD_Down:
		// if !requestsBelow(e) && !e.Requests[e.Floor][config.BT_HallDown] {
		// 	e.Requests[e.Floor][config.BT_HallUp] = false
		// 	requestDone <- config.ButtonEvent{Floor: e.Floor, Button: config.BT_HallDown}
		// }
		e.Requests[e.Floor][config.BT_HallDown] = false
		requestDone <- config.ButtonEvent{Floor: e.Floor, Button: config.BT_HallDown}

	case config.MD_Stop:
		e.Requests[e.Floor][config.BT_HallUp] = false
		e.Requests[e.Floor][config.BT_HallDown] = false
		requestDone <- config.ButtonEvent{Floor: e.Floor, Button: config.BT_HallUp}
		requestDone <- config.ButtonEvent{Floor: e.Floor, Button: config.BT_HallDown}
	default:
	}
	return e
}
