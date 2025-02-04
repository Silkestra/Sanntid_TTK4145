package elevator

import (
	"fmt"
	"time"
)

type Elevator struct {
	Floor     int
	Dirn      Direction
	Behaviour Behaviour
	Requests  [][]bool
	Config    ElevatorConfig
}

type ElevatorConfig struct {
	DoorOpenDuration    time.Duration
	ClearRequestVariant ClearRequestVariant
}

type Button struct {
	Floor int
	Type  ButtonType
}

type Direction int

type Behaviour int

type ClearRequestVariant int

type OutputDevice struct{}

func (od *OutputDevice) DoorLight(on bool) {
	fmt.Printf("Door light: %t\n", on)
}

func (od *OutputDevice) MotorDirection(dir Direction) {
	fmt.Printf("Motor direction: %d\n", dir)
}

func (od *OutputDevice) FloorIndicator(floor int) {
	fmt.Printf("Floor indicator: %d\n", floor)
}

var elevator Elevator
var outputDevice OutputDevice

func init() {
	elevator = Elevator{
		Floor:     -1,
		Dirn:      0,
		Behaviour: 0,
		Requests:  make([][]bool, 4),
		Config: ElevatorConfig{
			DoorOpenDuration: 3 * time.Second,
		},
	}
}

func fsmOnRequestButtonPress(btnFloor int, btnType ButtonType) {
	fmt.Printf("\n\nRequest button pressed: Floor %d, Type %d\n", btnFloor, btnType)
	printElevator()

	switch elevator.Behaviour {
	case DoorOpen:
		if shouldClearImmediately(elevator, btnFloor, btnType) {
			startTimer(elevator.Config.DoorOpenDuration)
		} else {
			elevator.Requests[btnFloor][btnType] = true
		}
	case Moving:
		elevator.Requests[btnFloor][btnType] = true
	case Idle:
		elevator.Requests[btnFloor][btnType] = true
		dirn, behaviour := chooseDirection(elevator)
		elevator.Dirn = dirn
		elevator.Behaviour = behaviour
		switch behaviour {
		case DoorOpen:
			outputDevice.DoorLight(true)
			startTimer(elevator.Config.DoorOpenDuration)
			elevator = clearAtCurrentFloor(elevator)
		case Moving:
			outputDevice.MotorDirection(elevator.Dirn)
		case Idle:
		}
	}
	setAllLights()
	fmt.Println("\nNew state:")
	printElevator()
}

func fsmOnFloorArrival(newFloor int) {
	fmt.Printf("\n\nFloor arrival: %d\n", newFloor)
	printElevator()
	elevator.Floor = newFloor
	outputDevice.FloorIndicator(elevator.Floor)

	switch elevator.Behaviour {
	case Moving:
		if shouldStop(elevator) {
			outputDevice.MotorDirection(0)
			outputDevice.DoorLight(true)
			elevator = clearAtCurrentFloor(elevator)
			startTimer(elevator.Config.DoorOpenDuration)
			setAllLights()
			elevator.Behaviour = DoorOpen
		}
	}
	fmt.Println("\nNew state:")
	printElevator()
}

func fsmOnDoorTimeout() {
	fmt.Println("\n\nDoor timeout")
	printElevator()
	dirn, behaviour := chooseDirection(elevator)
	elevator.Dirn = dirn
	elevator.Behaviour = behaviour
	switch behaviour {
	case DoorOpen:
		startTimer(elevator.Config.DoorOpenDuration)
		elevator = clearAtCurrentFloor(elevator)
		setAllLights()
	case Moving, Idle:
		outputDevice.DoorLight(false)
		outputDevice.MotorDirection(elevator.Dirn)
	}
	fmt.Println("\nNew state:")
	printElevator()
}
