package single_elevator

import (
	"fmt"
)

type Button = ButtonType

type DirnBehaviourPair struct {
	Dirn      MotorDirection
	Behaviour ElevatorBehaviour
}

func requests_above(e *Elevator) bool {
	for f := e.Floor + 1; f < N_FLOORS; f++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if e.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func requests_below(e *Elevator) bool {
	for f := 0; f < e.Floor; f++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if e.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func requests_here(e *Elevator) bool {
	for btn := 0; btn < N_BUTTONS; btn++ {
		if e.Requests[e.Floor][btn] {
			return true
		}
	}
	return false
}

func Requests_chooseDirection(e *Elevator) DirnBehaviourPair {
	switch e.Dirn {
	case MD_Up:
		if requests_above(e) {
			return DirnBehaviourPair{MD_Up, EB_Moving}
		} else if requests_here(e) {
			return DirnBehaviourPair{MD_Down, EB_DoorOpen}
		} else if requests_below(e) {
			return DirnBehaviourPair{MD_Down, EB_Moving}
		} else {
			return DirnBehaviourPair{MD_Stop, EB_Idle}
		}

	case MD_Down:
		if requests_below(e) {
			return DirnBehaviourPair{MD_Down, EB_Moving}
		} else if requests_here(e) {
			return DirnBehaviourPair{MD_Up, EB_DoorOpen}
		} else if requests_above(e) {
			return DirnBehaviourPair{MD_Up, EB_Moving}
		} else {
			return DirnBehaviourPair{MD_Stop, EB_Idle}
		}

	case MD_Stop:
		// Stop case: Arbitrary check for up or down first
		if requests_here(e) {
			return DirnBehaviourPair{MD_Stop, EB_DoorOpen}
		} else if requests_above(e) {
			return DirnBehaviourPair{MD_Up, EB_Moving}
		} else if requests_below(e) {
			return DirnBehaviourPair{MD_Down, EB_Moving}
		} else {
			return DirnBehaviourPair{MD_Stop, EB_Idle}
		}

	default:
		return DirnBehaviourPair{MD_Stop, EB_Idle}
	}
}

func Requests_shouldStop(e *Elevator) bool {
	switch e.Dirn {
	case MD_Down:
		fmt.Printf("check1før")
		if (e.Requests[e.Floor][BT_HallDown]) ||
			(e.Requests[e.Floor][BT_Cab]) ||
			!requests_below(e) {
			fmt.Printf("check1")
		}
		return (e.Requests[e.Floor][BT_HallDown]) ||
			(e.Requests[e.Floor][BT_Cab]) ||
			!requests_below(e)
	case MD_Up:
		fmt.Printf("check2før")
		if (e.Requests[e.Floor][BT_HallUp]) ||
			(e.Requests[e.Floor][BT_Cab]) ||
			!requests_above(e) {
			fmt.Printf("check2")
		}
		return (e.Requests[e.Floor][BT_HallUp]) ||
			(e.Requests[e.Floor][BT_Cab]) ||
			!requests_above(e)

	case MD_Stop:
		fmt.Printf("check3")
		fallthrough

	default:
		fmt.Printf("check4")
		return true
	}
}

func Requests_shouldClearImmediately(e *Elevator, btn_floor int, btn_type Button) bool {
	switch e.Config.ClearRequestVariant {
	case CV_All:
		return e.Floor == btn_floor

	case CV_InDirn:
		return e.Floor == btn_floor &&
			(e.Dirn == MD_Up && btn_type == BT_HallUp ||
				e.Dirn == MD_Down && btn_type == BT_HallDown ||
				e.Dirn == MD_Stop ||
				btn_type == BT_Cab)

	default:
		return false
	}
}

func ClearRequestsAtCurrentFloor(e *Elevator) *Elevator {
	switch e.Config.ClearRequestVariant {
	case CV_All:
		for btn := 0; btn < N_BUTTONS; btn++ {
			e.Requests[e.Floor][btn] = false
		}

	case CV_InDirn:
		e.Requests[e.Floor][BT_Cab] = false

		switch e.Dirn {
		case MD_Up:
			if !requests_above(e) && e.Requests[e.Floor][BT_HallUp] == false {
				e.Requests[e.Floor][BT_HallDown] = false
			}
			e.Requests[e.Floor][BT_HallUp] = false

		case MD_Down:
			if !requests_below(e) && e.Requests[e.Floor][BT_HallDown] == false {
				e.Requests[e.Floor][BT_HallUp] = false
			}
			e.Requests[e.Floor][BT_HallDown] = false

		case MD_Stop:
		default:
			e.Requests[e.Floor][BT_HallUp] = false
			e.Requests[e.Floor][BT_HallDown] = false
		}
	}
	return e
}
