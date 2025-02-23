package elevator

import "Driver-go/modules/elevio"

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

type HallRequestStates int

const (
	Uncomfirmed HallRequestStates = iota
	Confirmed
	Done
	Unknown
)

// Assume everyone waiting for the elevator gets on the elevator, even if
// they will be traveling in the "wrong" direction for a while
// Assume that only those that want to travel in the current direction
// enter the elevator, and keep waiting outside otherwise

type Elevator struct {
	Floor int
	Dirn  elevio.MotorDirection
	//Requests [elevio.N_FLOORS][elevio.N_BUTTONS]int
	Requests  [4][3]bool
	OrderBook [4][2]HallRequestStates
	Behaviour ElevatorBehaviour
	Config    Config
}
type Config struct {
	ClearRequestVariant ClearRequestVariant
	DoorOpenDuration_s  float64
}

type Worldview struct {
	Elevators [3]Elevator
}

func MakeHallRequests(elev *Elevator) [][2]bool {
	output := make([][2]bool, len(elev.OrderBook))

	for i, row := range elev.OrderBook {
		for j, val := range row {
			if val == Uncomfirmed || val == Confirmed {
				output[i][j] = true
			} else {
				output[i][j] = false
			}
		}
	}
	return output
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
	conf := Config{ClearRequestVariant: CV_All, DoorOpenDuration_s: 3}
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
