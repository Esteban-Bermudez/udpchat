package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/pion/stun"
	"github.com/Esteban-Bermudez/udpchat/pkg/udpchat"
)

func main() {
	c, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4allrouter, // Listen on all interfaces
	})
	if err != nil {
		log.Fatalf("Failed to listen on UDP: %v", err)
	}
	log.Printf("Listening internally on %s\n", c.LocalAddr())

	stunServerAddr := "stun.l.google.com:19302"
	log.Printf("Contacting STUN server %s to find public address", stunServerAddr)

	raddr, err := net.ResolveUDPAddr("udp", stunServerAddr)
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

	// Start chat
	udpchat.Start(c, peerAddr)
}
