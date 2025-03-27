package hallassigner

import (
	"Driver-go/modules/config"
	"Driver-go/modules/singleElevator"
	"Driver-go/modules/worldview"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
)

type Elevator = singleElevator.Elevator

type HRAElevState struct {
	Behavior    string `json:"behaviour"`
	Floor       int    `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
}

type HRAInput struct {
	HallRequests [][config.N_hall_buttons]bool `json:"hallRequests"`
	States       map[string]HRAElevState       `json:"states"`
}

// Converting from elevator and worlview type to HRAElevState type
func FillHRAElevState(elev Elevator, world worldview.Worldview) HRAElevState {
	switch elev.Behaviour {
	case config.EB_Idle, config.EB_Moving, config.EB_DoorOpen:
		return HRAElevState{
			Behavior:    singleElevator.EbToString(elev.Behaviour),
			Floor:       elev.Floor,
			Direction:   singleElevator.DirectionToString(elev.Dirn),
			CabRequests: worldview.MakeCabRequests(world),
		}
	case config.EB_Disconnected:
		return HRAElevState{}
	default:
		return HRAElevState{}
	}
}

// Converting from worldview type to correct input-format for hallrequest assigner HRAInput
func FillHRAInput(world worldview.Worldview) HRAInput {
	states := make(map[string]HRAElevState)
	for key, elev := range world.Elevators {
		elev_state := FillHRAElevState(elev, world)
		if !isEmptyHRAElevState(elev_state) && !(elev.Behaviour == config.EB_Disconnected || (!elev.Available && key != world.ID)) {
			states[strconv.Itoa(key)] = elev_state
		}
	}
	return HRAInput{
		HallRequests: worldview.MakeHallRequests(world),
		States:       states,
	}
}

func isEmptyHRAElevState(state HRAElevState) bool {
	return state.Behavior == "" && state.Floor == 0 && state.Direction == "" && len(state.CabRequests) == 0
}

// Interacts with (gives input and returns output from) Hallassigner executable
func HallAssigner(world worldview.Worldview) map[string][][config.N_hall_buttons]bool {
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

	output := new(map[string][][config.N_hall_buttons]bool)
	err = json.Unmarshal(ret, &output)
	if err != nil {
		fmt.Println("json.Unmarshal error: ", err)

	}

	/* fmt.Printf("output: \n")
	for k, v := range *output {
		fmt.Printf("%6v :  %+v\n", k, v)
	}  */

	return *output

}

// Processess output from Hallassigner-function, returns requests for id
func HallassignerToElevRequest(hallmap map[string][][config.N_hall_buttons]bool, id string) [config.N_floor_const][config.N_hall_buttons]bool {
	orders := hallmap[id]
	var requests [config.N_floor_const][config.N_hall_buttons]bool
	for i, ordersOnFloor := range orders {
		requests[i][0] = ordersOnFloor[0]
		requests[i][1] = ordersOnFloor[1]
	}
	return requests
}

// Handles hallarbitration logic in main-loop. Receives worldview-struct and returns assigned requests to elevator-module. Is ran as a goroutine.
func HallArbitrationRun(worldViewToArbitrationCh <-chan worldview.Worldview,
	hallRequestToElevatorCh chan<- [config.N_floor_const][config.N_hall_buttons]bool,
	ID string) {
	for {
		select {
		case worldToArbitration := <-worldViewToArbitrationCh:
			hallRequestToElevatorCh <- HallassignerToElevRequest(HallAssigner(worldToArbitration), ID)
		}
	}
}
