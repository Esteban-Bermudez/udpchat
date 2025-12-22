package udpchat

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/pion/stun"
	"github.com/pion/turn/v2"
)

// Connect uses a STUN server to discover the public address and prompts for the
// peer's address. It returns the resolved peer address.
func Connect(c *net.UDPConn) (*net.UDPAddr, error) {
	stunServerAddr := "stun.l.google.com:19302"
	publicAddr, err := publicAddr(c, stunServerAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to discover public address: %w", err)
	}
	log.Printf("My public address: %s", publicAddr)

	// Prompt for peer's address
	fmt.Print("Enter peer's address (ip:port): ")
	reader := bufio.NewReader(os.Stdin)
	peerAddrStr, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read peer address: %w", err)
	}
	peerAddrStr = strings.TrimSpace(peerAddrStr)

	fmt.Printf("Connecting to peer at %s\n", peerAddrStr)
	// Resolve peer's address
	peerAddr, err := net.ResolveUDPAddr("udp", peerAddrStr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peer address: %w", err)
	}

	return peerAddr, nil
}

// publicAddr contacts the STUN server to discover the public address of the UDP
// connection.
func publicAddr(c *net.UDPConn, stunAddr string) (*net.UDPAddr, error) {
	log.Printf("Contacting STUN server %s to find public address", stunAddr)

	raddr, err := net.ResolveUDPAddr("udp", stunAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve STUN server address: %w", err)
	}

	// Build the STUN Binding Request message
	msg := stun.MustBuild(stun.TransactionID, stun.BindingRequest)

	// Send the request to the STUN server
	_, err = c.WriteTo(msg.Raw, raddr)
	if err != nil {
		return nil, fmt.Errorf("failed to send STUN request: %w", err)
	}

	// Read the response
	buf := make([]byte, 1024)
	n, _, err := c.ReadFrom(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read STUN response: %w", err)
	}

	res := &stun.Message{Raw: buf[:n]}
	if err := res.Decode(); err != nil {
		return nil, fmt.Errorf("failed to decode STUN response: %w", err)
	}

	var xorMappedAddr stun.XORMappedAddress
	if err := xorMappedAddr.GetFrom(res); err != nil {
		return nil, fmt.Errorf("failed to get XOR-MAPPED-ADDRESS: %w", err)
	}

	publicAddr := &net.UDPAddr{IP: xorMappedAddr.IP, Port: xorMappedAddr.Port}

	return publicAddr, nil
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

// ConnectTURN initializes a TURN client and prompts for the peer's address.
// It returns the allocated PacketConn and the resolved peer address.
func ConnectTURN(turnServerAddr, username, password string) (net.PacketConn, *net.UDPAddr, error) {
	// 1. Create an underlying UDP connection for the TURN client to use.
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to listen on UDP for TURN client: %w", err)
	}
	defer conn.Close() // Ensure the connection is closed when done

	// 2. Create a TURN client
	client, err := turn.NewClient(
		&turn.ClientConfig{
			Conn:           conn,
			TURNServerAddr: turnServerAddr,
			Username:       username,
			Password:       password,
			// Realm:         "pion.ly", // Match your TURN server's realm if applicable
		},
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create TURN client: %w", err)
	}

	// 3. Start a goroutine to handle inbound packets for the TURN client.
	// This is crucial for the TURN client to process control messages and relayed data.
	// go func() {
	// 	buf := make([]byte, 1600) // UDP MTU
	// 	for {
	// 		n, from, err := conn.ReadFrom(buf)
	// 		if err != nil {
	// 			log.Printf("Error reading from underlying TURN connection: %v", err)
	// 			return // Exit goroutine on error
	// 		}

	// 		// Pass the received data to HandleInbound
	// 		handled, err := client.HandleInbound(buf[:n], from)
	// 		if err != nil {
	// 			log.Printf("Error handling inbound TURN packet: %v", err)
	// 			continue
	// 		}

	// 		if !handled {
	// 			// This means it was application data, but it should be handled by
	// 			// relayConn.ReadFrom
	// 			// This branch is mostly for debugging or if you were not using relayConn.
	// 			// For this chat app, all application data will come via relayConn.
	// 			// log.Printf("Received unhandled application data from %s: %s", from.String(),
	// 			// string(buf[:n]))
	// 		}
	// 	}
	// }()

	// 4. Allocate a relay address on the TURN server.
	// This `relayConn` is a net.PacketConn that represents your allocated port on the TURN server.
	relayConn, err := client.Allocate()
	if err != nil {
		client.Close()
		return nil, nil, fmt.Errorf("failed to allocate TURN relay: %w", err)
	}
	defer relayConn.Close() // Ensure the relay connection is closed when done
	log.Printf("TURN client allocated relayed address: %s", relayConn.LocalAddr().String())

	// Prompt for peer's address
	fmt.Print("Enter peer's address (ip:port): ")
	reader := bufio.NewReader(os.Stdin)
	peerAddrStr, err := reader.ReadString('\n')
	if err != nil {
		client.Close()
		return nil, nil, fmt.Errorf("failed to read peer address: %w", err)
	}
	peerAddrStr = strings.TrimSpace(peerAddrStr)

	fmt.Printf("Connecting to peer at %s via TURN\n", peerAddrStr)
	// Resolve peer's address
	peerAddr, err := net.ResolveUDPAddr("udp", peerAddrStr)
	if err != nil {
		client.Close()
		return nil, nil, fmt.Errorf("failed to resolve peer address: %w", err)
	}

	return relayConn, peerAddr, nil
}

// StartTURN handles the UDP chat logic using a TURN client.
func StartTURN(turnConn net.PacketConn, peerAddr *net.UDPAddr) {
	defer func() {
		if err := turnConn.Close(); err != nil {
			log.Printf("Error closing TURN connection: %v", err)
		}
	}()

	// Start a goroutine to listen for incoming messages on the relayed connection
	go func() {
		buf := make([]byte, 1024)
		for {
			n, addr, err := turnConn.ReadFrom(buf)
			if err != nil {
				// Check if the error is due to the connection being closed
				if strings.Contains(err.Error(), "use of closed network connection") {
					log.Println("TURN relayed connection closed, stopping read goroutine.")
					return
				}
				log.Printf("Error reading from TURN relayed connection: %v", err)
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
		_, err := turnConn.WriteTo([]byte(text), peerAddr) // Use WriteTo with peerAddr
		if err != nil {
			log.Printf("Error sending message via TURN: %v", err)
		}
		fmt.Print("\rYou: ")
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading from stdin: %v", err)
	}
}
