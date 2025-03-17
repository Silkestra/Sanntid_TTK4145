package worldview

import (
	"Driver-go/modules/elevio"
	"Driver-go/modules/network/peers"
	"Driver-go/modules/single_elevator"
	"fmt"
	"slices"
	"strconv"
	"time"
)

type RequestStates int

const (
	Unconfirmed RequestStates = iota
	Confirmed
	Done
	Unknown
)

type Worldview struct {
	Elevators     [3]single_elevator.Elevator
	OrderBooks    [3][4][2]RequestStates
	ID            int
	CabOrderBooks [3][3][4]RequestStates
}

func InitWorldview(elev single_elevator.Elevator, id string) *Worldview {
	num, err := strconv.Atoi(id)
	fmt.Println("worldview31")
	if err != nil {
		fmt.Errorf("invalid ID, must be an integer: %v", err)
	}
	fmt.Println("worldview35")
	if num < 0 || num >= len([3]single_elevator.Elevator{}) {
		fmt.Errorf("ID %d is out of valid range [0,2]", num)
	}
	fmt.Println("worldview39")
	world := &Worldview{
		ID: num,
	}
	fmt.Println("worldview43")
	world.Elevators[num] = elev
	fmt.Println("worldview45")
	for i := range world.Elevators {
		if i != num {
			world.Elevators[i].Behaviour = single_elevator.EB_Disconnected
		}
	}
	fmt.Println("worldview51")

	for i := range world.OrderBooks {
		for j := range world.OrderBooks[i] {
			for k := range world.OrderBooks[i][j] {
				world.OrderBooks[i][j][k] = Unknown
			}
		}
	}
	fmt.Println("worldview60")

	for i := range world.CabOrderBooks {
		for j := range world.CabOrderBooks[i] {
			for k := range world.CabOrderBooks[i][j] {
				world.CabOrderBooks[i][j][k] = Unknown
			}
		}
	}
	fmt.Println("worldview69")
	return world
}

func MakeHallRequests(world Worldview) [][2]bool {
	output := make([][2]bool, len(world.OrderBooks[world.ID]))

	for i, row := range world.OrderBooks[world.ID] {
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

func CombineHallAndCabReq(myWorld Worldview) [4][3]bool {
	halls := MakeHallRequests(myWorld) // [4][2]bool
	fmt.Println("halls", halls, "wolrdvew 99")
	cabs := MakeCabRequests(myWorld) // [4]bool
	fmt.Println("cabs", cabs, "wolrdvew 101")
	var combined [4][3]bool // [4][3]bool result

	for floor := 0; floor < 4; floor++ {
		combined[floor][0] = halls[floor][0] // Hall up
		combined[floor][1] = halls[floor][1] // Hall down
		combined[floor][2] = cabs[floor]     // Cab request
	}
	fmt.Println(combined, "wolrd 109")
	return combined
}

//Ta imot newWorldview over kanal, sette inn den heisen den har mottat worlview fra inn i mitt worldview av den heisen

/* func UpdateMyWorldview(myWorld Worldview, newWorld Worldview) Worldview {
	myWorld.Elevators[newWorld.ID] = newWorld.Elevators[newWorld.ID]
	myWorld.OrderBooks[newWorld.ID] = newWorld.OrderBooks[newWorld.ID]

	for j := 0; j < 4; j++ {
		for k := 0; k < 2; k++ {
			allUnconfirmed := true

			for i := 0; i < 3; i++ {
				if myWorld.OrderBooks[i][j][k] != Unconfirmed {
					allUnconfirmed = false
					break
				}
			}
			if allUnconfirmed {
				for i := 0; i < 3; i++ {
					myWorld.OrderBooks[i][j][k] = Confirmed
				}
			}
		}
	}
	return myWorld
}
*/

func UpdateMyElevator(newestElev single_elevator.Elevator, myWorld *Worldview) {
	myWorld.Elevators[myWorld.ID] = newestElev
}

// får tilsendt Buttontype og Floor fra channels
func InsertInOrderBook(btnpressed elevio.ButtonEvent, myWorld *Worldview) {
	if btnpressed.Button == elevio.BT_HallUp || btnpressed.Button == elevio.BT_HallDown {
		myWorld.OrderBooks[myWorld.ID][btnpressed.Floor][btnpressed.Button] = Unconfirmed
	}

	if btnpressed.Button == elevio.BT_Cab {
		myWorld.CabOrderBooks[myWorld.ID][myWorld.ID][btnpressed.Floor] = Unconfirmed
	}
}

// requestDone fås inn som kanal fra cab_request/FSM når en request cleares, main
func DoneInOrderBook(myWorld *Worldview, requestDoneCh elevio.ButtonEvent) {
	floor := requestDoneCh.Floor
	button := int(requestDoneCh.Button)
	fmt.Println("Doneinorderbook, worldview 159 floor and button", floor, button)
	// var lost []int
	// for i, elev := range myWorld.Elevators {
	// 	if elev.Behaviour == single_elevator.EB_Disconnected {
	// 		lost = append(lost, i)
	// 	}
	// }
	// fmt.Println("elevators lost in doneinordferbook 166: ", lost)

	// if button == elevio.BT_Cab {
	// 	tick := 0
	// 	for tick != 3-len(lost) {
	// 		tick = 0
	// 		for i := 0; i < 3; i++ {
	// 			if !slices.Contains(lost, i) && myWorld.CabOrderBooks[myWorld.ID][i][floor] == Confirmed {
	// 				tick += 1
	// 			}
	// 		}
	// 	}
	// 	myWorld.CabOrderBooks[myWorld.ID][myWorld.ID][floor] = Done
	// 	fmt.Println("Cab order cleared, worldview 159")
	if button == elevio.BT_Cab {
		myWorld.CabOrderBooks[myWorld.ID][myWorld.ID][floor] = Done
		fmt.Println("Cab order cleared, worldview 159")

	} else {
		myWorld.OrderBooks[myWorld.ID][floor][button] = Done
		fmt.Println("Hall order cleared, worldview 159")
	}
}

// send peers list from network heartbeat module
func MarkAsUnknown(peer_new string, myWorld *Worldview) {
	if peer_new == strconv.Itoa(myWorld.ID) {
		for i := range myWorld.OrderBooks {
			for j := range myWorld.OrderBooks[i] {
				for k := range myWorld.OrderBooks[i][j] {
					myWorld.OrderBooks[i][j][k] = Unknown
				}
			}
		}

		for i := range myWorld.CabOrderBooks[myWorld.ID] {
			if i != myWorld.ID {
				for j := range myWorld.CabOrderBooks[myWorld.ID][i] {
					myWorld.CabOrderBooks[myWorld.ID][i][j] = Unknown
				}
			}
		}
	}
}

func MarkAsDisconnected(peer_lost []string, myWorld *Worldview) {
	for _, id := range peer_lost {
		num, err := strconv.Atoi(id)
		if err != nil {
			fmt.Errorf("invalid ID, must be an integer: %v", err)
		}
		if num <= 3 && num >= 0 {
			myWorld.Elevators[num].Behaviour = single_elevator.EB_Disconnected

		}
	}
}

func UpdateWorldview(myWorld Worldview, newWorld Worldview) Worldview {
	myWorld.Elevators[newWorld.ID] = newWorld.Elevators[newWorld.ID]
	myWorld.OrderBooks[newWorld.ID] = newWorld.OrderBooks[newWorld.ID]
	myWorld.CabOrderBooks[newWorld.ID] = newWorld.CabOrderBooks[newWorld.ID]
	myWorld.CabOrderBooks[myWorld.ID][newWorld.ID] = newWorld.CabOrderBooks[newWorld.ID][newWorld.ID]
	var lost []int
	for i, elev := range myWorld.Elevators {
		if elev.Behaviour == single_elevator.EB_Disconnected {
			lost = append(lost, i)
		}
	}
	//Orderbook cylic counter
	for j := 0; j < 4; j++ {
		for k := 0; k < 2; k++ {

			switch myWorld.OrderBooks[myWorld.ID][j][k] {

			case Unconfirmed:
				canConfirmOrder := true
				for n := 0; n < 3; n++ {
					if !slices.Contains(lost, n) {
						if myWorld.OrderBooks[n][j][k] == Done {
							canConfirmOrder = false
							break
						}
					}
				}
				if canConfirmOrder {
					myWorld.OrderBooks[myWorld.ID][j][k] = Confirmed
				}

			case Confirmed:
				doneFound := false
				for n := 0; n < 3; n++ {
					if !slices.Contains(lost, n) {
						if myWorld.OrderBooks[n][j][k] == Done {
							doneFound = true
							break
						}
					}
				}
				if doneFound {
					myWorld.OrderBooks[myWorld.ID][j][k] = Done
				}

			case Done:
				unconfirmedFound := false
				for n := 0; n < 3; n++ {
					if !slices.Contains(lost, n) {
						if myWorld.OrderBooks[n][j][k] == Unconfirmed {
							unconfirmedFound = true
							break
						}
					}
				}
				if unconfirmedFound {
					myWorld.OrderBooks[myWorld.ID][j][k] = Unconfirmed
				}

			case Unknown:
				myWorld.OrderBooks[myWorld.ID][j][k] = newWorld.OrderBooks[newWorld.ID][j][k]

			default:
				fmt.Println("Unknown state encountered")
			}
		}
	}
	//Caborderbook cylic counter
	for k := 0; k < 4; k++ {
		switch myWorld.CabOrderBooks[myWorld.ID][myWorld.ID][k] {

		case Unconfirmed:
			canConfirmOrder := true
			for n := 0; n < 3; n++ {
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
			for n := 0; n < 3; n++ {
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
			for n := 0; n < 3; n++ {
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

func WorldView_Run(peerUpdates <-chan peers.PeerUpdate, //updates on lost and new elevs comes from network module over channel
	localRequest <-chan elevio.ButtonEvent, //local hall request event in elevator
	updatedLocalElevator <-chan single_elevator.Elevator, //recives newest updates on local elevator
	recieveWorldView <-chan Worldview,
	worldViewToArbitration chan<- Worldview, //sends current worldview to arbitration logic
	transmittWorldView chan<- Worldview,
	requestDoneCh <-chan elevio.ButtonEvent,
	requestForLightsCh chan<- [4][3]bool,
	worldviewToCab chan<- []bool,
	world *Worldview) { //worldview from peer on network

	ticker := time.NewTicker(500 * time.Millisecond) //rate of sending myworldview to network
	defer ticker.Stop()
	for {
		select {

		case a := <-peerUpdates: // should be struct containing Lost and new part of Peersupdate
			MarkAsDisconnected(a.Lost, world) //
			MarkAsUnknown(a.New, world)

		case a := <-updatedLocalElevator:
			UpdateMyElevator(a, world)
			requestForLightsCh <- CombineHallAndCabReq(*world)

		case a := <-localRequest:
			fmt.Printf("recived local, worlview")
			InsertInOrderBook(a, world)
			requestForLightsCh <- CombineHallAndCabReq(*world)
			worldviewToCab <- MakeCabRequests(*world)

		case a := <-recieveWorldView:
			*world = UpdateWorldview(*world, a)
			fmt.Println("requestsforlights, worldview 356", CombineHallAndCabReq(*world))
			requestForLightsCh <- CombineHallAndCabReq(*world)
			fmt.Println(world.Elevators[world.ID].Behaviour)
			fmt.Println("cab and hall sent to elevio, worldview 358")
			fmt.Println("Caborderbook, wordlview 362", world.CabOrderBooks)
			b := MakeCabRequests(*world)
			fmt.Println("Cabreq, Worldview 363", b)
			worldviewToCab <- b
			fmt.Println("Sent cabs to world, worldview 360")

		case a := <-requestDoneCh:
			fmt.Println("request is done")
			DoneInOrderBook(world, a)
			requestForLightsCh <- CombineHallAndCabReq(*world)
			//worldviewToCab <- MakeCabRequests(*world)
			fmt.Println("Request done, worldview 365")
		case a := <-ticker.C:
			worldViewToArbitration <- *world
			transmittWorldView <- *world
			fmt.Println("Ticker worldview 367", a)
		}
	}
}
