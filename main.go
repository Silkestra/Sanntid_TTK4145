package main

import (
	"Driver-go/modules/config"
	"Driver-go/modules/elevio"
	"Driver-go/modules/hallassigner"
	"Driver-go/modules/network"
	"Driver-go/modules/network/peers"
	"Driver-go/modules/singleElevator"
	"Driver-go/modules/worldview"
	"os"
)

type Elevator = singleElevator.Elevator

func main() {
	port := os.Args[2] // Reads port from terminal

	elevio.Init("localhost:" + port) 

	// Network channels 
	peerTxEnableCh := make(chan bool) 			
	peerUpdateCh := make(chan peers.PeerUpdate) 		   // Network -> Worldview 
	transmittWorldviewCh := make(chan worldview.Worldview) // Worldview -> Network
	recieveWorldviewCh := make(chan worldview.Worldview)   // Network -> Worldview 

	// Single elevator channels 
	setDoorCh := make(chan bool)                        	// SingleElevator -> Elevio 
	requestDoneCh := make(chan config.ButtonEvent)      	// SingleElevator -> Worldview 
	motorDirectionCh := make(chan config.MotorDirection)  	// SingleElevator -> Elevio 
	stopLampCh := make(chan bool)                       	// SingleElevator -> Elevio 
	updatedLocalElevatorCh := make(chan config.Elevator) 	// SingleElevator -> Worldview 

	// Worldview channels 
	requestForLightsCh := make(chan [config.N_floor_const][config.N_buttons_const]bool) 	// Worldview -> Elevio
	worldviewToArbitrationCh := make(chan worldview.Worldview)       						// Worldview -> HallAssigner                     
	worldviewToCabCh := make(chan []bool, 1024)          									// Worldview -> SingleElevator 

	//Hallassigner
	hallRequestToElevatorCh := make(chan [config.N_floor_const][config.N_hall_buttons]bool) // HallAssigner -> SingleElevator 

	// Hardware channels 
	drvButtons := make(chan config.ButtonEvent) 	// Elevio -> Worldview 
	drvFloors := make(chan int)						// Elevio -> SingleElevator 
	drvObstr := make(chan bool)						// Elevio -> SingleElevator 
	drvStop := make(chan bool)						// Elevio -> SingleElevator 
	drvTimeoutDoor := make(chan bool)				// Elevio -> SingleElevator 
	drvTimeoutAvailable := make(chan bool)			// Elevio -> SingleElevator 

	ID := network.InitNetwork(peerUpdateCh, 
		peerTxEnableCh,
		transmittWorldviewCh,
		recieveWorldviewCh)

	floor := elevio.InitHardWare(drvButtons,
		drvFloors,
		drvObstr,
		drvStop,
		drvTimeoutDoor,
		drvTimeoutAvailable)

	var elev = singleElevator.InitElevator(floor) 
	var world = worldview.InitWorldview(*elev, ID) 

	go singleElevator.PollDoorTimeout(drvTimeoutDoor, *elev)
	go singleElevator.PollAvailableTimeout(drvTimeoutAvailable, elev)

	go elevio.ElevatorIORun(motorDirectionCh,
		setDoorCh,
		drvFloors,
		stopLampCh,
		requestForLightsCh)

	go singleElevator.SingleElevatorRun(hallRequestToElevatorCh, 
		updatedLocalElevatorCh,
		drvFloors,
		drvObstr,
		drvStop,
		drvTimeoutDoor,
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

	go worldview.WorldviewRun(peerUpdateCh, 
		drvButtons,          //localRequestsCh in Worldview    
		updatedLocalElevatorCh, 
		recieveWorldviewCh,
		worldviewToArbitrationCh, 
		transmittWorldviewCh,
		requestDoneCh,
		requestForLightsCh,
		worldviewToCabCh,
		world)

	if config.BackupEnable {
		go singleElevator.HeartbeatToBackup(ID, port)
	}
	select {}
}
