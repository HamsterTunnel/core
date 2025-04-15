package tcpmux

import (
	"encoding/binary"
	"io"
	"net"
	"testing"
	"time"

	"github.com/HamsterTunnel/core/log"
)

func readTunnel(conn net.Conn) (uint32, []byte, error) {
	// A read timeout
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))

	header := make([]byte, 8)
	_, err := io.ReadFull(conn, header)
	if err != nil {
		return 0, nil, err
	}

	userID := binary.BigEndian.Uint32(header[:4])
	dataLen := binary.BigEndian.Uint32(header[4:8])

	if dataLen > 16*1024 {
		return 0, nil, io.ErrUnexpectedEOF
	}

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))

	data := make([]byte, dataLen)
	_, err = io.ReadFull(conn, data)
	if err != nil {
		return 0, nil, err
	}

	return userID, data, nil
}

func TestSMultiplexer(t *testing.T) {
	m := NewSMultiplexer()
	go m.Start(":9090", ":9091")

	// Aspetta che il server inizi
	time.Sleep(1 * time.Second)

	// Connessione al client e agli utenti
	clientConn, err := net.Dial("tcp", "localhost:9091")
	if err != nil {
		t.Fatalf("Unable to connect to the client: %v", err)
	}

	user1Conn, err := net.Dial("tcp", "localhost:9090")
	if err != nil {
		t.Fatalf("Unable to connect to user1: %v", err)
	}

	user2Conn, err := net.Dial("tcp", "localhost:9090")
	if err != nil {
		t.Fatalf("Unable to connect to user2: %v", err)
	}

	user3Conn, err := net.Dial("tcp", "localhost:9090")
	if err != nil {
		t.Fatalf("Unable to connect to user3: %v", err)
	}

	// Nel test, invio i messaggi con padding:
	string1 := "Im the 1"
	string2 := "Im the 2"
	string3 := "Im the 3"

	// Invia i pacchetti con i messaggi e il padding
	user1Conn.Write([]byte(string1))
	user2Conn.Write([]byte(string2))
	user3Conn.Write([]byte(string3))
	time.Sleep(2 * time.Second)

	receivedMessages := make(map[uint32][]byte)

	for {
		userID, data, err := readTunnel(clientConn)
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				log.Info("Timeout: connection empty for 2 seconds.")
				break // o continua, a seconda di cosa vuoi fare
			} else if err == io.EOF {
				log.Info("Connection close from client.")
				break
			} else {
				log.Error("Reading error:", err)
				break
			}
		}
		receivedMessages[userID] = append(receivedMessages[userID], data...)
	}
	if string(receivedMessages[1]) != string1 {
		t.Errorf("utente1 send: %s , recived: %s", string1, receivedMessages[1])
	}
	if string(receivedMessages[2]) != string2 {
		t.Errorf("utente1 send: %s , recived: %s", string2, receivedMessages[2])
	}
	if string(receivedMessages[3]) != string3 {
		t.Errorf("utente1 send: %s , recived: %s", string3, receivedMessages[3])
	}
}
