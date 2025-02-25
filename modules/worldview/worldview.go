package worldview

import(
	"Driver-go/modules/elevator"
	"Driver-go/modules/elevio"
)

func InitWorldview(elev *elevator.Elevator) *elevator.Worldview{
	world := elevator.Worldview{
		Elevators: [3]elevator.Elevator{

		},
	}

	world.Elevators[elev.ID] = *elev
	return &world 
}

func mergeWorldview(ch <-chan elevator.Worldview){
	recieved_world := <- ch
	
	


}