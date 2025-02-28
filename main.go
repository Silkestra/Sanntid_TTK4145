package main

import (
	"Driver-go/modules/elevator"
	"Driver-go/modules/elevio"
	"Driver-go/modules/hallassigner"
	"Driver-go/modules/network/bcast"
	"Driver-go/modules/network/localip"
	"Driver-go/modules/network/peers"
	"flag"
	"fmt"
	"os"
	"time"

	//"Driver-go/modules/elevio"
	//"Driver-go/modules/fsm"
	//"Driver-go/modules/timer"
	//"fmt"
	//"Driver-go/modules/hallrequests"
	"Driver-go/modules/network"
	"Driver-go/modules/worldview"
)

type Elevator = elevator.Elevator

func main() {
    //numFloors := 4
    //elevio.Init("localhost:15657", numFloors)
   

    var elev = elevator.Elevator_uninitialized(id)
    //var d elevio.MotorDirection = elevio.MD_Up
    //elevio.SetMotorDirection(d)

    var world = worldview.InitWorldview(elev)


    //Network to Wv
    peerUpdate := make (chan PeerUpdate)

    

    //Make channels
    drv_buttons := make(chan elevio.ButtonEvent)
    drv_floors := make(chan int)
    drv_obstr := make(chan bool)
    drv_stop := make(chan bool)
    drv_timeout := make(chan bool)

    requestChan := make(chan [4][2]bool) //from hallassginer to single elevator 

    a = <- requestChan

    

    
    //sejkk om dårliug kodekvalitet at parameter definisjon er lik navn 
    
    //Network chans for communcation over 

    //Network interface chans
    peerUpdateCh := make(chan peers.PeerUpdate)
    peerTxEnable := make(chan bool)
    transmittWorldView := make(chan elevator.Worldview)
    recieveWorldView := make(chan elevator.Worldview)
    

    //TODO: Package network functionality here 

    
      /*   go func() {
            for {
                helloTx <- *world
                time.Sleep(3 * time.Second)   //transmission rate for sending myworld 
            }
        }()

        go func() {
            for {
                select {
                case p := <-peerUpdateCh:
                    fmt.Printf("Peer update:\n")
                    fmt.Printf("  Peers:    %q\n", p.Peers)
                    fmt.Printf("  New:      %q\n", p.New)
                    fmt.Printf("  Lost:     %q\n", p.Lost)
        
                case a := <-recieveWorldView:
                    fmt.Printf("Received: %#v\n", a)
                }
            }
        }() */
                    //mulig denne funksjonen blir overflødig 
    
    }
}
