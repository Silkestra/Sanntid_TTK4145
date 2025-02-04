package elevator

import "fmt"
type ElevatorBehaviour int

const (
    EB_Idle ElevatorBehaviour = iota
    EB_DoorOpen
    EB_Moving
)

type ClearRequestVariant int

const (
    CV_All ClearRequestVariant = iota
    CV_InDirn
)

// Assume everyone waiting for the elevator gets on the elevator, even if 
// they will be traveling in the "wrong" direction for a while
// Assume that only those that want to travel in the current direction 
// enter the elevator, and keep waiting outside otherwise

type Elevator struct {
    floor int
    dirn Dirn
    request [N_FLOORS][N_BUTTONS] int 
    behaviour ElevatorBehaviour
    config Config
}
 type Config struct {
    clearRequestVariant ClearRequestVariant
    doorOpenDuration_s double
}

func elevator_print(es Elevator) void {

}

func eb_toString (eb ElevatorBehaviour) *char {
    switch eb {
    case EB_idle:
        return "EB_Idle"
    case EB_DoorOpen:
        return "EB_DoorOpen"
    case EB_Moving:
        return "EB_Moving"
    default:
        "EB_UNDEFINED"
    }
}


func elevatorPrint(es Elevator) void {
    fmt.Println("  +--------------------+")
    fmt.Printf(
        "  |floor = %-2d          |\n"+
            "  |dirn  = %-12s|\n"+
            "  |behav = %-12s|\n",
        es.floor,
        elevio_dirn_toString(es.dirn),
        eb_toString(es.behaviour)
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
}
func elevator_uninitialized(void) *Elevator{
    conf := config{clearRequestVariant : CV_All, doorOpenDuration_s: 3}
    p := Elevator{floor: -1, dirn: D_Stop, behaviour : EB_Idle, config: conf}
    return &p
}

