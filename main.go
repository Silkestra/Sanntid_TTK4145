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
	peerUpdateCh := make(chan peers.PeerUpdate)            // Network -> Worldview
	transmittWorldviewCh := make(chan worldview.Worldview) // Worldview -> Network
	recieveWorldviewCh := make(chan worldview.Worldview)   // Network -> Worldview

	// Single elevator channels
	setDoorCh := make(chan bool)                         // SingleElevator -> Elevio
	requestDoneCh := make(chan config.ButtonEvent)       // SingleElevator -> Worldview
	motorDirectionCh := make(chan config.MotorDirection) // SingleElevator -> Elevio
	stopLampCh := make(chan bool)                        // SingleElevator -> Elevio
	updatedLocalElevatorCh := make(chan config.Elevator) // SingleElevator -> Worldview

	// Worldview channels
	requestForLightsCh := make(chan [config.N_floor_const][config.N_buttons_const]bool) // Worldview -> Elevio
	worldviewToArbitrationCh := make(chan worldview.Worldview)                          // Worldview -> HallAssigner
	worldviewToCabCh := make(chan []bool, 1024)                                         // Worldview -> SingleElevator

	//Hallassigner
	hallRequestToElevatorCh := make(chan [config.N_floor_const][config.N_hall_buttons]bool) // HallAssigner -> SingleElevator

	// Hardware channels
	drvFloors := make(chan int)
	drvTimeoutDoor := make(chan bool)      // Elevio -> SingleElevator
	drvTimeoutAvailable := make(chan bool) // Elevio -> SingleElevator


	ID := network.InitNetwork(peerUpdateCh,
		transmittWorldviewCh,
		recieveWorldviewCh)

	go singleElevator.SingleElevatorRun(hallRequestToElevatorCh,
		updatedLocalElevatorCh,
		setDoorCh,
		requestDoneCh,
		motorDirectionCh,
		stopLampCh,
		worldviewToCabCh,
		drvFloors)

	go elevio.ElevatorIORun(motorDirectionCh,
		setDoorCh,
		drvFloors,
		stopLampCh,
		requestForLightsCh)

	go hallassigner.HallArbitrationRun(worldviewToArbitrationCh,
		hallRequestToElevatorCh,
		ID)

	go worldview.WorldviewRun(peerUpdateCh, //localRequestsCh in Worldview
		updatedLocalElevatorCh,
		recieveWorldviewCh,
		worldviewToArbitrationCh,
		transmittWorldviewCh,
		requestDoneCh,
		requestForLightsCh,
		worldviewToCabCh,
		ID)

	if config.BackupEnable {
		go singleElevator.HeartbeatToBackup(ID, port)
	}
	select {}
}
