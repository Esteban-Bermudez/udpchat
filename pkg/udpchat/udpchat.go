package udpchat

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/pion/stun"
)

// Connect uses a STUN server to discover the public address and prompts for the
// peer's address. It returns the resolved peer address.
func Connect(c *net.UDPConn) *net.UDPAddr {
	stunServerAddr := "stun.l.google.com:19302"
	publicAddr := publicAddr(c, stunServerAddr)
	log.Printf("My public address: %s", publicAddr)

	// Prompt for peer's address
	fmt.Print("Enter peer's address (ip:port): ")
	reader := bufio.NewReader(os.Stdin)
	peerAddrStr, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Failed to read peer address: %v", err)
	}
	peerAddrStr = strings.TrimSpace(peerAddrStr)

	fmt.Printf("Connecting to peer at %s\n", peerAddrStr)
	// Resolve peer's address
	peerAddr, err := net.ResolveUDPAddr("udp", peerAddrStr)
	if err != nil {
		log.Fatalf("Failed to resolve peer address: %v", err)
	}

	return peerAddr
}

// publicAddr contacts the STUN server to discover the public address of the UDP
// connection.
func publicAddr(c *net.UDPConn, stunAddr string) *net.UDPAddr {
	log.Printf("Contacting STUN server %s to find public address", stunAddr)

	raddr, err := net.ResolveUDPAddr("udp", stunAddr)
	if err != nil {
		log.Fatalf("Failed to resolve STUN server address: %v", err)
	}

	// Build the STUN Binding Request message
	msg := stun.MustBuild(stun.TransactionID, stun.BindingRequest)

	// Send the request to the STUN server
	_, err = c.WriteTo(msg.Raw, raddr)
	if err != nil {
		log.Fatalf("Failed to send STUN request: %v", err)
	}

	// Read the response
	buf := make([]byte, 1024)
	n, _, err := c.ReadFrom(buf)
	if err != nil {
		log.Fatalf("Failed to read STUN response: %v", err)
	}

	res := &stun.Message{Raw: buf[:n]}
	if err := res.Decode(); err != nil {
		log.Fatalf("Failed to decode STUN response: %v", err)
	}

	var xorMappedAddr stun.XORMappedAddress
	if err := xorMappedAddr.GetFrom(res); err != nil {
		log.Fatalf("Failed to get XOR-MAPPED-ADDRESS from STUN response: %v", err)
	}

	publicAddr := &net.UDPAddr{IP: xorMappedAddr.IP, Port: xorMappedAddr.Port}

	return publicAddr
}

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
