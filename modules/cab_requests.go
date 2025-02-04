package cabrequests

import(
	"fmt"
    "elevator"
)

type DirnBehaviourPair int
const (
	Dirn dirn = iota
	ElevatorBehaviour behaviour
)

func requests_above(e Elevator)  bool {
    for f := e.floor + 1; f < N_FLOORS; f++ {
        for btn := 0; btn < N_BUTTONS; btn++ {
            if e.requests[f][btn] {
                return true 
            }
        }
    }
    return false 
}

func requests_below (e Elevator) bool {
    for f := 0; f < e.floor; f++{
        for btn := 0; btn < N_BUTTONS; btn++{
            if e.requests[f][btn]{
                return true
            }
        }
    }
    return false 
}

func requests_here(e Elevator) bool {
    for btn:= 0; btn < N_BUTTONS; btn ++{
        if e.requests[e.floor][btn]{
            return true 
        }
    }
    return false 
}

func requests_chooseDirection(e Elevator) DirnBehaviourPair {
	switch e.dirn {
	case D_Up:
		if requestsAbove(e) {
			return DirnBehaviourPair{D_Up, EB_Moving}
		} else if requestsHere(e) {
			return DirnBehaviourPair{D_Down, EB_DoorOpen}
		} else if requestsBelow(e) {
			return DirnBehaviourPair{D_Down, EB_Moving}
		} else {
			return DirnBehaviourPair{D_Stop, EB_Idle}
		}

	case D_Down:
		if requestsBelow(e) {
			return DirnBehaviourPair{D_Down, EB_Moving}
		} else if requestsHere(e) {
			return DirnBehaviourPair{D_Up, EB_DoorOpen}
		} else if requestsAbove(e) {
			return DirnBehaviourPair{D_Up, EB_Moving}
		} else {
			return DirnBehaviourPair{D_Stop, EB_Idle}
		}

	case D_Stop:
		// Stop case: Arbitrary check for up or down first
		if e.requestsHere() {
			return DirnBehaviourPair{D_Stop, EB_DoorOpen}
		} else if e.requestsAbove() {
			return DirnBehaviourPair{D_Up, EB_Moving}
		} else if e.requestsBelow() {
			return DirnBehaviourPair{D_Down, EB_Moving}
		} else {
			return DirnBehaviourPair{D_Stop, EB_Idle}
		}

	default:
		return DirnBehaviourPair{D_Stop, EB_Idle}
	}
}


func requests_shouldStop(e Elevator) bool {
	switch e.dirn {
	case D_Down:
		return e.requests[e.floor][B_HallDown] ||
			e.requests[e.floor][B_Cab] ||
			!e.requestsBelow()

	case D_Up:
		return e.requests[e.floor][B_HallUp] ||
			e.requests[e.floor][B_Cab] ||
			!e.requestsAbove()

	case D_Stop:
		fallthrough 

	default:
		return true
	}
}

func requests_shouldClearImmediately(e Elevator, btn_floor int, btn_type Button) bool {
    switch e.config.clearRequestVariant {
    case CV_All:
        return e.floor == btnFloor

    case CV_InDirn:
        return e.floor == btnFloor &&
            (e.dirn == D_Up && btnType == B_HallUp ||
                e.dirn == D_Down && btnType == B_HallDown ||
                e.dirn == D_Stop ||
                btnType == B_Cab)

    default:
        return false
    }
}

func clearRequestsAtCurrentFloor(e Elevator) Elevator{
    switch e.config.clearRequestVariant {
    case CV_All:
        for btn := 0; btn < N_BUTTONS; btn++ {
            e.requests[e.floor][btn] = 0
        }

    case CV_InDirn:
        e.requests[e.floor][B_Cab] = 0

        switch e.dirn {
        case D_Up:
            if !requestsAbove(e) && e.requests[e.floor][B_HallUp] == 0 {
                e.requests[e.floor][B_HallDown] = 0
            }
            e.requests[e.floor][B_HallUp] = 0

        case D_Down:
            if !requestsBelow(e) && e.requests[e.floor][B_HallDown] == 0 {
                e.requests[e.floor][B_HallUp] = 0
            }
            e.requests[e.floor][B_HallDown] = 0

        case D_Stop:
        default:
            e.requests[e.floor][B_HallUp] = 0
            e.requests[e.floor][B_HallDown] = 0
        }
    }
    return e
}