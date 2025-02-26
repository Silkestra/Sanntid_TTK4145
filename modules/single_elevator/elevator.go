package single_elevator

import (
	"Driver-go/"
)

// Defining elevator


type ElevatorBehaviour int

const (
	EB_Idle ElevatorBehaviour = iota
	EB_DoorOpen
	EB_Moving
	EB_Disconnected
)

var ObstructionActive bool

type ClearRequestVariant int

const (
	CV_All ClearRequestVariant = iota
	CV_InDirn
)

type Elevator struct {
	Floor int
	Dirn  elevio.MotorDirection
	//Requests [elevio.N_FLOORS][elevio.N_BUTTONS]int
	Requests  [4][3]bool
	Behaviour ElevatorBehaviour
	Config    Config
}

type Config struct {
	ClearRequestVariant ClearRequestVariant
	DoorOpenDuration_s  float64
}

func Eb_toString(eb ElevatorBehaviour) string {
	switch eb {
	case EB_Idle:
		return "idle"
	case EB_DoorOpen:
		return "doorOpen"
	case EB_Moving:
		return "moving"
	case EB_Disconnected:
		return "disconnected"
	default:
		return "disconnected"
	}
}

func Direction_toString(dirn MotorDirection) string {
	switch dirn {
	case MD_Up:
		return "up"
	case MD_Down:
		return "down"
	case MD_Stop:
		return "stop"
	default:
		return "disconnected"
	}
}



func Elevator_uninitialized() *Elevator {
	conf := Config{ClearRequestVariant: CV_InDirn, DoorOpenDuration_s: 3}
	p := Elevator{Floor: GetFloor(), Dirn: MD_Stop, Behaviour: EB_Idle, Config: conf}
	if p.Floor == -1 {
		SetMotorDirection(MD_Up)
		for {
			p.Floor = GetFloor()
			if p.Floor != -1 {
				SetMotorDirection(MD_Stop)
				break
			}
		}
	}
	return &p
}

