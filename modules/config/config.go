package config

import (
	"time"
)

const PollRate = 20 * time.Millisecond
const N_floor_const int = 4
const N_buttons_const int = 3
const N_hall_buttons int = 2
const N_send_myworld_rate int = 100 //milliseconds

const Door_open_time float64 = 3.0 //seconds
const N_elevators int = 3
const Interval = 15 * time.Millisecond
const Timeout = 2000 * time.Millisecond
const BackupEnable bool = false

type ButtonType int

const (
	BT_HallUp   ButtonType = 0
	BT_HallDown            = 1
	BT_Cab                 = 2
)

type ButtonEvent struct {
	Floor  int
	Button ButtonType
}

type MotorDirection int

const (
	MD_Up   MotorDirection = 1
	MD_Down                = -1
	MD_Stop                = 0
)

type ElevatorBehaviour int

const (
	EB_Idle ElevatorBehaviour = iota
	EB_DoorOpen
	EB_Moving
	EB_Disconnected
)

type Elevator struct {
	Floor              int
	Dirn               MotorDirection
	Requests           [N_floor_const][N_buttons_const]bool
	Behaviour          ElevatorBehaviour
	Available          bool
	ObstructionActive  bool
	DoorOpenDuration_s float64
}
