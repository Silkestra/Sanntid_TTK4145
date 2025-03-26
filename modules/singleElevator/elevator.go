package singleElevator

import (
	"Driver-go/modules/config"
	"Driver-go/modules/elevio"
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
func InitElevator() Elevator {
	floor := elevio.GetFloor()
	if floor == -1 {
		elevio.SetMotorDirection(config.MD_Up)
		for {
			floor = elevio.GetFloor()
			if floor != -1 {
				elevio.SetMotorDirection(config.MD_Stop)
				break
			}
		}
	}
	elev := Elevator{Floor: floor, Dirn: config.MD_Stop, Behaviour: config.EB_Idle, ObstructionActive: false, Available: true, DoorOpenDuration_s: config.Door_open_time}
	TimerStart(elev.DoorOpenDuration_s, "door")

	return elev
}

// Controls Single Elevator in main-loop. Ran as a goroutine.
func SingleElevatorRun(hallRequestToElevatorCh <-chan [config.N_floor_const][config.N_hall_buttons]bool, // Hallrequests recived from hallarbitration
	updatedLocalElevatorCh chan<- Elevator, // Output channel for updating single elevator to worldview
	setDoorCh chan<- bool, // IO interaction
	requestDoneCh chan<- config.ButtonEvent, // Communicates with Worldview-module that a order is completed
	motorDirectionCh chan<- config.MotorDirection, // IO interaction
	stopLampCh chan<- bool, // IO interaction
	worldviewToCabCh <-chan []bool, // Recieves cabrequests from Worldview
	drvFloors chan int) {

	drvObstr := make(chan bool)            // Elevio -> SingleElevator
	drvStop := make(chan bool)             // Elevio -> SingleElevator
	drvTimeoutDoor := make(chan bool)      // Elevio -> SingleElevator
	drvTimeoutAvailable := make(chan bool) // Elevio -> SingleElevator

	elev := InitElevator()
	go PollAvailableTimeout(drvTimeoutAvailable, &elev)
	go PollDoorTimeout(drvTimeoutDoor, elev)
	go elevio.PollFloorSensor(drvFloors)
	go elevio.PollObstructionSwitch(drvObstr)
	go elevio.PollStopButton(drvStop)

	for {
		updatedLocalElevatorCh <- elev
		select {
		case hallRequest := <-hallRequestToElevatorCh:
			elev = updateHallRequests(hallRequest, elev)
			switch elev.Behaviour {
			case config.EB_Idle:
				output := RequestsChooseDirection(elev)
				elev.Dirn = output.Dirn
				elev.Behaviour = output.Behaviour

				switch elev.Behaviour {
				case config.EB_DoorOpen:
					setDoorCh <- true
					fmt.Printf("hall")
					TimerStart(elev.DoorOpenDuration_s, "door")
					//elev = ClearRequestsAtCurrentFloor(elev, requestDoneCh)
				case config.EB_Moving:
					motorDirectionCh <- elev.Dirn
				case config.EB_Idle:
					//TimerStart(elev.DoorOpenDuration_s, "available") //Starting available deadline
				}
			default:
			}

		case cabRequest := <-worldviewToCabCh:
			elev = updateCabRequests(cabRequest, elev)
			switch elev.Behaviour {
			case config.EB_Idle:
				output := RequestsChooseDirection(elev)
				elev.Dirn = output.Dirn
				elev.Behaviour = output.Behaviour

				switch elev.Behaviour {
				case config.EB_DoorOpen:
					setDoorCh <- true
					fmt.Printf("cab")
					TimerStart(elev.DoorOpenDuration_s, "door")
					// elev = ClearRequestsAtCurrentFloor(elev, requestDoneCh)
				case config.EB_Moving:
					motorDirectionCh <- elev.Dirn
				case config.EB_Idle:
					// TimerStart(elev.DoorOpenDuration_s, "available") //Starting available deadline
				}
			default:
			}

		case floor := <-drvFloors:
			// FsmOnFloorArrival(floor, elev, requestDoneCh, motorDirectionCh, setDoorCh)
			elev.Floor = floor
			elevio.SetFloorIndicator(elev.Floor)
			switch elev.Behaviour {
			case config.EB_Moving:
				if RequestsShouldStop(elev) {
					setDoorCh <- true
					// elev = ClearRequestsAtCurrentFloor(elev, requestDoneCh)
					motorDirectionCh <- config.MotorDirection(0)
					fmt.Printf("drvfloors")
					TimerStart(elev.DoorOpenDuration_s, "door")
					elev.Behaviour = config.EB_DoorOpen
				}
			}
			elev.Available = true
			TimerStart(elev.DoorOpenDuration_s, "available") //Resetting available deadline

		case obstruction := <-drvObstr:
			elev.ObstructionActive = obstruction
			if !obstruction {
				fmt.Printf("drvobst")
				TimerStart(elev.DoorOpenDuration_s, "door")
			}

		case <-drvStop:
			fmt.Println("Terminating...")
			stopLampCh <- true
			motorDirectionCh <- config.MD_Stop
			os.Exit(0)

		case <-drvTimeoutDoor:
			if !elev.ObstructionActive {
				output := RequestsChooseDirection(elev)
				elev.Dirn = output.Dirn
				elev.Behaviour = output.Behaviour
				switch elev.Behaviour {
				case config.EB_DoorOpen:
					fmt.Printf("drvtimeoutdoor")
					TimerStart(elev.DoorOpenDuration_s, "door")
					elev = ClearRequestsAtCurrentFloor(elev, requestDoneCh)
				case config.EB_Moving, config.EB_Idle:
					setDoorCh <- false
					motorDirectionCh <- elev.Dirn
				}
			}
		case <-drvTimeoutAvailable:
			elev.Available = false
		}
	}
}
