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

// Initializes Wordlview module with all other elevators as Disconnected and all Orderbooks in state Unknown 
func InitWorldview(elev Elevator, id string) *Worldview {
	num, err := strconv.Atoi(id)
	if err != nil {
		fmt.Errorf("invalid ID, must be an integer: %v", err)
	}
	if num < 0 || num >= len([config.N_elevators]Elevator{}) {
		fmt.Errorf("ID %d is out of valid range [0,2]", num)
	}
	world := &Worldview{
		ID: num,
	}
	
	world.Elevators[num] = elev
	for i := range world.Elevators {
		if i != num {
			world.Elevators[i].Behaviour = config.EB_Disconnected
		}
	}

	world.HallOrderBooks, world.CabOrderBooks = setAllStatesUnknown(world.HallOrderBooks, world.CabOrderBooks)
	return world
}


func setAllStatesUnknown(HallOrderBooks [config.N_elevators][config.N_floor_const][config.N_hall_buttons]RequestStates, 
	CabOrderBooks [config.N_elevators][config.N_elevators][config.N_floor_const]RequestStates)(
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

func CombineHallAndCabReq(myWorld Worldview) [config.N_floor_const][config.N_buttons_const]bool {
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

func UpdateMyElevator(newestElev Elevator, myWorld *Worldview) {
	myWorld.Elevators[myWorld.ID] = newestElev
}

// Inserts orders in OrderBooks according to ButtonEvent as Unconfirmed 
func InsertInOrderBook(btnpressed config.ButtonEvent, myWorld *Worldview) {
	if btnpressed.Button == config.BT_HallUp || btnpressed.Button == config.BT_HallDown {
		myWorld.HallOrderBooks[myWorld.ID][btnpressed.Floor][btnpressed.Button] = Unconfirmed
	}
	if btnpressed.Button == config.BT_Cab {
		myWorld.CabOrderBooks[myWorld.ID][myWorld.ID][btnpressed.Floor] = Unconfirmed
	}
}

func DoneInOrderBook(myWorld *Worldview, requestDoneCh config.ButtonEvent) {
	floor := requestDoneCh.Floor
	button := int(requestDoneCh.Button)

	if button == config.BT_Cab {
		myWorld.CabOrderBooks[myWorld.ID][myWorld.ID][floor] = Done

	} else {
		myWorld.HallOrderBooks[myWorld.ID][floor][button] = Done
	}
}

// Marks ids received as disconnected in own Worldview
func MarkAsDisconnected(peer_lost []string, myWorld *Worldview) {
	for _, id := range peer_lost {
		num, err := strconv.Atoi(id)
		if err != nil {
			fmt.Errorf("invalid ID, must be an integer: %v", err)
		}
		if num <= 3 && num >= 0 {
			myWorld.Elevators[num].Behaviour = config.EB_Disconnected

		}
	}
}

// Transitions RequestStates in HallOrderBooks
func CyclicCounterHallOrderBook(myWorld Worldview, newWorld Worldview, lost []int) Worldview {
	for j := 0; j < config.N_floor_const; j++ {
		for k := 0; k < config.N_hall_buttons; k++ {

			switch myWorld.HallOrderBooks[myWorld.ID][j][k] {

			case Unconfirmed:
				canConfirmOrder := true
				for n := 0; n < config.N_elevators; n++ {
					if !slices.Contains(lost, n) {
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
				for n := 0; n < config.N_elevators; n++ {
					if !slices.Contains(lost, n) {
						if myWorld.HallOrderBooks[n][j][k] == Done {
							doneFound = true
							break
						}
					}
				}
				if doneFound {
					myWorld.HallOrderBooks[myWorld.ID][j][k] = Done
				}

			case Done:
				unconfirmedFound := false
				for n := 0; n < config.N_elevators; n++ {
					if !slices.Contains(lost, n) {
						if myWorld.HallOrderBooks[n][j][k] == Unconfirmed {
							unconfirmedFound = true
							break
						}
					}
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
	return myWorld
}

// Transitions RequestStates in CabOrderBooks
func CyclicCounterCabOrderBook(myWorld Worldview, newWorld Worldview, lost []int) Worldview {
	for k := 0; k < config.N_floor_const; k++ {
		switch myWorld.CabOrderBooks[myWorld.ID][myWorld.ID][k] {

		case Unconfirmed:
			canConfirmOrder := true
			for n := 0; n < config.N_elevators; n++ {
				if !slices.Contains(lost, n) {
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
				if !slices.Contains(lost, n) {
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
				if !slices.Contains(lost, n) {
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
func UpdateWorldview(myWorld Worldview, newWorld Worldview) Worldview {
	myWorld.Elevators[newWorld.ID] = newWorld.Elevators[newWorld.ID]
	myWorld.HallOrderBooks[newWorld.ID] = newWorld.HallOrderBooks[newWorld.ID]
	myWorld.CabOrderBooks[newWorld.ID] = newWorld.CabOrderBooks[newWorld.ID]
	myWorld.CabOrderBooks[myWorld.ID][newWorld.ID] = newWorld.CabOrderBooks[newWorld.ID][newWorld.ID]

	var lost []int
	for i, elev := range myWorld.Elevators {
		if elev.Behaviour == config.EB_Disconnected {
			lost = append(lost, i)
		}
	}
	
	myWorld = CyclicCounterHallOrderBook(myWorld, newWorld, lost)
	myWorld = CyclicCounterCabOrderBook(myWorld, newWorld, lost)

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
	world *Worldview) { 

	ticker := time.NewTicker(time.Duration(config.N_send_myworld_rate) * time.Millisecond) // Rate of transmitting myworldview to network
	defer ticker.Stop()
	for {
		select {

		case peers := <-peerUpdateCh: 
			MarkAsDisconnected(peers.Lost, world) 

		case elev := <-updatedLocalElevatorCh:
			UpdateMyElevator(elev, world)
			requestForLightsCh <- CombineHallAndCabReq(*world)

		case buttonEvent := <-localRequestCh:
			InsertInOrderBook(buttonEvent, world)
			requestForLightsCh <- CombineHallAndCabReq(*world)
			worldviewToCabCh <- MakeCabRequests(*world)

		case receivedWorld := <-receiveWorldviewCh:
			*world = UpdateWorldview(*world, receivedWorld)
			requestForLightsCh <- CombineHallAndCabReq(*world)
			cabRequests := MakeCabRequests(*world)
			worldviewToCabCh <- cabRequests

		case buttonEvent := <-requestDoneCh:
			DoneInOrderBook(world, buttonEvent)
			requestForLightsCh <- CombineHallAndCabReq(*world)
			
		case <-ticker.C:
			worldviewToArbitrationCh <- *world
			transmittWorldviewCh <- *world
		}
	}
}
