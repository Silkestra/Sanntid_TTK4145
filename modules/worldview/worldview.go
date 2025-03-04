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

type HallRequestStates int

const (
	Unconfirmed HallRequestStates = iota
	Confirmed
	Done
	Unknown
)

type Worldview struct {
	Elevators  [3]single_elevator.Elevator
	OrderBooks [3][4][2]HallRequestStates
	ID         int
}

func InitWorldview(elev single_elevator.Elevator, id string) *Worldview {
	num, err := strconv.Atoi(id)
	if err != nil {
		fmt.Errorf("invalid ID, must be an integer: %v", err)
	}

	if num < 0 || num >= len([3]single_elevator.Elevator{}) {
		fmt.Errorf("ID %d is out of valid range [0,2]", num)
	}

	world := &Worldview{
		ID: num,
	}

	world.Elevators[num] = elev

	for i := range world.Elevators {
		if i != num {
			world.Elevators[i].Behaviour = single_elevator.EB_Disconnected
		}
	}

	for i := range world.OrderBooks {
		for j := range world.OrderBooks[i] {
			for k := range world.OrderBooks[i][j] {
				world.OrderBooks[i][j][k] = Unknown
			}
		}
	}

	return world
}

func MakeHallRequests(world Worldview) [][2]bool {
	output := make([][2]bool, len(world.OrderBooks[world.ID]))

	for i, row := range world.OrderBooks[world.ID] {
		for j, val := range row {
			if val == Unconfirmed || val == Confirmed {
				output[i][j] = true
			} else {
				output[i][j] = false
			}
		}
	}
	return output
}

func CombineHallAndCabReq(myWorld Worldview) [4][3]bool {
	halls := MakeHallRequests(myWorld)             // [4][2]bool
	cabs := myWorld.Elevators[myWorld.ID].Requests // [4][3]bool
	var combined [4][3]bool                        // [4][3]bool result

	for floor := 0; floor < 4; floor++ {
		combined[floor][0] = halls[floor][0] // Hall up
		combined[floor][1] = halls[floor][1] // Hall down
		combined[floor][2] = cabs[floor][2]  // Cab request
	}

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
}

// requestDone fås inn som kanal fra cab_request/FSM når en request cleares, main
func DoneInOrderBook(myWorld *Worldview, requestDoneCh elevio.ButtonEvent) {
	floor := requestDoneCh.Floor
	button := int(requestDoneCh.Button)
	myWorld.OrderBooks[myWorld.ID][floor][button] = Done
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
	}

}

func MarkAsDisconnected(peer_lost []string, myWorld *Worldview) {
	for _, id := range peer_lost {
		num, err := strconv.Atoi(id)
		if err != nil {
			fmt.Errorf("invalid ID, must be an integer: %v", err)
		}
		myWorld.Elevators[num].Behaviour = single_elevator.EB_Disconnected
	}
}

// TODO:  det her er fakked, legg til at dersom elevatoren som ordebooken tilhører har elev.Behaviour == EB_Disconnected --> ignorer
func UpdateWorldview(myWorld Worldview, newWorld Worldview) Worldview {
	myWorld.Elevators[newWorld.ID] = newWorld.Elevators[newWorld.ID]
	myWorld.OrderBooks[newWorld.ID] = newWorld.OrderBooks[newWorld.ID]
	var lost []int
	for i, elev := range myWorld.Elevators {
		if elev.Behaviour == single_elevator.EB_Disconnected {
			lost = append(lost, i)
		}
	}

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
				myWorld.OrderBooks[j][k] = newWorld.OrderBooks[j][k]

			default:
				fmt.Println("Unknown state encountered")
			}
		}
	}
	return myWorld
}

func WorldView_Run(peerUpdates <-chan peers.PeerUpdate, //updates on lost and new elevs comes from network module over channel
	localHallRequest <-chan elevio.ButtonEvent, //local hall request event in elevator
	updatedLocalElevator <-chan single_elevator.Elevator, //recives newest updates on local elevator
	recieveWorldView <-chan Worldview,
	worldViewToArbitration chan<- Worldview, //sends current worldview to arbitration logic
	transmittWorldView chan<- Worldview,
	requestDoneCh <-chan elevio.ButtonEvent,
	requestForLightsCh chan<- [4][3]bool,
	world *Worldview) { //worldview from peer on network

	ticker := time.NewTicker(1 * time.Second) //rate of sending myworldview to network
	defer ticker.Stop()
	for {
		select {

		case a := <-peerUpdates: // should be struct containing Lost and new part of Peersupdate
			MarkAsDisconnected(a.Lost, world) //
			MarkAsUnknown(a.New, world)

		case a := <-updatedLocalElevator:
			UpdateMyElevator(a, world)
			requestForLightsCh <- CombineHallAndCabReq(*world)

		case a := <-localHallRequest:
			InsertInOrderBook(a, world)
			requestForLightsCh <- CombineHallAndCabReq(*world)
		case a := <-recieveWorldView:
			*world = UpdateWorldview(*world, a)
			fmt.Println("requestsforlights", CombineHallAndCabReq(*world))
			requestForLightsCh <- CombineHallAndCabReq(*world)

		case a := <-requestDoneCh:
			DoneInOrderBook(world, a)
			requestForLightsCh <- CombineHallAndCabReq(*world)
		case a := <-ticker.C:
			worldViewToArbitration <- *world
			transmittWorldView <- *world
			fmt.Println(a)
		}
	}
}
