package network

import (
	"Driver-go/modules/network/bcast"
	"Driver-go/modules/network/localip"
	"Driver-go/modules/network/peers"
	"Driver-go/modules/worldview"
	"flag"
	"fmt"
	"os"
)

// Initializing network by fetching ID, and starting goroutines for UDP broadcast and peer tracking
func InitNetwork(peerUpdateCh chan peers.PeerUpdate,
	transmittWorldviewCh chan worldview.Worldview,
	recieveWorldviewCh chan worldview.Worldview) string {

	peerTxEnableCh := make(chan bool)
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
	go bcast.Transmitter(16569, transmittWorldviewCh)
	go bcast.Receiver(16569, recieveWorldviewCh)
	go peers.Transmitter(15647, id, peerTxEnableCh)
	go peers.Receiver(15647, peerUpdateCh)

	return id
}
