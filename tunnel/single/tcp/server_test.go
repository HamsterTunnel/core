package tunnel

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

func TestSingleUserTunnel_ForwardAndReconnect(t *testing.T) {
	userPort := ":9001"
	clientPort := ":9002"

	su := NewSingleServer()

	// Avvia il server tunnel in background
	go su.Start(userPort, clientPort)

	// Attendi che il server sia in ascolto
	time.Sleep(500 * time.Millisecond)

	for i := range 2 {
		// Simula il client che resta connesso
		clientConn, err := net.Dial("tcp", clientPort)
		if err != nil {
			t.Fatalf("Client tunnel connection failed: %v", err)
		}
		defer clientConn.Close()

		reader := bufio.NewReader(clientConn)

		// Due iterazioni in cui l'utente si connette, invia un messaggio e si disconnette
		time.Sleep(500 * time.Millisecond)
		userConn, err := net.Dial("tcp", userPort)
		if err != nil {
			t.Fatalf("User connection failed (attempt %d): %v", i+1, err)
		}

		message := fmt.Sprintf("Hello from user %d \n", i)
		_, err = userConn.Write([]byte(message))
		if err != nil {
			t.Fatalf("Failed to write message from user (attempt %d): %v", i+1, err)
		}

		// Leggi il messaggio lato tunnel (client)
		received, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("Client failed to read message (attempt %d): %v", i+1, err)
		}

		if strings.TrimSpace(received) != strings.TrimSpace(message) {
			t.Errorf("Mismatch on attempt %d: expected '%s', got '%s'", i+1, message, received)
		}

		userConn.Close()

		// Dai tempo al tunnel di rilevare la disconnessione
		time.Sleep(500 * time.Millisecond)
	}
}
