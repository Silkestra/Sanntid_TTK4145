package hallassigner

import (
	"Driver-go/modules/single_elevator"
	"Driver-go/modules/worldview"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
)

// Struct members must be public in order to be accessible by json.Marshal/.Unmarshal
// This means they must start with a capital letter, so we need to use field renaming struct tags to make them camelCase
type Elevator = single_elevator.Elevator

type HRAElevState struct {
	Behavior    string `json:"behaviour"`
	Floor       int    `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
}

type HRAInput struct {
	HallRequests [][2]bool               `json:"hallRequests"`
	States       map[string]HRAElevState `json:"states"`
}

func FillHRAElevState(elev Elevator) HRAElevState {
	switch elev.Behaviour {
	case single_elevator.EB_Idle, single_elevator.EB_Moving, single_elevator.EB_DoorOpen:
		var elev_cab []bool
		for i := 0; i < 4; i++ {
			elev_cab = append(elev_cab, elev.Requests[i][2])
		}
		return HRAElevState{
			Behavior:    single_elevator.Eb_toString(elev.Behaviour),
			Floor:       elev.Floor,
			Direction:   single_elevator.Direction_toString(elev.Dirn),
			CabRequests: elev_cab,
		}
	case single_elevator.EB_Disconnected:
		return HRAElevState{}
	default:
		return HRAElevState{}
	}
}

func FillHRAInput(world worldview.Worldview) HRAInput {
	states := make(map[string]HRAElevState)
	for key, elev := range world.Elevators {
		elev_state := FillHRAElevState(elev)
		if !isEmptyHRAElevState(elev_state) {
			states[strconv.Itoa(key)] = elev_state
		}
	}

	return HRAInput{
		HallRequests: worldview.MakeHallRequests(world), //fetch from orderBook, fetch all U and B
		States:       states,
	}
}

func isEmptyHRAElevState(state HRAElevState) bool {
	return state.Behavior == "" && state.Floor == 0 && state.Direction == "" && len(state.CabRequests) == 0
}

func HallAssigner(world worldview.Worldview) map[string][][2]bool {
	hraExecutable := ""
	switch runtime.GOOS {
	case "linux":
		hraExecutable = "hall_request_assigner"
	case "windows":
		hraExecutable = "hall_request_assigner.exe"
	default:
		panic("OS not supported")
	}

	input := FillHRAInput(world)

	jsonBytes, err := json.Marshal(input)
	if err != nil {
		fmt.Println("json.Marshal error: ", err)

	}

	ret, err := exec.Command(hraExecutable, "-i", string(jsonBytes)).CombinedOutput()
	if err != nil {
		fmt.Println("exec.Command error: ", err)
		fmt.Println(string(ret))

	}

	output := new(map[string][][2]bool)
	err = json.Unmarshal(ret, &output)
	if err != nil {
		fmt.Println("json.Unmarshal error: ", err)

	}

	fmt.Printf("output: \n")
	for k, v := range *output {
		fmt.Printf("%6v :  %+v\n", k, v)
	}

	return *output

}

func HallassignerToElevRequest(hallmap map[string][][2]bool, id string) [4][2]bool {
	orders := hallmap[id]
	var requests [4][2]bool
	for i, ordersOnFloor := range orders {
		requests[i][0] = ordersOnFloor[0]
		requests[i][1] = ordersOnFloor[1]
	}
	return requests
}

func HallArbitration_Run(worldViewToArbitration <-chan worldview.Worldview,
	hallRequestToElevator chan<- [4][2]bool,
	ID string) { //recives wolrdviev and outputs to elevator
	for {
		select {
		case a := <-worldViewToArbitration:
			hallRequestToElevator <- HallassignerToElevRequest(HallAssigner(a), ID)
		}
	}
}
