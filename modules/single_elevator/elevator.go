package single_elevator

import (
	"Driver-go/modules/elevio"
	"fmt"
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
	Available bool
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

func Single_Elevator_Run(reqChan <-chan [4][2]bool, //new request recived from hallarbitration
	elevToWorld chan<- Elevator, // output channel from single elevator to worldview
	drv_buttons chan elevio.ButtonEvent,
	drv_floors <-chan int,
	drv_obstr <-chan bool,
	drv_stop <-chan bool,
	drv_timeout <-chan bool,
	setDoorCh chan<- bool,
	requestDoneCh chan<- elevio.ButtonEvent,
	motorDirectionCh chan<- elevio.MotorDirection,
	stopLampCh chan<- bool,
	drv_timeout_Available <-chan bool,
	cabChan <-chan []bool,
	elev *Elevator) { // buttons from hardware

	for {
		select {
		case newRequest := <-reqChan:
			for i := 0; i < 4; i++ {
				elev.Requests[i][0] = newRequest[i][0]
				elev.Requests[i][1] = newRequest[i][1]
			}

			FsmOnRequestButtonPress(-1, elevio.BT_Nil, elev, setDoorCh, requestDoneCh, motorDirectionCh) //FSM is called to striclty act on what is already modified in requests
			elevToWorld <- *elev
			if elev.Behaviour == EB_Idle {
				TimerStart(elev.Config.DoorOpenDuration_s, "available")
			}

		/* case a := <-drv_buttons:
		if ((elev.Behaviour == EB_DoorOpen) || (elev.Behaviour == EB_Idle) || (elev.Behaviour == EB_Moving)) && ((a.Button == elevio.BT_HallUp) || (a.Button == elevio.BT_HallDown)) {
			localHallRequestChan <- a //cend the hallcall to worldview
			continue
		}
		FsmOnRequestButtonPress(a.Floor, a.Button, elev, setDoorCh, requestDoneCh, motorDirectionCh) // Fsm should only be called of button presses when CABcall or when disconnected
		fmt.Printf("%+v\n", a)
		elevToWorld <- *elev */

		case cabRequest := <-cabChan:
			for i := 0; i < 4; i++ {
				elev.Requests[i][2] = cabRequest[i]
			}
			FsmOnRequestButtonPress(-1, elevio.BT_Nil, elev, setDoorCh, requestDoneCh, motorDirectionCh) //FSM is called to striclty act on what is already modified in requests
			elevToWorld <- *elev
			fmt.Println("Request from reChan: ", elev.Requests)

		case a := <-drv_floors:
			fmt.Printf(" Floorarrive ")
			FsmOnFloorArrival(a, elev, requestDoneCh, motorDirectionCh, setDoorCh)
			elevToWorld <- *elev
			elev.Available = true
			TimerStart(elev.Config.DoorOpenDuration_s, "available") //start availible check

		case a := <-drv_obstr:
			fmt.Printf("%+v\n", a)
			if elev.Behaviour == EB_DoorOpen {
				ObstructionActive = a
				fmt.Println("obs:-", ObstructionActive)
			}
			if !a {
				TimerStart(elev.Config.DoorOpenDuration_s, "door")
			}
			fmt.Println("obs:-", ObstructionActive)
			elevToWorld <- *elev

		case a := <-drv_stop:
			fmt.Println("help......help.......help.......mayday....mayday...your.....teaching.....them....to...solve....the...synchronization.....problem.....with......atom....errrrrr.....arghhhh", "%+v\n", a)
			stopLampCh <- true
			close(drv_buttons)

		case a := <-drv_timeout:
			if !ObstructionActive { //Ignore timeout if obstruction is active
				fmt.Printf("%+v\n", a)
				FsmOnDoorTimeout(elev, requestDoneCh, motorDirectionCh, setDoorCh)
				elevToWorld <- *elev
			}
		case a := <-drv_timeout_Available:
			fmt.Printf("%+v\n", a)
			elev.Available = false

		}
	}
}
