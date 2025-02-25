package main

import (
	"Driver-go/modules/elevator"
	//"Driver-go/modules/elevio"
	//"Driver-go/modules/fsm"
	//"Driver-go/modules/timer"
	//"fmt"
	//"Driver-go/modules/hallrequests"
	"Driver-go/modules/network"

)
type Elevator = elevator.Elevator
func main() {
	//numFloors := 4
	//elevio.Init("localhost:15657", numFloors)
	network.RunNetwork()

	/* var elev = elevator.Elevator_uninitialized()
	//var d elevio.MotorDirection = elevio.MD_Up
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
			
			hallrequests.FillElevRequest(hallrequests.HallAssigner(world, elev), elev)
			fsm.FsmOnRequestButtonPress(a.Floor, a.Button, elev)
			fmt.Printf("%+v\n", a)

		case a := <-drv_floors:
			fmt.Printf("check5")
			fsm.FsmOnFloorArrival(a, elev)

		case a := <-drv_obstr:
			fmt.Printf("%+v\n", a)
			if elev.Behaviour == elevator.EB_DoorOpen{
				elevator.ObstructionActive = a
				fmt.Println("obs:-", elevator.ObstructionActive)
			}
			if !a {
				timer.TimerStart(elev.Config.DoorOpenDuration_s)
			}
			fmt.Println("obs:-", elevator.ObstructionActive)

		case a := <-drv_stop:
			fmt.Println("help......help.......help.......mayday....mayday...your.....teaching.....them....to...solve....the...synchronization.....problem.....with......atom....errrrrr.....arghhhh","%+v\n", a)
			close(drv_buttons)

		
		case a := <-drv_timeout:
			if !elevator.ObstructionActive { //Ignore timeout if obstruction is active
				fmt.Printf("%+v\n", a)
				fsm.FsmOnDoorTimeout(elev)
				}
		}
	} */
}
