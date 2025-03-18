package singleElevator

import (
	"Driver-go/modules/config"
	"fmt"
	"os"
)

type Elevator = config.Elevator

func EbToString(eb config.ElevatorBehaviour) string {
	switch eb {
	case config.EB_Idle:
		return "idle"
	case config.EB_DoorOpen:
		return "doorOpen"
	case config.EB_Moving:
		return "moving"
	case config.EB_Disconnected:
		return "disconnected"
	default:
		return "disconnected"
	}
}

func DirectionToString(dirn config.MotorDirection) string {
	switch dirn {
	case config.MD_Up:
		return "up"
	case config.MD_Down:
		return "down"
	case config.MD_Stop:
		return "stop"
	default:
		return "disconnected"
	}
}

func ElevatorUninitialized(floor int) *Elevator {
	conf := config.Config{ClearRequestVariant: config.CV_InDirn, DoorOpenDuration_s: config.Door_open_time}
	elev := Elevator{Floor: floor, Dirn: config.MD_Stop, Behaviour: config.EB_Idle, ObstructionActive: false, Available: true, Config: conf}
	TimerStart(elev.Config.DoorOpenDuration_s, "door")

	return &elev
}

func SingleElevatorRun(hallRequestToElevatorCh <-chan [config.N_floor_const][config.N_hall_buttons]bool, //new request recived from hallarbitration
	updatedLocalElevatorCh chan<- Elevator, // output channel from single elevator to worldview
	drvButtons chan config.ButtonEvent,
	drvFloors <-chan int,
	drvObstr <-chan bool,
	drvStop <-chan bool,
	drvTimeout <-chan bool,
	setDoorCh chan<- bool,
	requestDoneCh chan<- config.ButtonEvent,
	motorDirectionCh chan<- config.MotorDirection,
	stopLampCh chan<- bool,
	drvTimeoutAvailable <-chan bool,
	worldviewToCabCh <-chan []bool,
	elev *Elevator) { // buttons from hardware

	for {
		select {
		case newRequest := <-hallRequestToElevatorCh:
			for i := 0; i < config.N_floor_const; i++ {
				elev.Requests[i][0] = newRequest[i][0]
				elev.Requests[i][1] = newRequest[i][1]
			}
			FsmOnRequest(elev, setDoorCh, requestDoneCh, motorDirectionCh) //FSM is called to striclty act on what is already modified in requests
			updatedLocalElevatorCh <- *elev
			if elev.Behaviour == config.EB_Idle {
				TimerStart(elev.Config.DoorOpenDuration_s, "available")
			}

		case cabRequest := <-worldviewToCabCh:
			for i := 0; i < config.N_floor_const; i++ {
				elev.Requests[i][2] = cabRequest[i]
			}
			FsmOnRequest(elev, setDoorCh, requestDoneCh, motorDirectionCh)
			updatedLocalElevatorCh <- *elev

		case floor := <-drvFloors:
			FsmOnFloorArrival(floor, elev, requestDoneCh, motorDirectionCh, setDoorCh)
			updatedLocalElevatorCh <- *elev
			elev.Available = true
			TimerStart(elev.Config.DoorOpenDuration_s, "available") //start availible check

		case obstruction := <-drvObstr:
			elev.ObstructionActive = obstruction

			if !obstruction {
				TimerStart(elev.Config.DoorOpenDuration_s, "door")
			}
			fmt.Println("obs:-", elev.ObstructionActive)
			updatedLocalElevatorCh <- *elev

		case <-drvStop:
			fmt.Println("help......help.......help.......mayday....mayday...your.....teaching.....them....to...solve....the...synchronization.....problem.....with......atom....errrrrr.....arghhhh")
			stopLampCh <- true
			motorDirectionCh <- config.MD_Stop
			close(drvButtons)
			os.Exit(0)

		case timeout := <-drvTimeout:
			if !elev.ObstructionActive { //Ignore timeout if obstruction is active
				fmt.Printf("%+v\n", timeout)
				FsmOnDoorTimeout(elev, requestDoneCh, motorDirectionCh, setDoorCh)
				updatedLocalElevatorCh <- *elev
			}
		case unavailable := <-drvTimeoutAvailable:
			fmt.Printf("%+v\n", unavailable)
			elev.Available = false

		}
	}
}
