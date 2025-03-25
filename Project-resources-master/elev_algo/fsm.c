
#include "fsm.h"

#include <stdio.h>

#include "con_load.h"
#include "elevator.h"
#include "elevator_io_device.h"
#include "requests.h"
#include "timer.h"

static Elevator             elevator;
static ElevOutputDevice     outputDevice;


static void __attribute__((constructor)) fsm_init(){
    elevator = elevator_uninitialized();
    
    con_load("elevator.con",
        con_val("doorOpenDuration_s", &elevator.config.doorOpenDuration_s, "%lf")
        con_enum("clearRequestVariant", &elevator.config.clearRequestVariant,
            con_match(CV_All)
            con_match(CV_InDirn)
        )
    )
    
    outputDevice = elevio_getOutputDevice();
}

static void setAllLights(Elevator es){
    for(int floor = 0; floor < N_FLOORS; floor++){
        for(int btn = 0; btn < N_BUTTONS; btn++){
            outputDevice.requestButtonLight(floor, btn, es.requests[floor][btn]);
        }
    }
}

void fsm_onInitBetweenFloors(void){
    outputDevice.motorDirection(D_Down);
    elevator.dirn = D_Down;
    elevator.behaviour = EB_Moving;
}


void fsm_onRequestButtonPress(int btn_floor, Button btn_type){
    printf("\n\n%s(%d, %s)\n", __FUNCTION__, btn_floor, elevio_button_toString(btn_type));
    elevator_print(elevator);
    
    switch(elevator.behaviour){
    case EB_DoorOpen:
        if(requests_shouldClearImmediately(elevator, btn_floor, btn_type)){
            timer_start(elevator.config.doorOpenDuration_s);
        } else {
            elevator.requests[btn_floor][btn_type] = 1;
        }
        break;

    case EB_Moving:
        elevator.requests[btn_floor][btn_type] = 1;
        break;
        
    case EB_Idle:    
        elevator.requests[btn_floor][btn_type] = 1;
        DirnBehaviourPair pair = requests_chooseDirection(elevator);
        elevator.dirn = pair.dirn;
        elevator.behaviour = pair.behaviour;
        switch(pair.behaviour){
        case EB_DoorOpen:
            outputDevice.doorLight(1);
            timer_start(elevator.config.doorOpenDuration_s);
            elevator = requests_clearAtCurrentFloor(elevator);
            break;

        case EB_Moving:
            outputDevice.motorDirection(elevator.dirn);
            breakvoid elevator_print(Elevator es){
    printf("  +--------------------+\n");
    printf(
        "  |floor = %-2d          |\n"
        "  |dirn  = %-12.12s|\n"
        "  |behav = %-12.12s|\n",
        es.floor,
        elevio_dirn_toString(es.dirn),
        eb_toString(es.behaviour)
    );
    printf("  +--------------------+\n");
    printf("  |  | up  | dn  | cab |\n");
    for(int f = N_FLOORS-1; f >= 0; f--){
        printf("  | %d", f);
        for(int btn = 0; btn < N_BUTTONS; btn++){
            if((f == N_FLOORS-1 && btn == B_HallUp)  || 
               (f == 0 && btn == B_HallDown) 
            ){
                printf("|     ");
            } else {
                printf(es.requests[f][btn] ? "|  #  " : "|  -  ");
            }
        }
        printf("|\n");
    }
    printf("  +--------------------+\n");
};
            
        case EB_Idle:
            break;
        }
        break;
    }
    
    setAllLights(elevator);
    
    printf("\nNew state:\n");
    elevator_print(elevator);
}




void fsm_onFloorArrival(int newFloor){
    printf("\n\n%s(%d)\n", __FUNCTION__, newFloor);
    elevator_print(elevator);
    
    elevator.floor = newFloor;
    
    outputDevice.floorIndicator(elevator.floor);
    
    switch(elevator.behaviour){
    case EB_Moving:
        if(requests_shouldStop(elevator)){
            outputDevice.motorDirection(D_Stop);
            outputDevice.doorLight(1);
            elevator = requests_clearAtCurrentFloor(elevator);
            timer_start(elevator.config.doorOpenDuration_s);
            setAllLights(elevator);
            elevator.behaviour = EB_DoorOpen;
        }
        break;
    default:
        break;
    }
    
    printf("\nNew state:\n");
    elevator_print(elevator); 
}




void fsm_onDoorTimeout(void){
    printf("\n\n%s()\n", __FUNCTION__);
    elevator_print(elevator);
    
    switch(elevator.behaviour){
    case EB_DoorOpen:;
        DirnBehaviourPair pair = requests_chooseDirection(elevator);
        elevator.dirn = pair.dirn;
        elevator.behaviour = pair.behaviour;
        
        switch(elevator.behaviour){
        case EB_DoorOpen:
            timer_start(elevator.config.doorOpenDuration_s);
            elevator = requests_clearAtCurrentFloor(elevator);
            setAllLights(elevator);
            break;
        case EB_Moving:
        case EB_Idle:
            outputDevice.doorLight(0);
            outputDevice.motorDirection(elevator.dirn);
            break;
        }
        
        break;
    default:
        break;
    }
    
    printf("\nNew state:\n");
    elevator_print(elevator);
}



package elevator

import (
	"fmt"
	"time"
)

type Elevator struct {
	Floor               int
	Dirn                Direction
	Behaviour           Behaviour
	Requests           [][]bool
	Config             ElevatorConfig
}

type ElevatorConfig struct {
	DoorOpenDuration time.Duration
	ClearRequestVariant ClearRequestVariant
}

type Button struct {
	Floor int
	Type  ButtonType
}

type Direction int

type Behaviour int

type ClearRequestVariant int

type OutputDevice struct {}

func (od *OutputDevice) DoorLight(on bool) {
	fmt.Printf("Door light: %t\n", on)
}

func (od *OutputDevice) MotorDirection(dir Direction) {
	fmt.Printf("Motor direction: %d\n", dir)
}

func (od *OutputDevice) FloorIndicator(floor int) {
	fmt.Printf("Floor indicator: %d\n", floor)
}

var elevator Elevator
var outputDevice OutputDevice

func init() {
	elevator = Elevator{
		Floor:     -1,
		Dirn:      0,
		Behaviour: 0,
		Requests:  make([][]bool, 4),
		Config: ElevatorConfig{
			DoorOpenDuration: 3 * time.Second,
		},
	}
}

func fsmOnRequestButtonPress(btnFloor int, btnType ButtonType) {
	fmt.Printf("\n\nRequest button pressed: Floor %d, Type %d\n", btnFloor, btnType)
	printElevator()

	switch elevator.Behaviour {
	case DoorOpen:
		if shouldClearImmediately(elevator, btnFloor, btnType) {
			startTimer(elevator.Config.DoorOpenDuration)
		} else {
			elevator.Requests[btnFloor][btnType] = true
		}
	case Moving:
		elevator.Requests[btnFloor][btnType] = true
	case Idle:
		elevator.Requests[btnFloor][btnType] = true
		dirn, behaviour := chooseDirection(elevator)
		elevator.Dirn = dirn
		elevator.Behaviour = behaviour
		switch behaviour {
		case DoorOpen:
			outputDevice.DoorLight(true)
			startTimer(elevator.Config.DoorOpenDuration)
			elevator = clearAtCurrentFloor(elevator)
		case Moving:
			outputDevice.MotorDirection(elevator.Dirn)
		case Idle:
		}
	}
	setAllLights()
	fmt.Println("\nNew state:")
	printElevator()
}

func fsmOnFloorArrival(newFloor int) {
	fmt.Printf("\n\nFloor arrival: %d\n", newFloor)
	printElevator()
	elevator.Floor = newFloor
	outputDevice.FloorIndicator(elevator.Floor)

	switch elevator.Behaviour {
	case Moving:
		if shouldStop(elevator) {
			outputDevice.MotorDirection(0)
			outputDevice.DoorLight(true)
			elevator = clearAtCurrentFloor(elevator)
			startTimer(elevator.Config.DoorOpenDuration)
			setAllLights()
			elevator.Behaviour = DoorOpen
		}
	}
	fmt.Println("\nNew state:")
	printElevator()
}

func fsmOnDoorTimeout() {
	fmt.Println("\n\nDoor timeout")
	printElevator()
	dirn, behaviour := chooseDirection(elevator)
	elevator.Dirn = dirn
	elevator.Behaviour = behaviour
	switch behaviour {
	case DoorOpen:
		startTimer(elevator.Config.DoorOpenDuration)
		elevator = clearAtCurrentFloor(elevator)
		setAllLights()
	case Moving, Idle:
		outputDevice.DoorLight(false)
		outputDevice.MotorDirection(elevator.Dirn)
	}
	fmt.Println("\nNew state:")
	printElevator()
}










