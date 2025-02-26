package worldview

import (
	"Driver-go/modules/elevio"
	"Driver-go/modules/single_elevator"
	"fmt"
	"strconv"
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

	for i := range world.OrderBooks {
		for j := range world.OrderBooks[i] {
			for k := range world.OrderBooks[i][j] {
				world.OrderBooks[i][j][k] = Done
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
		myOrderBook := myWorld.OrderBooks[myWorld.ID]
		myOrderBook[btnpressed.Floor][btnpressed.Button] = Unconfirmed
	}
}

// hvordan finne rjeg tilbake til denne
func DoneInOrderBook(floor int, myWorld *Worldview) {
	for i, elev := range myWorld.Elevators {

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

// TODO:  det her er fakked, legg til at dersom elevatoren som ordebooken tilhører har elev.Behaviour == EB_Disconnecteed --> ignorer
func UpdateWorldview(myWorld Worldview, newWorld Worldview) Worldview {
	myWorld.Elevators[newWorld.ID] = newWorld.Elevators[newWorld.ID]
	myWorld.OrderBooks[newWorld.ID] = newWorld.OrderBooks[newWorld.ID]

	for j := 0; j < 4; j++ {
		for k := 0; k < 2; k++ {
			for i := 0; i < 3; i++ {

				switch myWorld.OrderBooks[i][j][k] {

				case Unconfirmed:
					allUnconfirmed := true
					for n := 0; n < 3; n++ {
						if myWorld.OrderBooks[n][j][k] != Unconfirmed {
							allUnconfirmed = false
							break
						}
					}
					if allUnconfirmed {
						myWorld.OrderBooks[i][j][k] = Confirmed
					}

				case Confirmed:
					doneFound := false
					for n := 0; n < 3; n++ {
						if myWorld.OrderBooks[n][j][k] == Done {
							doneFound = true
							break
						}
					}
					if doneFound {
						myWorld.OrderBooks[i][j][k] = Done
					}

				case Done:
					unconfirmedFound := false
					for n := 0; n < 3; n++ {
						if myWorld.OrderBooks[n][j][k] == Unconfirmed {
							unconfirmedFound = true
							break
						}
					}
					if unconfirmedFound {
						myWorld.OrderBooks[i][j][k] = Unconfirmed
					}

				case Unknown:
					myWorld.OrderBooks = newWorld.OrderBooks

				default:
					fmt.Println("Unknown state encountered")
				}
			}
		}
	}
	return myWorld
}
