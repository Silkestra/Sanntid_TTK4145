package main

import (
	"Driver-go/modules/config"
	"Driver-go/modules/elevio"
	"Driver-go/modules/hallassigner"
	"Driver-go/modules/network"
	"Driver-go/modules/network/peers"
	"Driver-go/modules/singleElevator"
	"Driver-go/modules/worldview"
	"fmt"
	"os"
)

var backupEnable bool = false

type Elevator = singleElevator.Elevator

func main() {
	port := os.Args[2]

	elevio.Init("localhost:" + port) //"localhost:15657"
	fmt.Printf("elevio inited")

	//Network
	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnableCh := make(chan bool)
	transmittWorldviewCh := make(chan worldview.Worldview)
	recieveWorldviewCh := make(chan worldview.Worldview)

	//Single elevator
	setDoorCh := make(chan bool)                         // channel for setting door state
	requestDoneCh := make(chan config.ButtonEvent)       // channel for signaling when request is done
	motorDirectionCh := make(chan config.MotorDirection) // channel for motor direction
	stopLampCh := make(chan bool)                        //setting stoplamp
	requestForLightsCh := make(chan [config.N_floor_const][config.N_buttons_const]bool)

	// Example initialization of channels
	worldviewToArbitrationCh := make(chan worldview.Worldview)                              // read-only channel for Worldview
	hallRequestToElevatorCh := make(chan [config.N_floor_const][config.N_hall_buttons]bool) // write-only channel for hall requests

	//Hardware
	drvButtons := make(chan config.ButtonEvent)
	drvFloors := make(chan int)
	drvObstr := make(chan bool)
	drvStop := make(chan bool)
	drvTimeout := make(chan bool)
	drvTimeoutAvailable := make(chan bool)

	ID := network.InitNetwork(peerUpdateCh, //init og runnework deles for å unngå go i go
		peerTxEnableCh,
		transmittWorldviewCh,
		recieveWorldviewCh)
	fmt.Println("Id", ID)

	floor := elevio.HardWareInit(drvButtons,
		drvFloors,
		drvObstr,
		drvStop,
		drvTimeout,
		drvTimeoutAvailable)
	fmt.Printf("hardware inited")

	var elev = singleElevator.ElevatorUninitialized(floor)
	fmt.Printf("elevator inited")

	//Worldview
	worldviewToCabCh := make(chan []bool, 1)             // Read-only channel for local hall request events
	updatedLocalElevatorCh := make(chan config.Elevator) // Read-only channel for updates on local elevator

	var world = worldview.InitWorldview(*elev, ID)

	fmt.Printf("world inited")

	go singleElevator.PollTimeout(drvTimeout, *elev)
	go singleElevator.PollAvailableTimeout(drvTimeoutAvailable, elev)

	go elevio.ElevatorIORun(motorDirectionCh,
		setDoorCh,
		drvFloors,
		stopLampCh,
		requestForLightsCh)

	go singleElevator.SingleElevatorRun(hallRequestToElevatorCh, //new request recived from hallarbitration
		updatedLocalElevatorCh, // output channel from single elevator to worldview
		drvButtons,
		drvFloors,
		drvObstr,
		drvStop,
		drvTimeout,
		setDoorCh,
		requestDoneCh,
		motorDirectionCh,
		stopLampCh,
		drvTimeoutAvailable,
		worldviewToCabCh,
		elev)

	go hallassigner.HallArbitrationRun(worldviewToArbitrationCh,
		hallRequestToElevatorCh,
		ID)

	go worldview.WorldviewRun(peerUpdateCh, //updates on lost and new elevs comes from network module over channel
		drvButtons,             //local hall request event in elevator (TODO: Not same in WorldviewRun)
		updatedLocalElevatorCh, //recives newest updates on local elevator
		recieveWorldviewCh,
		worldviewToArbitrationCh, //sends current worldview to arbitration logic
		transmittWorldviewCh,
		requestDoneCh,
		requestForLightsCh,
		worldviewToCabCh,
		world)

	if backupEnable {
		go singleElevator.HeartbeatToBackup(ID, port)
	}
	select {}
}
