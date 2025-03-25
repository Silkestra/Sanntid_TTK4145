TTK4145 Elevator project 
========================
Implemented Peer-to-Peer system for several elevators. Uses go-channels to communicate between different modules. 
All modules have a "run"-function that is ran as a goroutine.
---
Modules: 
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

```
Possible Improvements: 
    - Making all functions pure by not passing pointers as arguments to functions 
        - currenctly accepted using pointers because pointers are not passed between modules 
    - True modularity by not using fixed sized arrays (slices/maps instead)
        - currenctly solved by having a max-number of elevators defined in config-file 
    - Acceptance test for error handling (before Wordlview Merging for example)

```









