package main

import (
	"Driver-go/modules/elevio"
	"Driver-go/modules/hallassigner"
	"Driver-go/modules/network"
	"Driver-go/modules/network/peers"
	"Driver-go/modules/single_elevator"
	"Driver-go/modules/worldview"
	"fmt"
)

type Elevator = single_elevator.Elevator

func HardWareInit(drv_buttons chan<- elevio.ButtonEvent,
	drv_floors chan<- int,
	drv_obstr chan<- bool,
	drv_stop chan<- bool,
	drv_timeout chan<- bool) {

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)
	go single_elevator.PollTimeout(drv_timeout)
}

func main() {

	numFloors := 4
	elevio.Init("localhost:15657", numFloors)
	fmt.Printf("elevio inited")

	//Network
	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)
	transmittWorldView := make(chan worldview.Worldview)
	recieveWorldView := make(chan worldview.Worldview)

	//Single elevator
	setDoorCh := make(chan bool)                         // channel for setting door state
	requestDoneCh := make(chan elevio.ButtonEvent)       // channel for signaling when request is done
	motorDirectionCh := make(chan elevio.MotorDirection) // channel for motor direction
	stopLampCh := make(chan bool)                        //setting stoplamp
	requestForLightsCh := make(chan [4][3]bool)

	// Example initialization of channels
	worldViewToArbitration := make(chan worldview.Worldview) // read-only channel for Worldview
	hallRequestToElevator := make(chan [4][2]bool)           // write-only channel for hall requests

	//Hardware
	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)
	drv_timeout := make(chan bool)

	//Worldview
	localHallRequest := make(chan elevio.ButtonEvent)           // Read-only channel for local hall request events
	updatedLocalElevator := make(chan single_elevator.Elevator) // Read-only channel for updates on local elevator

	HardWareInit(drv_buttons,
		drv_floors,
		drv_obstr,
		drv_stop,
		drv_timeout)
	fmt.Printf("hardware inited")

	ID := network.InitNetwork(peerUpdateCh, //init og runnework deles for å unngå go i go
		peerTxEnable,
		transmittWorldView,
		recieveWorldView)
	fmt.Println("Id", ID)

	var elev = single_elevator.Elevator_uninitialized()
	fmt.Printf("elevator inited")

	var world = worldview.InitWorldview(*elev, ID)
    fmt.Printf("world inited")


	go elevio.Elevator_io_run(motorDirectionCh,
		setDoorCh,
		drv_floors,
		stopLampCh,
		requestForLightsCh)
    

	go single_elevator.Single_Elevator_Run(hallRequestToElevator, //new request recived from hallarbitration
		updatedLocalElevator, // output channel from single elevator to worldview
		drv_buttons,
		drv_floors,
		drv_obstr,
		drv_stop,
		drv_timeout,
		setDoorCh,
		requestDoneCh,
		motorDirectionCh,
		localHallRequest,
		stopLampCh,
		elev)
    

	go hallassigner.HallArbitration_Run(worldViewToArbitration,
		hallRequestToElevator,
		ID)

	go worldview.WorldView_Run(peerUpdateCh, //updates on lost and new elevs comes from network module over channel
		localHallRequest,     //local hall request event in elevator
		updatedLocalElevator, //recives newest updates on local elevator
		recieveWorldView,
		worldViewToArbitration, //sends current worldview to arbitration logic
		transmittWorldView,
		requestDoneCh,
		requestForLightsCh,
		world)

	select {}
}

//TODO: Package network functionality here

/*   go func() {
         for {
             helloTx <- *world
             time.Sleep(3 * time.Second)   //transmission rate for sending myworld
         }
     }()

     go func() {
         for {
             select {
             case p := <-peerUpdateCh:
                 fmt.Printf("Peer update:\n")
                 fmt.Printf("  Peers:    %q\n", p.Peers)
                 fmt.Printf("  New:      %q\n", p.New)
                 fmt.Printf("  Lost:     %q\n", p.Lost)

             case a := <-recieveWorldView:
                 fmt.Printf("Received: %#v\n", a)
             }
         }
     }() */
//mulig denne funksjonen blir overflødig
