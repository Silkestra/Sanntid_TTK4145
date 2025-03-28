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

	ticker := time.NewTicker(time.Duration(config.N_send_myworld_rate) * time.Millisecond)
	defer ticker.Stop()

	world := InitWorldview(ID)

	for {
		select {

		case peers := <-peerUpdateCh:
			world = markAsDisconnected(peers.Lost, world)
			fmt.Println("received peerupdate", time.Now())
		case elev := <-updatedLocalElevatorCh:
			oldWorld := world
			world.Elevators[world.ID] = elev
			world = updateWorldview(oldWorld, world)
			requestForLightsCh <- combineHallAndCabReq(world)

		case buttonEvent := <-localRequestCh:
			oldWorld := world
			world = insertInOrderBook(buttonEvent, world)
			world = updateWorldview(oldWorld, world)
			requestForLightsCh <- combineHallAndCabReq(world)
			worldviewToCabCh <- MakeCabRequests(world)

		case receivedWorld := <-receiveWorldviewCh:
			fmt.Println("Received world, check my view on him", receivedWorld.HallOrderBooks[receivedWorld.ID])
			oldWorld := world
			fmt.Println("OldWorld:", oldWorld.HallOrderBooks[world.ID])
			world = updateWorldview(oldWorld, receivedWorld)
			requestForLightsCh <- combineHallAndCabReq(world)
			worldviewToCabCh <- MakeCabRequests(world)
			fmt.Println("NewWorld:", world.HallOrderBooks[world.ID])

		case buttonEvent := <-requestDoneCh:
			fmt.Println("Buttoneven reqDone: ", buttonEvent)
			oldWorld := world
			fmt.Println("Requests 1:", world.HallOrderBooks)
			world = doneInOrderBook(world, buttonEvent)
			fmt.Println("Requests 2:", world.HallOrderBooks)
			world = updateWorldview(oldWorld, world)
			fmt.Println("Requests 3:", world.HallOrderBooks)
			requestForLightsCh <- combineHallAndCabReq(world)
			//world.NewPeer = [config.N_elevators]bool{}

		case <-ticker.C:
			worldviewToArbitrationCh <- world
			transmittWorldviewCh <- world
		}
	}
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

// send peers list from network heartbeat module
func MarkAsUnknown(ID int, myWorld Worldview) Worldview {
	if ID != myWorld.ID {
		for i := range myWorld.HallOrderBooks[ID] {
			for j := range myWorld.HallOrderBooks[ID][i] {
				myWorld.HallOrderBooks[ID][i][j] = Unknown
			}
		}
	}
	return myWorld
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
	halls := MakeHallRequests(myWorld)
	cabs := MakeCabRequests(myWorld)
	var combined [config.N_floor_const][config.N_buttons_const]bool

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
		fmt.Println("sets done", myWorld.HallOrderBooks[myWorld.ID][floor][button])
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
		}
	}
	return myWorld
}

// Transitions RequestStates in HallOrderBooks
func cyclicCounterHallOrderBook(myWorld Worldview, newWorld Worldview, lost []int) Worldview {
	//fmt.Print("lost in cyclic hall:", lost)
	for j := 0; j < config.N_floor_const; j++ {
		for k := 0; k < config.N_hall_buttons; k++ {

			switch myWorld.HallOrderBooks[myWorld.ID][j][k] {

			case Unconfirmed:
				canConfirmOrder := true
				for n := 0; n < config.N_elevators; n++ {
					if !slices.Contains(lost, n) && n != myWorld.ID {
						if myWorld.HallOrderBooks[n][j][k] == Done {
							canConfirmOrder = false
							break
						}
					}
				}
				if canConfirmOrder {
					myWorld.HallOrderBooks[myWorld.ID][j][k] = Confirmed
				}

			case Confirmed:
				doneFound := false
				/* confirmedNewFound := false
				unconfirmedNewFound := false */
				for n := 0; n < config.N_elevators; n++ {
					if !slices.Contains(lost, n) && n != myWorld.ID {
						if myWorld.HallOrderBooks[n][j][k] == Done {
							doneFound = true
							fmt.Printf("no dont ")
							break
						}
					}
					/* if myWorld.ID != n && (myWorld.HallOrderBooks[n][j][k] == Confirmed) {
						fmt.Println("in confirmed found new conf")
						confirmedNewFound = true
					}
					if  (myWorld.HallOrderBooks[n][j][k] == Unconfirmed) {
						unconfirmedNewFound = true
						fmt.Println("in confirmed found new unconf")
					} */
				}
				/* if unconfirmedNewFound {
					myWorld.HallOrderBooks[myWorld.ID][j][k] = Unconfirmed
				} */
				/* if  doneFound && !confirmedNewFound {
					myWorld.HallOrderBooks[myWorld.ID][j][k] = Done
				} */
				if doneFound {
					myWorld.HallOrderBooks[myWorld.ID][j][k] = Done
				}

			case Done:
				unconfirmedFound := false
				//confirmedNewFound := false
				for n := 0; n < config.N_elevators; n++ {
					if !slices.Contains(lost, n) && n != myWorld.ID {
						if myWorld.HallOrderBooks[n][j][k] == Unconfirmed {
							unconfirmedFound = true
							fmt.Println("in unconfirmed found don")
						}
					}
					/* if myWorld.ID != n && (myWorld.HallOrderBooks[n][j][k] == Confirmed) {
						confirmedNewFound = true
						fmt.Println("in confirmed found done", myWorld.NewPeer, n)
					} */
				}
				/* if confirmedNewFound {
					myWorld.HallOrderBooks[myWorld.ID][j][k] = Confirmed
				} */
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

func allUnknown(world Worldview, ID int) bool {
	allIsUnknown := true
	for j := 0; j < config.N_floor_const; j++ {
		for k := 0; k < config.N_hall_buttons; k++ {
			if world.HallOrderBooks[ID][j][k] != Unknown {
				allIsUnknown = false
			}
		}
	}
	return allIsUnknown
}

func mergeSpecial(myWorld Worldview, newWorld Worldview) Worldview {

	for j := 0; j < config.N_floor_const; j++ {
		for k := 0; k < config.N_hall_buttons; k++ {
			switch myWorld.HallOrderBooks[myWorld.ID][j][k] {
			case Unconfirmed:
			case Done:
				if !(newWorld.HallOrderBooks[newWorld.ID][j][k] == Unknown) {
					myWorld.HallOrderBooks[myWorld.ID][j][k] = newWorld.HallOrderBooks[newWorld.ID][j][k]
				}
			case Confirmed:
			case Unknown:
				myWorld.HallOrderBooks[myWorld.ID][j][k] = newWorld.HallOrderBooks[newWorld.ID][j][k]
			default:
			}
		}

	}

	//myWorld.HallOrderBooks[newWorld.ID] = myWorld.HallOrderBooks[myWorld.ID]
	return myWorld
}

// Updates worldview by merging received worldview into own worldview
func updateWorldview(myWorld Worldview, newWorld Worldview) Worldview {

	var lost []int
	for i, elev := range myWorld.Elevators {
		if elev.Behaviour == config.EB_Disconnected {
			lost = append(lost, i)
			if i != myWorld.ID {
				myWorld = MarkAsUnknown(i, myWorld)
			}
		}
	}
	fmt.Println("lost list", lost)
	if allUnknown(myWorld, newWorld.ID) && allUnknown(newWorld, myWorld.ID) {
		myWorld.Elevators[newWorld.ID] = newWorld.Elevators[newWorld.ID]
		myWorld.HallOrderBooks[newWorld.ID] = newWorld.HallOrderBooks[newWorld.ID]

		myWorld = mergeSpecial(myWorld, newWorld)
		return myWorld

	} else {
		myWorld.Elevators[newWorld.ID] = newWorld.Elevators[newWorld.ID]
		myWorld.HallOrderBooks[newWorld.ID] = newWorld.HallOrderBooks[newWorld.ID]
		myWorld.CabOrderBooks[newWorld.ID] = newWorld.CabOrderBooks[newWorld.ID]
		myWorld.CabOrderBooks[myWorld.ID][newWorld.ID] = newWorld.CabOrderBooks[newWorld.ID][newWorld.ID]

		myWorld = cyclicCounterHallOrderBook(myWorld, newWorld, lost)
		myWorld = cyclicCounterCabOrderBook(myWorld, newWorld, lost)

		return myWorld
	}

}

/* // Updates worldview by merging received worldview into own worldview
func updateWorldview(myWorld Worldview, newWorld Worldview) Worldview {
	myWorld.Elevators[newWorld.ID] = newWorld.Elevators[newWorld.ID]
	myWorld.HallOrderBooks[newWorld.ID] = newWorld.HallOrderBooks[newWorld.ID]
	myWorld.CabOrderBooks[newWorld.ID] = newWorld.CabOrderBooks[newWorld.ID]
	myWorld.CabOrderBooks[myWorld.ID][newWorld.ID] = newWorld.CabOrderBooks[newWorld.ID][newWorld.ID]
	var lost []int
	for i, elev := range myWorld.Elevators {
		if elev.Behaviour == config.EB_Disconnected {
			lost = append(lost, i)
			if i != myWorld.ID {
				myWorld = MarkAsUnknown(i, myWorld)
			}
		}
	}
	myWorld = cyclicCounterHallOrderBook(myWorld, newWorld, lost)
	myWorld = cyclicCounterCabOrderBook(myWorld, newWorld, lost)
	//myWorld.NewPeer[newWorld.ID] = false

	return myWorld
}
*/
