package main

import (
	"Driver-go/modules/elevator"
	"Driver-go/modules/elevio"
	"Driver-go/modules/fsm"
	"Driver-go/modules/timer"
	"fmt"
)
type Elevator = elevator.Elevator
func main() {
	var elev = elevator.Elevator_uninitialized()

	numFloors := 4

	elevio.Init("localhost:15657", numFloors)

	var d elevio.MotorDirection = elevio.MD_Up
	//elevio.SetMotorDirection(d)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)
	drv_timeout := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)
	go timer.PollTimeout(drv_timeout)
	
	for {
		select {
		case a := <-drv_buttons:
			fmt.Printf("%+v\n", a)
			//elevio.SetButtonLamp(a.Button, a.Floor, true)
			fsm.FsmOnRequestButtonPress(a.Floor, a.Button, elev)
			fmt.Printf("%+v\n", a)

		case a := <-drv_floors:
			fmt.Printf("check5")
			fsm.FsmOnFloorArrival(a, elev)

		case a := <-drv_obstr:
			fmt.Printf("%+v\n", a)
			if a {
				elevio.SetMotorDirection(elevio.MD_Stop)
			} else {
				elevio.SetMotorDirection(d)
			}

		case a := <-drv_stop:
			fmt.Printf("%+v\n", a)
			for f := 0; f < numFloors; f++ {
				for b := elevio.ButtonType(0); b < 3; b++ {
					elevio.SetButtonLamp(b, f, false)
				}
			}
		
		case a := <-drv_timeout:
			fmt.Printf("%+v\n", a)
			fsm.FsmOnDoorTimeout(elev)
		}
	fmt.Println("behvaiour", elev.Behaviour)
	}
}
