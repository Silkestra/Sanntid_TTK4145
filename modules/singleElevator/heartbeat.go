package singleElevator

import (
	"fmt"
	"net"
	"os/exec"
	"time"
)

func HeartbeatToBackup(ID string, port string) {
	//Create backup
	cmd := exec.Command("gnome-terminal", "--", "go", "run", "./backup/backup.go", ID, port)
	cmd.Run()

	//Network conn
	serverAddr, err := net.ResolveUDPAddr("udp", "localhost:800"+ID)
	if err != nil {
		fmt.Println("Error server address", err)
		return
	}

	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		fmt.Println("Error connecting", err)
		return
	}
	defer conn.Close()

	for {
		time.Sleep(1 * time.Second)
		fmt.Println("Attempting to send data...")
		_, err = conn.Write([]byte("hei"))
		if err != nil {
			fmt.Println("Error sending:", err)
		} else {
			fmt.Println("Data sent successfully")
		}

	}
}
