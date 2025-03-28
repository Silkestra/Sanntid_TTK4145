package worldview

import (
	"Driver-go/modules/config"
	"Driver-go/modules/network/peers"
	"Driver-go/modules/singleElevator"
	"fmt"
	"slices"
	"strconv"
	"time"
)

type Elevator = config.Elevator

// States for the requests for cyclic counter.
// Uncomfirmed = detected request (local or received over network)
// Confirmed = all active elevators have detected the request (in Unconfirmed state)
// Done = request has been serviced (local or received over network)
// Unknown = elevator has not detected a different state after initialization
type RequestStates int

const (
	Unconfirmed RequestStates = iota
	Confirmed
	Done
	Unknown
)

// Worldview struct consists of a list of all Elevators, a HallOrderBook storing RequestStates for hallrequests for all elevators
// and a CabOrderBook storing RequestStates for cabrequests for all elevators
type Worldview struct {
	ID             int
	Elevators      [config.N_elevators]singleElevator.Elevator
	HallOrderBooks [config.N_elevators][config.N_floor_const][config.N_hall_buttons]RequestStates
	CabOrderBooks  [config.N_elevators][config.N_elevators][config.N_floor_const]RequestStates
	NewPeer        [config.N_elevators]bool
}

// Initializes Wordlview module with all other elevators as Disconnected and all Orderbooks in state Unknown
func InitWorldview(id string) Worldview {
	num, err := strconv.Atoi(id)
	if err != nil {
		fmt.Printf("invalid ID, must be an integer: %v", err)
	}
	if num < 0 || num >= len([config.N_elevators]Elevator{}) {
		fmt.Printf("ID %d is out of valid range [0,2]", num)
	}
	world := Worldview{
		ID: num,
	}
	world.HallOrderBooks, world.CabOrderBooks = setAllStatesUnknown(world.HallOrderBooks, world.CabOrderBooks)
	return world
}

func setAllStatesUnknown(HallOrderBooks [config.N_elevators][config.N_floor_const][config.N_hall_buttons]RequestStates,
	CabOrderBooks [config.N_elevators][config.N_elevators][config.N_floor_const]RequestStates) (
	[config.N_elevators][config.N_floor_const][config.N_hall_buttons]RequestStates,
	[config.N_elevators][config.N_elevators][config.N_floor_const]RequestStates) {
	for i := range HallOrderBooks {
		for j := range HallOrderBooks[i] {
			for k := range HallOrderBooks[i][j] {
				HallOrderBooks[i][j][k] = Unknown
			}
		}
	}
	for i := range CabOrderBooks {
		for j := range CabOrderBooks[i] {
			for k := range CabOrderBooks[i][j] {
				CabOrderBooks[i][j][k] = Unknown
			}
		}
	}
	return HallOrderBooks, CabOrderBooks
}

func MakeHallRequests(world Worldview) [][config.N_hall_buttons]bool {
	output := make([][2]bool, len(world.HallOrderBooks[world.ID]))

	for i, row := range world.HallOrderBooks[world.ID] {
		for j, val := range row {
			if val == Confirmed {
				output[i][j] = true
			} else {
				output[i][j] = false
			}
		}
	}
	return output
}

func MakeCabRequests(world Worldview) []bool {
	output := make([]bool, len(world.CabOrderBooks[world.ID][world.ID]))
	for i, val := range world.CabOrderBooks[world.ID][world.ID] {
		if val == Confirmed {
			output[i] = true
		} else {
			output[i] = false
		}
	}
	return output
}

func combineHallAndCabReq(myWorld Worldview) [config.N_floor_const][config.N_buttons_const]bool {
	halls := MakeHallRequests(myWorld)                              // [4][2]bool
	cabs := MakeCabRequests(myWorld)                                // [4]bool
	var combined [config.N_floor_const][config.N_buttons_const]bool // [4][3]bool result

	for floor := 0; floor < config.N_floor_const; floor++ {
		combined[floor][0] = halls[floor][0] // Hall up
		combined[floor][1] = halls[floor][1] // Hall down
		combined[floor][2] = cabs[floor]     // Cab request
	}
	return combined
}

// Inserts orders in OrderBooks according to ButtonEvent as Unconfirmed
func insertInOrderBook(btnpressed config.ButtonEvent, myWorld Worldview) Worldview {
	if btnpressed.Button == config.BT_HallUp || btnpressed.Button == config.BT_HallDown {
		myWorld.HallOrderBooks[myWorld.ID][btnpressed.Floor][btnpressed.Button] = Unconfirmed
	}
	if btnpressed.Button == config.BT_Cab {
		myWorld.CabOrderBooks[myWorld.ID][myWorld.ID][btnpressed.Floor] = Unconfirmed
	}
	return myWorld
}

func doneInOrderBook(myWorld Worldview, requestDoneCh config.ButtonEvent) Worldview {
	floor := requestDoneCh.Floor
	button := int(requestDoneCh.Button)

	if button == config.BT_Cab {
		myWorld.CabOrderBooks[myWorld.ID][myWorld.ID][floor] = Done

	} else {
		myWorld.HallOrderBooks[myWorld.ID][floor][button] = Done
	}
	return myWorld
}

// Marks ids received as disconnected in own Worldview
func markAsDisconnected(peerLost []string, myWorld Worldview) Worldview {
	for _, id := range peerLost {
		num, err := strconv.Atoi(id)
		if err != nil {
			fmt.Printf("invalid ID, must be an integer: %v", err)
		}
		if num <= 3 && num >= 0 && num != myWorld.ID {
			myWorld.Elevators[num].Behaviour = config.EB_Disconnected
			fmt.Println("mark as disconnected: ", myWorld.Elevators[num].Behaviour)

		}
	}
	fmt.Println("Peer Lost", peerLost)
	/* 	newID, err := strconv.Atoi(newPeer)
	   	if err != nil {
	   		fmt.Printf("invalid ID new peer, must be an integer: %v", err)
	   	}
	   	myWorld.Elevators[newID].Behaviour = config.EB_New */
	if len(peerLost) != 0 {
		lostid, _ := strconv.Atoi(peerLost[0])
		fmt.Println(myWorld.Elevators[lostid].Behaviour)
	}

	return myWorld
}

// Transitions RequestStates in HallOrderBooks
func cyclicCounterHallOrderBook(myWorld Worldview, newWorld Worldview, lost []int) Worldview {
	//fmt.Println("\n Lost list in cyclic", lost)

	for j := 0; j < config.N_floor_const; j++ {
		for k := 0; k < config.N_hall_buttons; k++ {

			switch myWorld.HallOrderBooks[myWorld.ID][j][k] {

			case Unconfirmed:
				canConfirmOrder := true
				for n := 0; n < config.N_elevators; n++ {
					if !slices.Contains(lost, n) && n != myWorld.ID {
						fmt.Println("\n SLICES CONTAIN? :", slices.Contains(lost, n))
						if myWorld.HallOrderBooks[n][j][k] == Done {
							canConfirmOrder = false
							fmt.Println("wrong, should not go here")
							break
						}
					}
				}
				if canConfirmOrder {
					myWorld.HallOrderBooks[myWorld.ID][j][k] = Confirmed
				}

			case Confirmed:
				doneFound := false
				confirmedNewFound := false
				unconfirmedNewFound := false
				for n := 0; n < config.N_elevators; n++ {
					fmt.Println("\n newpeer check", myWorld.NewPeer, time.Now())
					if !slices.Contains(lost, n) && n != myWorld.ID && !myWorld.NewPeer[n] { //&& (n != newPeer && newPeer != myWorld.ID) {
						if myWorld.HallOrderBooks[n][j][k] == Done {
							doneFound = true
							fmt.Println("doneFound", time.Now())
							break
						}
					}
					if myWorld.NewPeer[n] && (myWorld.HallOrderBooks[n][j][k] == Confirmed) {
						confirmedNewFound = true
					}
					if myWorld.NewPeer[n] && (myWorld.HallOrderBooks[n][j][k] == Unconfirmed) {
						unconfirmedNewFound = true
					}
				}
				if doneFound && !confirmedNewFound {
					myWorld.HallOrderBooks[myWorld.ID][j][k] = Done
					fmt.Println("nono its not done", time.Now())
				}
				if unconfirmedNewFound {
					myWorld.HallOrderBooks[myWorld.ID][j][k] = Unconfirmed
				}

			case Done:
				unconfirmedFound := false
				confirmedNewFound := false
				for n := 0; n < config.N_elevators; n++ {
					if !slices.Contains(lost, n) && n != myWorld.ID {
						if myWorld.HallOrderBooks[n][j][k] == Unconfirmed {
							unconfirmedFound = true
							break
						}
					}
					if myWorld.NewPeer[n] && (myWorld.HallOrderBooks[n][j][k] == Confirmed) {
						confirmedNewFound = true
						fmt.Println("confirmedNewFound 1")
					}
				}
				if confirmedNewFound {
					fmt.Println("confirmedNewFound")
					myWorld.HallOrderBooks[myWorld.ID][j][k] = Confirmed
				}
				if unconfirmedFound {
					myWorld.HallOrderBooks[myWorld.ID][j][k] = Unconfirmed
				}

			case Unknown:
				myWorld.HallOrderBooks[myWorld.ID][j][k] = newWorld.HallOrderBooks[newWorld.ID][j][k]

			default:
				fmt.Println("Unknown state encountered")
			}
		}
	}
	fmt.Println("\n peerlist", myWorld.NewPeer, time.Now())
	return myWorld
}

// Transitions RequestStates in CabOrderBooks
func cyclicCounterCabOrderBook(myWorld Worldview, newWorld Worldview, lost []int) Worldview {
	for k := 0; k < config.N_floor_const; k++ {
		switch myWorld.CabOrderBooks[myWorld.ID][myWorld.ID][k] {

		case Unconfirmed:
			canConfirmOrder := true
			for n := 0; n < config.N_elevators; n++ {
				if !slices.Contains(lost, n) && n != myWorld.ID {
					if myWorld.CabOrderBooks[n][myWorld.ID][k] == Done {
						canConfirmOrder = false
						break
					}
				}
			}
			if canConfirmOrder {
				myWorld.CabOrderBooks[myWorld.ID][myWorld.ID][k] = Confirmed
			}

		case Confirmed:
			doneFound := false
			for n := 0; n < config.N_elevators; n++ {
				if !slices.Contains(lost, n) && n != myWorld.ID {
					if myWorld.CabOrderBooks[n][myWorld.ID][k] == Done {
						doneFound = true
						break
					}
				}
			}
			if doneFound {
				myWorld.CabOrderBooks[myWorld.ID][myWorld.ID][k] = Done
			}

		case Done:
			unconfirmedFound := false
			for n := 0; n < config.N_elevators; n++ {
				if !slices.Contains(lost, n) && n != myWorld.ID {
					if myWorld.CabOrderBooks[n][myWorld.ID][k] == Unconfirmed {
						unconfirmedFound = true
						break
					}
				}
			}

			if unconfirmedFound {
				myWorld.CabOrderBooks[myWorld.ID][myWorld.ID][k] = Unconfirmed
			}

		case Unknown:
			myWorld.CabOrderBooks[myWorld.ID][myWorld.ID][k] = newWorld.CabOrderBooks[myWorld.ID][myWorld.ID][k]

		default:
			fmt.Println("Unknown state encountered")
		}
	}
	return myWorld
}

// Updates worldview by merging received worldview into own worldview
func updateWorldview(myWorld Worldview, newWorld Worldview) Worldview {
	myWorld.Elevators[newWorld.ID] = newWorld.Elevators[newWorld.ID]
	myWorld.HallOrderBooks[newWorld.ID] = newWorld.HallOrderBooks[newWorld.ID]
	myWorld.CabOrderBooks[newWorld.ID] = newWorld.CabOrderBooks[newWorld.ID]
	myWorld.CabOrderBooks[myWorld.ID][newWorld.ID] = newWorld.CabOrderBooks[newWorld.ID][newWorld.ID]

	var lost []int
	//var new int
	for i, elev := range myWorld.Elevators {
		//fmt.Println("update worldview elev behav", elev.Behaviour)
		if elev.Behaviour == config.EB_Disconnected {
			//fmt.Println("\n elev disconnected id:", i)
			lost = append(lost, i)
		}
		/* if elev.Behaviour == config.EB_New {
			new = i
		} */
	}
	//fmt.Println("\n in updateWorldview lost peers: ", lost)

	myWorld = cyclicCounterHallOrderBook(myWorld, newWorld, lost)
	myWorld = cyclicCounterCabOrderBook(myWorld, newWorld, lost)

	return myWorld
}

// Controls Worldview-module in main-loop. Ran as a goroutine
func WorldviewRun(peerUpdateCh <-chan peers.PeerUpdate, // Updates on lost and new elevators from network module
	localRequestCh <-chan config.ButtonEvent, // Local request event in elevator
	updatedLocalElevatorCh <-chan Elevator, // Receives newest updates on own elevator
	receiveWorldviewCh <-chan Worldview, // Worldview received over network, to be merged
	worldviewToArbitrationCh chan<- Worldview, // Sends current worldview to hallarbitration logic
	transmittWorldviewCh chan<- Worldview, // Transmitts own worldview to network
	requestDoneCh <-chan config.ButtonEvent, // Receives completed request
	requestForLightsCh chan<- [config.N_floor_const][config.N_buttons_const]bool, // Communicates with IO-module to set lights
	worldviewToCabCh chan<- []bool, // Sends cabrequests to single elevator
	ID string) {

	ticker := time.NewTicker(time.Duration(config.N_send_myworld_rate) * time.Millisecond) // Rate of transmitting myworldview to network
	defer ticker.Stop()

	world := InitWorldview(ID)
	fmt.Println("Inited : ", world.HallOrderBooks)

	for {
		select {

		case peers := <-peerUpdateCh:
			fmt.Println("Peers online", peers.Peers)
			fmt.Println("Peers LOST ", peers.Lost)
			fmt.Println("Peers NEW", peers.New)
			newID, _ := strconv.Atoi(peers.New)
			world.NewPeer[newID] = true
			world = markAsDisconnected(peers.Lost, world)
			//world = updateWorldview(oldWorld, world)

		case elev := <-updatedLocalElevatorCh:
			oldWorld := world
			//fmt.Println("\n Oldworld", oldWorld.HallOrderBooks[world.ID])
			world.Elevators[world.ID] = elev
			world = updateWorldview(oldWorld, world)
			//fmt.Println("\n New world", world.HallOrderBooks[world.ID])
			//fmt.Println("\n req for lights updated elev: ", combineHallAndCabReq(world))
			requestForLightsCh <- combineHallAndCabReq(world)

		case buttonEvent := <-localRequestCh:
			oldWorld := world
			world = insertInOrderBook(buttonEvent, world)
			world = updateWorldview(oldWorld, world)
			//fmt.Println("\n req for lights localreq button: ", combineHallAndCabReq(world))
			requestForLightsCh <- combineHallAndCabReq(world)
			worldviewToCabCh <- MakeCabRequests(world)

		case receivedWorld := <-receiveWorldviewCh:
			oldWorld := world
			fmt.Println("\n Oldworld", oldWorld.HallOrderBooks[world.ID], "\n time now", time.Now())
			world = updateWorldview(world, receivedWorld)
			fmt.Println("\n New world", world.HallOrderBooks[world.ID], "\n time now", time.Now())
			if oldWorld != world {
				//fmt.Println("\n req for lights updated world: ", combineHallAndCabReq(world))
				requestForLightsCh <- combineHallAndCabReq(world)
				worldviewToCabCh <- MakeCabRequests(world)
			}
			world.NewPeer = [config.N_elevators]bool{}
		/* 	requestForLightsCh <- combineHallAndCabReq(world)
		worldviewToCabCh <- MakeCabRequests(world) */

		case buttonEvent := <-requestDoneCh:
			//fmt.Println("\n in wolrdview on requestdone event. Buttonevent for clearing:", buttonEvent)
			oldWorld := world
			world = doneInOrderBook(world, buttonEvent)
			//fmt.Println("\n WorldviewRun: requestdonech", time.Now(), world.Elevators[world.ID].Requests)
			world = updateWorldview(oldWorld, world)
			//fmt.Println("\n req for lights request done: ", combineHallAndCabReq(world))
			requestForLightsCh <- combineHallAndCabReq(world)

		case <-ticker.C:
			//fmt.Println("\n World that gets sent to arbitration", world.Elevators[world.ID])
			worldviewToArbitrationCh <- world
			transmittWorldviewCh <- world
		}
	}
}
