package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) < 4 {
		log.Fatal("Usage: udpchat <my-port> <connect-port> <ip>")
	}

	p := os.Args[1]
	cp := os.Args[2]
	conIP := net.ParseIP(os.Args[3])

	port, err := strconv.Atoi(p)
	if err != nil {
		log.Fatalf("Invalid port number: %v", err)
	}
	conPort, err := strconv.Atoi(cp)
	if err != nil {
		log.Fatalf("Invalid port number: %v", err)
	}

	fmt.Println("UDP Chat Server", net.IPv4zero)
	c, err := net.ListenUDP("udp", &net.UDPAddr{
		Port: port,
		IP:   net.IPv4zero,
	})
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", p, err)
	}
	log.Printf("Listening on port %s", p)

	c.SetReadBuffer(1024 * 1024)
	// Receive UDP messages
	go func() {
		for {
			buf := make([]byte, 1024)
			n, addr, err := c.ReadFromUDP(buf)
			if err != nil {
				log.Printf("Error reading from UDP: %v", err)
				continue
			}
			fmt.Printf("\r%s: %s\n", addr, string(buf[:n]))
			fmt.Print("\rYou: ")
		}
	}()

	// Send UDP messages
	for {
		var msg string
		fmt.Print("You: ")
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			msg = scanner.Text()
		}
		addr := &net.UDPAddr{
			Port: conPort,
			IP: conIP,
		}
		if _, err := c.WriteToUDP([]byte(msg), addr); err != nil {
			log.Printf("Error sending UDP message: %v", err)
		}
	}
}
