package network

import (
	"Driver-go/modules/elevator"
	"Driver-go/modules/network/bcast"
	"Driver-go/modules/network/localip"
	"Driver-go/modules/network/peers"
	"flag"
	"fmt"
	"os"
	"time"
)

// We define some custom struct to send over the network.
// Note that all members we want to transmit must be public. Any private members
//

type Elevator = elevator.Elevator

// will be received as zero-values.
type HelloMsg struct {
	Message string
	Iter    int
}

func RunNetwork() elevator.Worldview {
	// Our id can be anything. Here we pass it on the command line, using
	//  `go run main.go -id=our_id`
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	// ... or alternatively, we can use the local IP address.
	// (But since we can run multiple programs on the same PC, we also append the
	//  process ID)
	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}

	// We make a channel for receiving updates on the id's of the peers that are
	//  alive on the network
	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)
	// We can disable/enable the transmitter after it has been started.
	// This could be used to signal that we are somehow "unavailable".
	go peers.Transmitter(15647, id, peerTxEnable)
	go peers.Receiver(15647, peerUpdateCh)

	// We make channels for sending and receiving our custom data types
	helloTx := make(chan elevator.Worldview)
	helloRx := make(chan elevator.Worldview)
	// ... and start the transmitter/receiver pair on some port
	// These functions can take any number of channels! It is also possible to
	//  start multiple transmitters/receivers on the same port.
	go bcast.Transmitter(16569, helloTx)
	go bcast.Receiver(16569, helloRx)

	// The example message. We just send one of these every second.

	elev1 := elevator.Elevator{
		ID: 1,
	}
	elev2 := elevator.Elevator{
		ID: 2,
	}
	elev3 := elevator.Elevator{
		ID: 3,
	}
	world := elevator.Worldview{
		Elevators: [3]Elevator{elev1, elev2, elev3},
	}

	go func() {
		for {
			helloTx <- world
			time.Sleep(3 * time.Second)
		}
	}()

	fmt.Println("Started")
	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

		case a := <-helloRx:
			fmt.Printf("Received: %#v\n", a)
			return a
		}
	}
}
