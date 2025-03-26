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
func updateHallRequests(hallRequest [config.N_floor_const][config.N_hall_buttons]bool, elev Elevator) Elevator {
	for i := 0; i < config.N_floor_const; i++ {
		elev.Requests[i][0] = hallRequest[i][0]
		elev.Requests[i][1] = hallRequest[i][1]
	}
	return elev
}

func updateCabRequests(cabRequest []bool, elev Elevator) Elevator {
	for i := 0; i < config.N_floor_const; i++ {
		elev.Requests[i][2] = cabRequest[i]
	}
	return elev
}

// Initializing single elevator module
func InitElevator(floor int) Elevator {
	elev := Elevator{Floor: floor, Dirn: config.MD_Stop, Behaviour: config.EB_Idle, ObstructionActive: false, Available: true, DoorOpenDuration_s: config.Door_open_time}
	//TimerStart(elev.DoorOpenDuration_s, "door")

	return elev
}

// Controls Single Elevator in main-loop. Ran as a goroutine.
func SingleElevatorRun(hallRequestToElevatorCh <-chan [config.N_floor_const][config.N_hall_buttons]bool, // Hallrequests recived from hallarbitration
	updatedLocalElevatorCh chan<- Elevator, // Output channel for updating single elevator to worldview
	drvFloors <-chan int, // IO interaction
	drvObstr <-chan bool, // IO interaction
	drvStop <-chan bool, // IO interaction
	drvTimeoutDoor <-chan bool, // Door-deadline time exceeded
	setDoorCh chan<- bool, // IO interaction
	requestDoneCh chan<- config.ButtonEvent, // Communicates with Worldview-module that a order is completed
	motorDirectionCh chan<- config.MotorDirection, // IO interaction
	stopLampCh chan<- bool, // IO interaction
	drvTimeoutAvailable <-chan bool, // Available-deadline time exceeded
	worldviewToCabCh <-chan []bool, // Recieves cabrequests from Worldview
	updatedLocalElevatorForTimer chan<- Elevator, // Output channel for updating single elevator to timer
	elev Elevator) {

	for {
		updatedLocalElevatorCh <- elev
		updatedLocalElevatorForTimer <- elev

		select {
		case hallRequest := <-hallRequestToElevatorCh:
			elev = updateHallRequests(hallRequest, elev)
			elev = FsmOnRequest(elev, setDoorCh, requestDoneCh, motorDirectionCh)
			if elev.Behaviour == config.EB_Idle {
				TimerStart(elev.DoorOpenDuration_s, "available")
			}

		case cabRequest := <-worldviewToCabCh:
			elev = updateCabRequests(cabRequest, elev)
			elev = FsmOnRequest(elev, setDoorCh, requestDoneCh, motorDirectionCh)

		case floor := <-drvFloors:
			elev = FsmOnFloorArrival(floor, elev, requestDoneCh, motorDirectionCh, setDoorCh)

		case obstruction := <-drvObstr:
			elev.ObstructionActive = obstruction
			/* if !obstruction {
				TimerStart(elev.DoorOpenDuration_s, "door")
			} */

		case <-drvStop:
			fmt.Println("Terminating...")
			stopLampCh <- true
			motorDirectionCh <- config.MD_Stop
			os.Exit(0)

		case <-drvTimeoutDoor:
			fmt.Print("Timeout door")
			if !elev.ObstructionActive {
				elev = FsmOnDoorTimeout(elev, requestDoneCh, motorDirectionCh, setDoorCh)
			}
			//elev.timerActiveDoor = false
		case <-drvTimeoutAvailable:
			elev.Available = false
			fmt.Println("Available deadline exceeded")
		}
	}
}
