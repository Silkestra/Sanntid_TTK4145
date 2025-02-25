package single_elevator

import (
	"Driver-go/modules/hallrequests"
)

// Defining elevator
type HRAElevState = hallrequests.HRAElevState

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
	Dirn  MotorDirection
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

/*
func elevatorPrint(es Elevator) {
    fmt.Println("  +--------------------+")
    fmt.Printf(
        "  |floor = %-2d          |\n"+
            "  |dirn  = %-12s|\n"+
            "  |behav = %-12s|\n",
        es.Floor,
        elevio_dirn_toString(es.Dirn),
        eb_toString(es.Behaviour),
    )
    fmt.Println("  +--------------------+")
    fmt.Println("  |  | up  | dn  | cab |")
    for f := N_FLOORS - 1; f >= 0; f-- {
        fmt.Printf("  | %d", f)
        for btn := 0; btn < N_BUTTONS; btn++ {
            if (f == N_FLOORS-1 && btn == B_HallUp) ||
                (f == 0 && btn == B_HallDown) {
                fmt.Print("|     ")
            } else {
                if es.requests[f][btn] {
                    fmt.Print("|  #  ")
                } else {
                    fmt.Print("|  -  ")
                }
            }
        }
        fmt.Println("|")
    }
    fmt.Println("  +--------------------+")
} */

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

//Conversion functions:

//TODO: Fix this

func ElevatorToHRAElevState(elev Elevator) HRAElevState {
	switch elev.Behaviour {
	case EB_Idle, EB_Moving, EB_DoorOpen:
		var elev_cab []bool
		for i := 0; i < 4; i++ {
			elev_cab = append(elev_cab, elev.Requests[i][2])
		}
		return HRAElevState{
			Behavior:    Eb_toString(elev.Behaviour),
			Floor:       elev.Floor,
			Direction:   Direction_toString(elev.Dirn),
			CabRequests: elev_cab,
		}
	case EB_Disconnected:
		return HRAElevState{}
	default:
		return HRAElevState{}
	}
}
