package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/Esteban-Bermudez/udpchat/pkg/udpchat"
	"github.com/joho/godotenv"
)

// CloudflareTURNResponse represents the structure of the JSON response from Cloudflare's TURN API.
type CloudflareTURNResponse struct {
	ICEServers []struct {
		URLs       []string `json:"urls"`
		Username   string   `json:"username"`
		Credential string   `json:"credential"`
	} `json:"iceServers"`
}

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: No .env file found or error loading it: %v", err)
	}

	// Define command-line flags
	useTurn := flag.Bool("turn", false, "Use TURN server for connection")
	turnServer := flag.String("turn-server", "", "TURN server address (e.g., turn.example.com:3478)")
	flag.Parse()

	if *useTurn {
		if *turnServer == "" {
			log.Fatalf("When using --turn, --turn-server is required.")
		}

		cloudflareAPIKey := os.Getenv("CLOUDFLARE_TURN_API_TOKEN")
		cloudflareTurnTokenID := os.Getenv("CLOUDFLARE_TURN_TOKEN_ID")

		if cloudflareAPIKey == "" || cloudflareTurnTokenID == "" {
			log.Fatalf("CLOUDFLARE_TURN_API_TOKEN and CLOUDFLARE_TURN_TOKEN_ID must be set in .env or environment variables for TURN.")
		}

		log.Println("Fetching TURN credentials from Cloudflare...")
		// Construct the URL for Cloudflare's TURN credential generation API
		apiURL := fmt.Sprintf("https://rtc.live.cloudflare.com/v1/turn/keys/%s/credentials/generate-ice-servers", cloudflareTurnTokenID)

		// Prepare the request body
		requestBody, err := json.Marshal(map[string]int{"ttl": 86400}) // TTL of 24 hours
		if err != nil {
			log.Fatalf("Failed to marshal request body: %v", err)
		}

		req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(requestBody))
		if err != nil {
			log.Fatalf("Failed to create HTTP request: %v", err)
		}

		req.Header.Set("Authorization", "Bearer "+cloudflareAPIKey)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 10 * time.Second}
		res, err := client.Do(req)
		if err != nil {
			log.Fatalf("Failed to make HTTP request to Cloudflare TURN API: %v", err)
		}
		defer res.Body.Close()

		bodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			log.Fatalf("Failed to read response body from Cloudflare TURN API: %v", err)
		}

		var turnResponse CloudflareTURNResponse
		if err := json.Unmarshal(bodyBytes, &turnResponse); err != nil {
			log.Fatalf("Failed to unmarshal Cloudflare TURN API response: %v", err)
		}

		if len(turnResponse.ICEServers) == 0 || turnResponse.ICEServers[0].Username == "" || turnResponse.ICEServers[0].Credential == "" {
			log.Fatalf("Cloudflare TURN API response did not contain valid credentials.")
		}

		turnUsername := turnResponse.ICEServers[0].Username
		turnPassword := turnResponse.ICEServers[0].Credential

		log.Println("Attempting to connect via TURN server...")
		turnConn, peerAddr, err := udpchat.ConnectTURN(*turnServer, turnUsername, turnPassword)
		if err != nil {
			log.Fatalf("Failed to connect via TURN: %v", err)
		}
		udpchat.StartTURN(turnConn, peerAddr)
	} else {
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

	// Keep the main goroutine alive until the chat is done
	// The chat functions (Start/StartTURN) run their read loops in goroutines
	// and block on stdin for sending. When stdin closes, they exit.
	// A simple way to keep the main goroutine alive for background listeners
	// is to wait for a signal or use a channel, but for a simple CLI chat,
	// the current blocking on stdin is sufficient. If the chat goroutines
	// were truly backgrounded and main exited, the program would terminate.
	// For now, the existing blocking behavior of Start/StartTURN is fine.

	// Exit cleanly after chat session (e.g., when stdin is closed)
	os.Exit(0)
}
