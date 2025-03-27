package singleElevator

import (
	"Driver-go/modules/config"
	"fmt"
	"os"
	"time"
)

type Elevator = config.Elevator

// Controls Single Elevator in main-loop. Ran as a goroutine.
func SingleElevatorRun(hallRequestToElevatorCh <-chan [config.N_floor_const][config.N_hall_buttons]bool, // Hallrequests recived from hallarbitration
	updatedLocalElevatorCh chan<- Elevator, // Output channel for updating single elevator to worldview
	drvFloors <-chan int, // IO interaction
	drvObstr <-chan bool, // IO interaction
	drvStop <-chan bool, // IO interaction
	setDoorCh chan<- bool, // IO interaction
	requestDoneCh chan<- config.ButtonEvent, // Communicates with Worldview-module that a order is completed
	motorDirectionCh chan<- config.MotorDirection, // IO interaction
	stopLampCh chan<- bool, // IO interaction
	worldviewToCabCh <-chan []bool, // Recieves cabrequests from Worldview
	initedFloor int) {

	timerDoorCh := make(chan time.Duration)
	timerAvailableCh := make(chan time.Duration)
	timerDoor := time.NewTimer(0)      // Initial timer with zero duration
	timerAvailable := time.NewTimer(0) // Initial timer with zero duration
	go TimerStarting(timerDoor, timerDoorCh)
	go TimerStarting(timerAvailable, timerAvailableCh)

	elev := InitElevator(initedFloor)

	for {
		updatedLocalElevatorCh <- elev

		select {
		case hallRequest := <-hallRequestToElevatorCh:
			hasChanged := false
			elev, hasChanged = updateHallRequests(hallRequest, elev)
			if hasChanged {
				elev = FsmOnRequest(elev, setDoorCh, requestDoneCh, motorDirectionCh, timerDoorCh, timerAvailableCh)
			}
			if elev.Behaviour == config.EB_Idle {
				timerAvailableCh <- time.Duration(elev.AvailableDuration_s) * time.Second
			}

		case cabRequest := <-worldviewToCabCh:
			elev = updateCabRequests(cabRequest, elev)
			elev = FsmOnRequest(elev, setDoorCh, requestDoneCh, motorDirectionCh, timerDoorCh, timerAvailableCh)

		case floor := <-drvFloors:
			elev = FsmOnFloorArrival(floor, elev, requestDoneCh, motorDirectionCh, setDoorCh, timerDoorCh, timerAvailableCh)

		case obstruction := <-drvObstr:
			elev.ObstructionActive = obstruction

		case <-drvStop:
			fmt.Println("Terminating...")
			stopLampCh <- true
			motorDirectionCh <- config.MD_Stop
			os.Exit(0)

		case <-timerDoor.C:
			fmt.Print("Timeout door")
			if !elev.ObstructionActive {
				elev = FsmOnDoorTimeout(elev, requestDoneCh, motorDirectionCh, setDoorCh, timerDoorCh)
			}

		case <-timerAvailable.C:
			activeRequests := false
			for i := 0; i < config.N_floor_const; i++ {
				for j := 0; j < config.N_buttons_const; j++ {
					if elev.Requests[i][j] {
						activeRequests = true
						break
					}
				}
			}
			if activeRequests {
				elev.Available = false
				fmt.Println("Available deadline exceeded")
			}
		}

	}
}

// Timer functionality for available and door
func TimerStarting(timer *time.Timer, timerDurationCh chan time.Duration) {
	for {
		select {
		case duration := <-timerDurationCh:
			if duration > 0 {
				timer.Reset(duration)
			}
		}
	}
}

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
func updateHallRequests(hallRequest [config.N_floor_const][config.N_hall_buttons]bool, elev Elevator) (Elevator, bool) {
	hasChanged := false
	for i := 0; i < config.N_floor_const; i++ {
		if elev.Requests[i][0] != hallRequest[i][0] || elev.Requests[i][1] != hallRequest[i][1] {
			hasChanged = true
		}
		elev.Requests[i][0] = hallRequest[i][0]
		elev.Requests[i][1] = hallRequest[i][1]
	}
	return elev, hasChanged
}

func updateCabRequests(cabRequest []bool, elev Elevator) Elevator {
	for i := 0; i < config.N_floor_const; i++ {
		elev.Requests[i][2] = cabRequest[i]
	}
	return elev
}

// Initializing single elevator module
func InitElevator(floor int) Elevator {
	elev := Elevator{Floor: floor, Dirn: config.MD_Stop, Behaviour: config.EB_Idle, ObstructionActive: false,
		Available: true, DoorOpenDuration_s: config.Door_open_time, AvailableDuration_s: config.Available_time}
	return elev
}
