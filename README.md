TTK4145 Elevator project
========================
Implemented Peer-to-Peer system for several elevators. Uses go-channels to communicate between different modules. 
All modules have a "run"-function that is ran as a goroutine.

**Modules**: 
    - Single Elevator:
        - Responsible for acting on a request matrix that contains requests that should be serviced by this elevator using logic defined in FSM and CabRequests
        - Heartbeat sent to backup program
        - Module is not concerned with interaction with other elevators
    - Elevio:
        - Module for Elevator IO/Hardware 
        - Responsible for interacting with Hardware through TCP connection 
    - HallAssigner:
        - Responsible for running hallarbitration_executable and fetching the output
    - Config (not a true "module"):
        - Contains constants, types and struct used in several modules 
    - Network:
        - Using UDP Broadcast for communication with peers/other elevators
    - Worldview:
        - Logic for interaction between elevators
        - Worldview struct: contains states for all elevators
        - Cyclic counter on Hall and Cab orders: (Unknown ->) Unconfirmed -> Confirmed -> Done -> Unconfirmed ....
        - Cyclic counters is implemented to ensure sufficient consistency, to handle self-negation (flip-flop)
    - Backup:
        - responsible for launching main program in the event of a software crash 

**Possible Improvements**: 
    - Making all functions pure by not passing pointers as arguments to functions 
        - currenctly accepted using pointers because pointers are not passed between modules 
    - True modularity by not using fixed sized arrays (slices/maps instead)
        - currenctly solved by having a max-number of elevators defined in config-file 
    - Acceptance test for error handling (before Wordlview Merging for example)

**How to run**:
Should be ran using go 1.24 or newer. 

Before running the main.go program, hallassigner should be compiled by utilizing this command in terminal:
"dmd main.d config.d elevator_algorithm.d elevator_state.d optimal_hall_requests.d d-json/jsonx.d -w -g -ofhall_request_assigner;". 

In terminal type: go run main.go -id="fill with id between 0 and N_elevators" "elevatorserver port" 
Ex: "go run main.go -id=0 15657"








