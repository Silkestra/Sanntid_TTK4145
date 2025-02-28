package main

import (
	"Driver-go/modules/elevator"
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

    func elevator_io_run(motorDirection <-chan MotorDirection, ){
        for{
            select{
            case a:= <- motorDirection:
                SetMotorDirection(a)
            case a:= <- setDoor:
                SetDoorOpenLamp(a)
            case a:= <- floorIndicator:
                SetFloorIndicator(a)
            case a:= <- stopLamp:
            case a:= <- buttonLamp:
            }
        }

    }

    func single_elevator_run(reqChan <-chan [4][2]bool,  //new request recived from worldview 
						   elevToWorld chan<- Elevator,  // output channel from single elevator to worldview 
						   drv_buttons  <-chan elevio.ButtonEvent,

                           elev *Elevator){  // buttons from hardware 

        for{
            select{
            case newRequest := <- reqChan :
                for i = 0; i < 4; i++{
                    elev.Request[i][0] = newRequest[i][0]
                    elev.Request[i][1] = newRequest[i][1]
                }
                fsm.FsmOnRequestButtonPress(-1, BT_Nil, elev)   //FSM is called to striclty act on what is already modified in requests
                elevToWolrd <- elev

            case a := <-drv_buttons:
                if EB_Connected && (a == EB_Hall){ 
                    localHallRequestChan <- a     //send the hallcall to worldview 
                    continue 
                }
                fsm.FsmOnRequestButtonPress(a.Floor, a.Button, elev) // Fsm should only be called of button presses when CABcall or when disconnected 
                fmt.Printf("%+v\n", a)
                elevToWorld <- elev

            case a := <-drv_floors:
                fmt.Printf("check5")
                fsm.FsmOnFloorArrival(a, elev)
                elevToWorld <- elev

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
                elevToWorld <- elev

            case a := <-drv_stop:
                fmt.Println("help......help.......help.......mayday....mayday...your.....teaching.....them....to...solve....the...synchronization.....problem.....with......atom....errrrrr.....arghhhh","%+v\n", a)
                close(drv_buttons)
                elevToWorld <- elev


            case a := <-drv_timeout:
                if !elevator.ObstructionActive { //Ignore timeout if obstruction is active
                    fmt.Printf("%+v\n", a)
                    fsm.FsmOnDoorTimeout(elev)
                    elevToWorld <- elev
                    }
                    
            }
        }
    }
    


    func worldView_run(peerUpdates <-chan lost&new,  //updates on lost and new elevs comes from network module over channel 
					 localHallRequest <-chan elevio.ButtonEvent,  //local hall request event in elevator
					 updatedLocalElevator <-chan elev,   //recives newest updates on local elevator 
					 recieveWorldView <-chan WorldView,
                     worldViewToArbitration chan<- WorldView,     //sends current worldview to arbitration logic 
                     world *WorldView){  //worldview from peer on network 

        ticker := time.NewTicker(3 * time.Second)    //rate of sending myworldview to network
        defer ticker.Stop()               
        for{
            select{

            case a := <- peerUpdates:    // should be struct containing Lost and new part of Peersupdate
                MarkAsDisconnected(a.Lost, world) //
                MarkAsUnknown(a.New, world)
                
            case a := <- updatedLocalElevator: 
                UpdateMyElevator(a)

            case a := <- localHallRequest:
                InsertInOrderBook(a.ButtonEvent, world)
                
            case a := <- recieveWorldView:               
                *world = UpdateWorldView(world, a)

            case a := <-ticker.C:
                transmittWorldView <- world
            }
        }
    }


    //if I see that elevator is disconnected 

    func hallarbitration_run(worldViewToArbitration <-chan WorldView,    
                             hallRequestToElevator chan<- request[4][2]bool,
                             ID string){   //recives wolrdviev and outputs to elevator 
        for{
            select{
            case a := <- worldViewToArbitration:
                hallRequestToElevator <- HallassignerToElevRequest(HallAssigner(a,ID))
            }
        }
    }

  


    //numFloors := 4
    //elevio.Init("localhost:15657", numFloors)
   

    var elev = elevator.Elevator_uninitialized(id)
    //var d elevio.MotorDirection = elevio.MD_Up
    //elevio.SetMotorDirection(d)

    var world = worldview.InitWorldview(elev)
    
    //Network chans for communcation over 

    //Network interface chans
    peerUpdateCh := make(chan peers.PeerUpdate)
    peerTxEnable := make(chan bool)
    transmittWorldView := make(chan elevator.Worldview)
    recieveWorldView := make(chan elevator.Worldview)
    

    //TODO: Package network functionality here 
    func InitNetwork(peerUpdateCh,     //init og runnework deles for å unngå go i go 
                     peerTxEnable,
                     transmittWorldView,
                     recieveWorldView) string {   //network init function that inits tansmission and peer heartbeat check
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
        go bcast.Transmitter(16569, transmittWorldView)
        go bcast.Receiver(16569, recieveWorldView)
        go peers.Transmitter(15647, id, peerTxEnable)
        go peers.Receiver(15647, peerUpdateCh)

        return id
    }

    
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

func main(){

}