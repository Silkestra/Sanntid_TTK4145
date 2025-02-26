package single_elevator

import (
	"Driver-go/modules/elevio"
)

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

func Direction_toString(dirn elevio.MotorDirection) string {
	switch dirn {
	case elevio.MD_Up:
		return "up"
	case elevio.MD_Down:
		return "down"
	case elevio.MD_Stop:
		return "stop"
	default:
		return "disconnected"
	}
}


func Elevator_uninitialized() *Elevator {
	conf := Config{ClearRequestVariant: CV_InDirn, DoorOpenDuration_s: 3}
	p := Elevator{Floor: elevio.GetFloor(), Dirn: elevio.MD_Stop, Behaviour: EB_Idle, Config: conf}
	if p.Floor == -1 {
		elevio.SetMotorDirection(elevio.MD_Up)
		for {
			p.Floor = elevio.GetFloor()
			if p.Floor != -1 {
				elevio.SetMotorDirection(elevio.MD_Stop)
				break
			}
		}
	}
	return &p
}