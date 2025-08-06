package udpchat

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

// Start handles the UDP chat logic.
func Start(conn *net.UDPConn, peerAddr *net.UDPAddr) {
	// Start a goroutine to listen for incoming messages
	go func() {
		buf := make([]byte, 1024)
		for {
			n, addr, err := conn.ReadFromUDP(buf)
			if err != nil {
				log.Printf("Error reading from UDP: %v", err)
				continue
			}
			// Print received message
			fmt.Printf("\r%s: %s\n", addr, string(buf[:n]))
			fmt.Print("You: ")
		}
	}()

	// Read from stdin and send messages to the peer
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("You: ")
	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			continue
		}
		_, err := conn.WriteToUDP([]byte(text), peerAddr)
		if err != nil {
			log.Printf("Error sending message: %v", err)
		}
		fmt.Print("\rYou: ")
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading from stdin: %v", err)
	}
}
