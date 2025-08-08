package main

import (
	"log"
	"net"

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

	// Connect to peer and start chat
	peerAddr, err := udpchat.Connect(c)
	if err != nil {
		log.Fatalf("Failed to connect to peer: %v", err)
	}
	udpchat.Start(c, peerAddr)
}
