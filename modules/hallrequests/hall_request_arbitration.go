package main

import (
	"Driver-go/modules/elevator"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
)

// Struct members must be public in order to be accessible by json.Marshal/.Unmarshal
// This means they must start with a capital letter, so we need to use field renaming struct tags to make them camelCase
type Elevator = elevator.Elevator

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

func FillElevState(elev Elevator) HRAElevState {
	switch elev.Behaviour {
	case elevator.EB_Idle, elevator.EB_Moving, elevator.EB_DoorOpen:
		var elev_cab []bool
		for i := 0; i < 4; i++ {
			elev_cab = append(elev_cab, elev.Requests[i][2])
		}
		return HRAElevState{
			Behavior:    elevator.Eb_toString(elev.Behaviour),
			Floor:       elev.Floor,
			Direction:   elevator.Direction_toString(elev.Dirn),
			CabRequests: elev_cab,
		}
	case elevator.EB_Disconnected:
		return HRAElevState{}
	default:
		return HRAElevState{}
	}
}


func isEmptyElevState(state HRAElevState) bool {
    return state.Behavior == "" && state.Floor == 0 && state.Direction == "" && len(state.CabRequests) == 0
}


func FillInput(world elevator.Worldview) HRAInput {
	states := make(map[string]HRAElevState)
	for i, elev := range world.elevators {
		elev_state := FillElevState(elev)
		if !isEmptyElevState(elev_state){
			states[strconv.Itoa(i)] = elev_state
		}
	}

	return HRAInput{
		HallRequests: world.hallRequests_list, //fetch from orderBook, fetch all U and B 
		States:       states,
	}
}


func main() {

	hraExecutable := ""
	switch runtime.GOOS {
	case "linux":
		hraExecutable = "hall_request_assigner"
	case "windows":
		hraExecutable = "hall_request_assigner.exe"
	default:
		panic("OS not supported")
	}

	input := HRAInput{
		HallRequests: [][2]bool{{false, false}, {true, false}, {false, false}, {false, true}},
		States: map[string]HRAElevState{
			"one": HRAElevState{
				Behavior:    "moving",
				Floor:       2,
				Direction:   "up",
				CabRequests: []bool{false, false, false, true},
			},
			"two": HRAElevState{
				Behavior:    "idle",
				Floor:       0,
				Direction:   "stop",
				CabRequests: []bool{false, false, false, false},
			},
		},
	}

	jsonBytes, err := json.Marshal(input)
	if err != nil {
		fmt.Println("json.Marshal error: ", err)
		return
	}

	ret, err := exec.Command("../hall_request_assigner/"+hraExecutable, "-i", string(jsonBytes)).CombinedOutput()
	if err != nil {
		fmt.Println("exec.Command error: ", err)
		fmt.Println(string(ret))
		return
	}

	output := new(map[string][][2]bool)
	err = json.Unmarshal(ret, &output)
	if err != nil {
		fmt.Println("json.Unmarshal error: ", err)
		return
	}

	fmt.Printf("output: \n")
	for k, v := range *output {
		fmt.Printf("%6v :  %+v\n", k, v)
	}
}
