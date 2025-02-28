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

func InitNetwork(peerUpdateCh chan peers.PeerUpdate, //init og runnework deles for å unngå go i go
	peerTxEnable chan bool,
	transmittWorldView chan worldview.Worldview,
	recieveWorldView chan worldview.Worldview) string { //network init function that inits tansmission and peer heartbeat check
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
