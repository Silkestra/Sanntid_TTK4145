package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"time"
)

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

// func backup() {
// 	ID := os.Args[1]
// 	port := os.Args[2]

// 	// Set up UDP connection
// 	serverAddr, err := net.ResolveUDPAddr("udp", "localhost:800"+ID)
// 	if err != nil {
// 		fmt.Println("Error resolving server address:", err)
// 		return
// 	}
// 	conn, err := net.DialUDP("udp", nil, serverAddr)
// 	if err != nil {
// 		fmt.Println("Error connecting:", err)
// 		return
// 	}
// 	defer conn.Close()

// 	// Track the last received message time
// 	lastMessageTime := time.Now()

// 	for {
// 		// Calculate deadline (10 seconds after the last received message)
// 		deadline := lastMessageTime.Add(10 * time.Second)
// 		err = conn.SetReadDeadline(deadline)
// 		if err != nil {
// 			fmt.Println("Error setting read deadline:", err)
// 			return
// 		}

// 		// Read incoming data
// 		buffer := make([]byte, 1024)
// 		n, err := conn.Read(buffer)
// 		if err != nil {
// 			// If we hit a timeout (no message received in the last 10 seconds)
// 			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
// 				// Check if 10 seconds have passed since the last message
// 				if time.Since(lastMessageTime) >= 10*time.Second {
// 					// No message for 10 seconds, trigger the action
// 					fmt.Println("Timeout reached without receiving data for 10 seconds. Launching new process...")
// 					cmd := exec.Command("gnome-terminal", "--", "go", "run", "../main.go", "-id="+ID, port)
// 					cmd.Run()
// 					return
// 				}
// 			} else {
// 				// Handle other errors (e.g., network issues)
// 				fmt.Println("Error reading:", err)
// 			}
// 		} else {
// 			// Successfully received a message, update the last received time
// 			lastMessageTime = time.Now()
// 			fmt.Println("hei")
// 			fmt.Println("Received data:", string(buffer[:n]))
// 		}
// 	}
// }

func main() {
	backup()
}
