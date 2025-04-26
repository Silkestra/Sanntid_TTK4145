package elevio

import (
	"Driver-go/modules/config"
	"fmt"
	"net"
	"sync"
	"time"
)

var _initialized bool = false
var _mtx sync.Mutex
var _conn net.Conn

// Controls elevator hardware in main-loop, through interaction with other modules. Is ran as a goroutine.
func ElevatorIORun(motorDirectionCh <-chan config.MotorDirection,
	setDoorCh <-chan bool,
	drvFloors <-chan int,
	stopLampCh <-chan bool,
	requestForLightsCh <-chan [config.N_floor_const][config.N_buttons_const]bool) {
	for {
		select {
		case motorDirn := <-motorDirectionCh:
			SetMotorDirection(motorDirn)
		case doorOpen := <-setDoorCh:
			SetDoorOpenLamp(doorOpen)
		case floor := <-drvFloors:
			SetFloorIndicator(floor)
		case stopLamp := <-stopLampCh:
			SetStopLamp(stopLamp)
		case requests := <-requestForLightsCh:
			setAllLights(requests)
		}
	}

}

// Initalizing elevator hardware in defined floor and starting go threads to interface with IO
func InitHardWare(drvButtons chan<- config.ButtonEvent,
	drvFloors chan<- int,
	drvObstr chan<- bool,
	drvStop chan<- bool, addr string) int {

	if _initialized {
		fmt.Println("Driver already initialized!")
	}
	_mtx = sync.Mutex{}
	var err error
	_conn, err = net.Dial("tcp", addr)
	if err != nil {
		panic(err.Error())
	}
	_initialized = true
	var zeros [config.N_floor_const][config.N_buttons_const]bool
	setAllLights(zeros)
	SetStopLamp(false)
	SetDoorOpenLamp(false)

	floor := GetFloor()
	if floor == -1 {
		SetMotorDirection(config.MD_Up)
		for {
			floor = GetFloor()
			if floor != -1 {
				SetMotorDirection(config.MD_Stop)
				break
			}
		}
	}

	go PollButtons(drvButtons)
	go PollFloorSensor(drvFloors)
	go PollObstructionSwitch(drvObstr)
	go PollStopButton(drvStop)

	return floor
}

// Setting lights for all request network
func setAllLights(HallAndCabReq [config.N_floor_const][config.N_buttons_const]bool) {
	for floor := 0; floor < config.N_floor_const; floor++ {
		for btn := 0; btn < config.N_buttons_const; btn++ {
			SetButtonLamp(config.ButtonType(btn), floor, HallAndCabReq[floor][btn])
		}
	}
}

func SetMotorDirection(dir config.MotorDirection) {
	write([4]byte{1, byte(dir), 0, 0})
}

func SetButtonLamp(button config.ButtonType, floor int, value bool) {
	write([4]byte{2, byte(button), byte(floor), toByte(value)})
}

func SetFloorIndicator(floor int) {
	write([4]byte{3, byte(floor), 0, 0})
}

func SetDoorOpenLamp(value bool) {
	write([4]byte{4, toByte(value), 0, 0})
}

func SetStopLamp(value bool) {
	write([4]byte{5, toByte(value), 0, 0})
}

func PollButtons(receiver chan<- config.ButtonEvent) {
	prev := make([][3]bool, config.N_floor_const)
	for {
		time.Sleep(config.PollRate)
		for f := 0; f < config.N_floor_const; f++ {
			for b := config.ButtonType(0); b < config.ButtonType(config.N_buttons_const); b++ {
				v := GetButton(b, f)
				if v != prev[f][b] && v != false {
					receiver <- config.ButtonEvent{Floor: f, Button: config.ButtonType(b)}
				}
				prev[f][b] = v

			}
		}
	}
}

func PollFloorSensor(receiver chan<- int) {
	prev := -1
	for {
		time.Sleep(config.PollRate)
		v := GetFloor()
		if v != prev && v != -1 {
			receiver <- v
		}
		prev = v
	}
}

func PollStopButton(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(config.PollRate)
		v := GetStop()
		if v != prev {
			receiver <- v
		}
		prev = v
	}
}

func PollObstructionSwitch(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(config.PollRate)
		v := GetObstruction()
		if v != prev {
			receiver <- v
		}
		prev = v
	}
}

func GetButton(button config.ButtonType, floor int) bool {
	a := read([4]byte{6, byte(button), byte(floor), 0})
	return toBool(a[1])
}

func GetFloor() int {
	a := read([4]byte{7, 0, 0, 0})
	if a[1] != 0 {
		return int(a[2])
	} else {
		return -1
	}
}

func GetStop() bool {
	a := read([4]byte{8, 0, 0, 0})
	return toBool(a[1])
}

func GetObstruction() bool {
	a := read([4]byte{9, 0, 0, 0})
	return toBool(a[1])
}

func read(in [4]byte) [4]byte {
	_mtx.Lock()
	defer _mtx.Unlock()

	_, err := _conn.Write(in[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}

	var out [4]byte
	_, err = _conn.Read(out[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}

	return out
}

func write(in [4]byte) {
	_mtx.Lock()
	defer _mtx.Unlock()

	_, err := _conn.Write(in[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}
}

func toByte(a bool) byte {
	var b byte = 0
	if a {
		b = 1
	}
	return b
}

func toBool(a byte) bool {
	var b bool = false
	if a != 0 {
		b = true
	}
	return b
}
