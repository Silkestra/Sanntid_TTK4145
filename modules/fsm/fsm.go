package fsm

import (
	"Driver-go/modules/cabrequests"
	"Driver-go/modules/elevator"
	"Driver-go/modules/elevio"
	"Driver-go/modules/timer"
	"fmt"
	"Driver-go/modules/network/bcast"

)

type Elevator = elevator.Elevator
type ElevatorConfig = elevator.Config
type ButtonType = elevio.ButtonType
type ElevatorBehaviour = elevator.ElevatorBehaviour



// func init() {
// 	elev = Elevator{
// 		Floor:     -1,
// 		Dirn:      0,
// 		Behaviour: 0,
// 		Config: ElevatorConfig{
	
// 			DoorOpenDuration_s: float64(3 * time.Second) / float64(time.Second),
// 		},
// 	}
// }


func setAllLights(ev *Elevator) {
    for floor := 0; floor < elevio.N_FLOORS; floor++ {
        for btn := 0; btn < elevio.N_BUTTONS; btn++ {
            elevio.SetButtonLamp(ButtonType(btn), floor, ev.Requests[floor][btn])
        }
    }
}

func FsmOnRequestButtonPress(btnFloor int, btnType ButtonType, elev *Elevator) {
	fmt.Printf("\n\nRequest button pressed: Floor %d, Type %d\n", btnFloor, btnType)
	//printElevator()

	switch elev.Behaviour {
	case elevator.EB_DoorOpen:
		if cabrequests.Requests_shouldClearImmediately(elev, btnFloor, btnType) {
			timer.TimerStart(elev.Config.DoorOpenDuration_s)
		} else {
			//her
			elev.Requests[btnFloor][btnType] = true

		}
	case elevator.EB_Moving:
		//her
		elev.Requests[btnFloor][btnType] = true

	case elevator.EB_Idle:
		//her
		elev.Requests[btnFloor][btnType] = true

		output := cabrequests.Requests_chooseDirection(elev)
		elev.Dirn = output.Dirn 
		elev.Behaviour = output.Behaviour
		fmt.Println("ouput", output.Behaviour)
		switch elev.Behaviour {
		case elevator.EB_DoorOpen:
			elevio.SetDoorOpenLamp(true)
			timer.TimerStart(elev.Config.DoorOpenDuration_s)
			elev = cabrequests.ClearRequestsAtCurrentFloor(elev)
		case elevator.EB_Moving:
			elevio.SetMotorDirection(elev.Dirn)
		case elevator.EB_Idle:
		}
	}
	setAllLights(elev)
	fmt.Println("\nNew state:")
	//printElevator()
}

func FsmOnFloorArrival(newFloor int, elev *Elevator) {
	fmt.Printf("\n\nFloor arrival: %d\n", newFloor)
	//printElevator()
	elev.Floor = newFloor
	elevio.SetFloorIndicator(elev.Floor)

	switch elev.Behaviour {
	case elevator.EB_Moving:
		fmt.Printf("cheeck10")
		if cabrequests.Requests_shouldStop(elev) {
			fmt.Printf("yes")
			elevio.SetMotorDirection(elevio.MotorDirection(0))
			elevio.SetDoorOpenLamp(true)
			elev = cabrequests.ClearRequestsAtCurrentFloor(elev)
			timer.TimerStart(elev.Config.DoorOpenDuration_s)
			setAllLights(elev)
			elev.Behaviour = elevator.EB_DoorOpen
		}
	}
	fmt.Println("\nNew state:")
	//printElevator()
}

func FsmOnDoorTimeout(elev *Elevator) {
	fmt.Println("\n\nDoor timeout")
	//printElevator()
	output := cabrequests.Requests_chooseDirection(elev)
	fmt.Println("ouput", output)
	elev.Dirn =  output.Dirn
	elev.Behaviour = output.Behaviour
	switch elev.Behaviour {
	case elevator.EB_DoorOpen:
		timer.TimerStart(elev.Config.DoorOpenDuration_s)
		elev = cabrequests.ClearRequestsAtCurrentFloor(elev)
		setAllLights(elev)
	case elevator.EB_Moving, elevator.EB_Idle:
		elevio.SetDoorOpenLamp(false)
		elevio.SetMotorDirection(elev.Dirn)
	}
	//fmt.Println("\nNew state:")
	//printElevator()
}
