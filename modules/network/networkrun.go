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
	peerTxEnableCh chan bool,
	transmittWorldviewCh chan worldview.Worldview,
	recieveWorldviewCh chan worldview.Worldview) string { 
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
	go bcast.Transmitter(16666, transmittWorldviewCh)
	go bcast.Receiver(16666, recieveWorldviewCh)
	go peers.Transmitter(15555, id, peerTxEnableCh)
	go peers.Receiver(15555, peerUpdateCh)

	return id
}
