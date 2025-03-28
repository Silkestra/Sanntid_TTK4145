# TTK4145 Elevator Project

Implemented Peer-to-Peer system for several elevators. Uses go-channels to communicate between different modules.  
All modules have a *"run"-function* that is executed as a goroutine.

## **Modules**:
- **Single Elevator**:
  - Responsible for acting on a request matrix that contains requests that should be serviced by this elevator using logic defined in `elevator.go` and `requests.go`.
  - Sends a heartbeat to the backup program.
  - The module is not concerned with interaction with other elevators.

- **Elevio**:
  - Module for **Elevator IO/Hardware**.
  - Responsible for interacting with hardware through a **TCP connection**.

- **HallAssigner**:
  - Responsible for running the **hallarbitration_executable** and fetching the output.

- **Config (not a true "module")**:
  - Contains **constants**, **types**, and **structs** used in several modules.

- **Network**:
  - Uses **UDP Broadcast** for communication with peers/other elevators.

- **Worldview**:
  - Logic for interaction between elevators.
  - The `Worldview` struct contains states for all elevators.
  - **Cyclic counter** on Hall and Cab orders: 
    - (Unknown ->) Unconfirmed -> Confirmed -> Done -> Unconfirmed ….
  - Cyclic counters are implemented to ensure sufficient consistency, and to handle self-negation (flip-flop).

- **Backup**:
  - Responsible for launching the main program in the event of a software crash.

## **Possible Improvements**:
- Not using global variables in timer module
  
- True modularity by not using fixed-size arrays (slices/maps instead):
  - Currently solved by having a max-number of elevators defined in the config file.

- Acceptance test for error handling (before Worldview Merging, for example).

## **How to Run**:
- Should be run using **Go 1.24 or newer**.

1. Before running the `main.go` program, the `hallassigner` should be compiled by utilizing the following command in the terminal:
   ```bash
   dmd main.d config.d elevator_algorithm.d elevator_state.d optimal_hall_requests.d d-json/jsonx.d -w -g -ofhall_request_assigner;








