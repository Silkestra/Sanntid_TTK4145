package worldview

import(
	"Driver-go/modules/single_elevator"
	"Driver-go/modules/hallassigner"

)


type HallRequestStates int

const (
	Uncomfirmed HallRequestStates = iota
	Confirmed
	Done
	Unknown
)

type Worldview struct{
	Elevators [3]single_elevator.Elevator
	OrderBook [4][2]HallRequestStates
}


func InitWorldview(elev single_elevator.Elevator) *Worldview{
	world := Worldview{
		Elevators: [3]single_elevator.Elevator{

		},
	}

	world.Elevators[elev.ID] = elev
	return &world 
}

func MakeHallRequests(world Worldview) [][2]bool {
	output := make([][2]bool, len(world.OrderBook))

	for i, row := range world.OrderBook {
		for j, val := range row {
			if val == Uncomfirmed || val == Confirmed {
				output[i][j] = true
			} else {
				output[i][j] = false
			}
		}
	}
	return output
}


func mergeWorldview(ch <-chan Worldview){
	recieved_world := <- ch
	
	


}