package single_elevator

import (
	"Driver-go/modules/timer"
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
	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			SetButtonLamp(ButtonType(btn), floor, ev.Requests[floor][btn])
		}
	}
}

func FsmOnRequestButtonPress(btnFloor int, btnType ButtonType, elev *Elevator) {
	fmt.Printf("\n\nRequest button pressed: Floor %d, Type %d\n", btnFloor, btnType)
	//printElevator()

	switch elev.Behaviour {
	case EB_DoorOpen:
		if Requests_shouldClearImmediately(elev, btnFloor, btnType) {
			timer.TimerStart(elev.Config.DoorOpenDuration_s)
		} else {
			if btnType != BT_Nil{
				elev.Requests[btnFloor][btnType] = true
			}

		}
	case EB_Moving:
		if btnType != BT_Nil{
			elev.Requests[btnFloor][btnType] = true
		}
		

	case EB_Idle:
		if btnType != BT_Nil{
			elev.Requests[btnFloor][btnType] = true
		}
		
		output := Requests_chooseDirection(elev)
		elev.Dirn = output.Dirn
		elev.Behaviour = output.Behaviour

		switch elev.Behaviour {
		case EB_DoorOpen:
			SetDoorOpenLamp(true)
			timer.TimerStart(elev.Config.DoorOpenDuration_s)
			elev = ClearRequestsAtCurrentFloor(elev)
		case EB_Moving:
			SetMotorDirection(elev.Dirn)
		case EB_Idle:
		}
	}
	setAllLights(elev) //move to control from worldview?, io own module?
}

func FsmOnFloorArrival(newFloor int, elev *Elevator) {
	fmt.Printf("\n\nFloor arrival: %d\n", newFloor)
	elev.Floor = newFloor
	SetFloorIndicator(elev.Floor)

	switch elev.Behaviour {
	case EB_Moving:
		if Requests_shouldStop(elev) {
			SetMotorDirection(MotorDirection(0))
			SetDoorOpenLamp(true)
			elev = ClearRequestsAtCurrentFloor(elev)
			timer.TimerStart(elev.Config.DoorOpenDuration_s)
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
		timer.TimerStart(elev.Config.DoorOpenDuration_s)
		elev = ClearRequestsAtCurrentFloor(elev)
		setAllLights(elev)
	case EB_Moving, EB_Idle:
		SetDoorOpenLamp(false)
		SetMotorDirection(elev.Dirn)
	}
}
