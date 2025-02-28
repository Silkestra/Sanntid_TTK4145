package main

import (
    "Driver-go/modules/elevator"
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

    go single_elevator_run(reqChan <-chan [4][2]bool,  //new request recived from worldview 
						   elevToWorld chan<- elev,  // output channel from single elevator to worldview 
						   drv_buttons  <-chan elevio.ButtonEvent, ){  // buttons from hardware 
        select{
        case newRequest := <- reqChan :
            for i = 0, i < 4, i++{
                elev.Request[i][0] = newRequest[i][0]
                elev.Request[i][1] = newRequest[i][1]
            }
            fsm.FsmOnRequestButtonPress(-1, BT_Nil, elev)   //FSM is called to striclty act on what is already modified in requests

        case a := <-drv_buttons:
            if EB_Connected & (a == EB_Hall){ 
                localHallRequestChan <- a     //
                continue 
            }
            fsm.FsmOnRequestButtonPress(a.Floor, a.Button, elev) // Fsm should only be called of button presses when CABcall or when disconnected 
            fmt.Printf("%+v\n", a)

        case a := <-drv_floors:
            fmt.Printf("check5")
            fsm.FsmOnFloorArrival(a, elev)




        case a := <-drv_obstr:
            fmt.Printf("%+v\n", a)
            if elev.Behaviour == elevator.EB_DoorOpen{
                elevator.ObstructionActive = a
                fmt.Println("obs:-", elevator.ObstructionActive)
            }
            if !a {
                timer.TimerStart(elev.Config.DoorOpenDuration_s)
            }
            fmt.Println("obs:-", elevator.ObstructionActive)

        case a := <-drv_stop:
            fmt.Println("help......help.......help.......mayday....mayday...your.....teaching.....them....to...solve....the...synchronization.....problem.....with......atom....errrrrr.....arghhhh","%+v\n", a)
            close(drv_buttons)


        case a := <-drv_timeout:
            if !elevator.ObstructionActive { //Ignore timeout if obstruction is active
                fmt.Printf("%+v\n", a)
                fsm.FsmOnDoorTimeout(elev)
                }
        }

        }
    }


    go worldView_run(peerUpdates <-chan lost&new,  //updates on lost and new elevs comes from network module over channel 
					 localHallRequest<-chan elevio.ButtonEvent,  //local hall request event in elevator
					 updatedLocalElevator <-chan elev,   //recives newest updates on local elevator 
					 recieveWorldView <-chan WorldView){  //worldview from peer on network 
        select{

        case a := <- peerUpdates:    // should be struct containing Lost and new part of Peersupdate
            world.MarkAsDisconnected(a.Lost, world) //
			world.MarkAsUnknown(a.New, world)
			
        case a := <- updatedLocalElevator: 
		UpdateMyWorldview(a)

		case a := <- localHallRequest:
			insertInOrderBook(a.ButtonEvent, myworld WolrdView)

		case a := <-


    

        }
    }


    //if I see that elevator is disconnected 

    go hallarbitration()

    go network()
		select {
			case 
		}
    //if I see that elevator is disconnected - > send this to worldview and empty that elevator 


    //numFloors := 4
    //elevio.Init("localhost:15657", numFloors)
    var id string
    flag.StringVar(&id, "id", "", "id of this peer")
    flag.Parse()
    if id == "" {
        localIP, err := localip.LocalIP()
        if err != nil {
            fmt.Println(err)
            localIP = "DISCONNECTED"
        }
        id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
    }

    var elev = elevator.Elevator_uninitialized(id)
    //var d elevio.MotorDirection = elevio.MD_Up
    //elevio.SetMotorDirection(d)

    var world = worldview.InitWorldview(elev)
    //Network
    
    peerUpdateCh := make(chan peers.PeerUpdate)
    peerTxEnable := make(chan bool)
    go peers.Transmitter(15647, id, peerTxEnable)
    go peers.Receiver(15647, peerUpdateCh)
    transmittWorldView := make(chan elevator.Worldview)
    recieveWorldView := make(chan elevator.Worldview)
    go bcast.Transmitter(16569, helloTx)
    go bcast.Receiver(16569, helloRx)

    //TODO: Package network functionality here 


    go func() {
        for {
            helloTx <- *world
            time.Sleep(3 * time.Second)
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
    }()
