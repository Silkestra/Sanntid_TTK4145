package singleElevator

import (
	"Driver-go/modules/config"
	"Driver-go/modules/elevio"
	"fmt"
	"os"
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

	elev := InitElevator(initedFloor)
	drvTimeoutDoor := make(chan bool)
	drvTimeoutAvailable := make(chan bool)
	updatedLocalElevatorForTimer := make(chan config.Elevator)

	go PollDoorTimeout(drvTimeoutDoor, elev)
	go PollAvailableTimeout(drvTimeoutAvailable, updatedLocalElevatorForTimer)

	for {
		updatedLocalElevatorCh <- elev
		updatedLocalElevatorForTimer <- elev

		select {
		case hallRequest := <-hallRequestToElevatorCh:
			elev = updateHallRequests(hallRequest, elev)
			elev = fsmOnRequest(elev, setDoorCh, motorDirectionCh)
			if elev.Behaviour == config.EB_Idle {
				timerStart("available")
			}

		case cabRequest := <-worldviewToCabCh:
			elev = updateCabRequests(cabRequest, elev)
			elev = fsmOnRequest(elev, setDoorCh, motorDirectionCh)

		case floor := <-drvFloors:

			elev.Floor = floor
			elevio.SetFloorIndicator(elev.Floor)

			switch elev.Behaviour {
			case config.EB_Moving:
				if requestsShouldStop(elev) {
					motorDirectionCh <- config.MotorDirection(0)
					setDoorCh <- true
					timerStart("door")
					elev.Behaviour = config.EB_DoorOpen
				}
			}
			elev.Available = true
			timerStart("available")

		case obstruction := <-drvObstr:
			elev.ObstructionActive = obstruction
			fmt.Println("obstructioin:", obstruction)
			if !obstruction {
				timerStart("door")
			}

		case <-drvStop:
			fmt.Println("Terminating...")
			stopLampCh <- true
			motorDirectionCh <- config.MD_Stop
			os.Exit(0)

		case <-drvTimeoutDoor:
			if !elev.ObstructionActive {

				switch elev.Behaviour {
				case config.EB_DoorOpen:
					elev = clearRequestsAtCurrentFloor(elev, requestDoneCh)
					output := requestsChooseDirection(elev)
					elev.Dirn = output.Dirn
					elev.Behaviour = output.Behaviour
					switch elev.Behaviour {
					case config.EB_DoorOpen:
						timerStart("door")
					case config.EB_Moving, config.EB_Idle:
						setDoorCh <- false
						motorDirectionCh <- elev.Dirn
					}
				}
			}

		case <-drvTimeoutAvailable:
			elev.Available = false
			fmt.Println("Available deadline exceeded")
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
	elev := Elevator{Floor: floor, Dirn: config.MD_Stop, Behaviour: config.EB_Idle, ObstructionActive: false, Available: true}
	return elev
}

// Called every time elev.Request is updated, acts according to how FSM is defined for a single elevator
func fsmOnRequest(elev Elevator, setDoorCh chan<- bool, motorDirectionCh chan<- config.MotorDirection) Elevator {
	switch elev.Behaviour {
	case config.EB_Idle:
		output := requestsChooseDirection(elev)
		elev.Dirn = output.Dirn
		elev.Behaviour = output.Behaviour

		switch elev.Behaviour {
		case config.EB_DoorOpen:
			setDoorCh <- true
			timerStart("door")
		case config.EB_Moving:
			motorDirectionCh <- elev.Dirn
		case config.EB_Idle:
			timerStart("available")
		}
	default:
	}
	return elev
}
