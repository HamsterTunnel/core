package tunnel

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

func TestSingleClientTunnel_ForwardAndReconnect(t *testing.T) {
	tunnelPort := ":9100"
	localPort := ":9101"

	go func() {
		for {
			listener, err := net.Listen("tcp", localPort)
			if err != nil {
				t.Logf("Local service listen error: %v", err)
				time.Sleep(200 * time.Millisecond)
				continue
			}

			conn, err := listener.Accept()
			if err != nil {
				t.Logf("Local service accept error: %v", err)
				listener.Close()
				continue
			}

			go func(c net.Conn) {
				defer c.Close()
				reader := bufio.NewReader(c)
				for {
					line, err := reader.ReadString('\n')
					if err != nil {
						return
					}
					response := fmt.Sprintf("Echo: %s", line)
					c.Write([]byte(response))
				}
			}(conn)
			listener.Close()
		}
	}()

	go func() {
		sc := NewSingleClient()
		sc.Start(tunnelPort, localPort)
	}()

	time.Sleep(500 * time.Millisecond)

	for i := range 2 {
		serverListener, err := net.Listen("tcp", tunnelPort)
		if err != nil {
			t.Fatalf("Tunnel server listen failed: %v", err)
		}

		t.Logf("Waiting for client tunnel to connect (attempt %d)", i+1)
		tunnelConn, err := serverListener.Accept()
		if err != nil {
			t.Fatalf("Failed to accept tunnel connection (attempt %d): %v", i+1, err)
		}
		reader := bufio.NewReader(tunnelConn)

		message := fmt.Sprintf("Hello %d\n", i)
		_, err = tunnelConn.Write([]byte(message))
		if err != nil {
			t.Fatalf("Failed to write to tunnel (attempt %d): %v", i+1, err)
		}

		response, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("Failed to read response from tunnel (attempt %d): %v", i+1, err)
		}

		expected := fmt.Sprintf("Echo: %s", message)
		if strings.TrimSpace(response) != strings.TrimSpace(expected) {
			t.Errorf("Mismatch on attempt %d: expected '%s', got '%s'", i+1, expected, response)
		}

		tunnelConn.Close()
		serverListener.Close()

		time.Sleep(500 * time.Millisecond)
	}
}
