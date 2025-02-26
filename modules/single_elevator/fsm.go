package single_elevator

import (
	"Driver-go/modules/elevio"
	"fmt"
)

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
			elevio.SetButtonLamp(elevio.ButtonType(btn), floor, ev.Requests[floor][btn])
		}
	}
}

func FsmOnRequestButtonPress(btnFloor int, btnType elevio.ButtonType, elev *Elevator) {
	fmt.Printf("\n\nRequest button pressed: Floor %d, Type %d\n", btnFloor, btnType)
	//printElevator()

	switch elev.Behaviour {
	case EB_DoorOpen:
		if Requests_shouldClearImmediately(elev, btnFloor, btnType) {
			TimerStart(elev.Config.DoorOpenDuration_s)
		} else {
			if btnType != elevio.BT_Nil {
				elev.Requests[btnFloor][btnType] = true
			}

		}
	case EB_Moving:
		if btnType != elevio.BT_Nil {
			elev.Requests[btnFloor][btnType] = true
		}

	case EB_Idle:
		if btnType != elevio.BT_Nil {
			elev.Requests[btnFloor][btnType] = true
		}

		output := Requests_chooseDirection(elev)
		elev.Dirn = output.Dirn
		elev.Behaviour = output.Behaviour

		switch elev.Behaviour {
		case EB_DoorOpen:
			elevio.SetDoorOpenLamp(true)
			TimerStart(elev.Config.DoorOpenDuration_s)
			elev = ClearRequestsAtCurrentFloor(elev)
		case EB_Moving:
			elevio.SetMotorDirection(elev.Dirn)
		case EB_Idle:
		}
	}
	setAllLights(elev) //move to control from worldview?, io own module?
}

func FsmOnFloorArrival(newFloor int, elev *Elevator) {
	fmt.Printf("\n\nFloor arrival: %d\n", newFloor)
	elev.Floor = newFloor
	elevio.SetFloorIndicator(elev.Floor)

	switch elev.Behaviour {
	case EB_Moving:
		if Requests_shouldStop(elev) {
			elevio.SetMotorDirection(elevio.MotorDirection(0))
			elevio.SetDoorOpenLamp(true)
			elev = ClearRequestsAtCurrentFloor(elev)
			TimerStart(elev.Config.DoorOpenDuration_s)
			setAllLights(elev)
			elev.Behaviour = EB_DoorOpen
		}
	}
}

func FsmOnDoorTimeout(elev *Elevator) {
	fmt.Println("\n\nDoor timeout")
	output := Requests_chooseDirection(elev)
	fmt.Println("ouput", output)
	elev.Dirn = output.Dirn
	elev.Behaviour = output.Behaviour
	switch elev.Behaviour {
	case EB_DoorOpen:
		TimerStart(elev.Config.DoorOpenDuration_s)
		elev = ClearRequestsAtCurrentFloor(elev)
		setAllLights(elev)
	case EB_Moving, EB_Idle:
		elevio.SetDoorOpenLamp(false)
		elevio.SetMotorDirection(elev.Dirn)
	}
}
