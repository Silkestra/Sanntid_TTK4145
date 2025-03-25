package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"time"
)

// Backup program ran in seperatly, receives heartbeat signals from primary and launches new primary in deadline exceeded
func backup() {
	ID := os.Args[1]
	port := os.Args[2]
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:800"+ID)
	if err != nil {
		fmt.Println("Error server address", err)
	}
	conn, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		fmt.Println("Error connecting", err)
	}
	defer conn.Close()
	for {
		deadline := time.Now().Add(10 * time.Second)
		_ = conn.SetReadDeadline(deadline)
		buffer := make([]byte, 1024)
		n, _, err := conn.ReadFromUDP(buffer)
		fmt.Println(string(buffer[:n]))
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				cmd := exec.Command("gnome-terminal", "--", "go", "run", "./main.go", "-id="+ID, port)
				cmd.Run()
				return

			}
		}
	}
}

func main() {
	backup()
}
