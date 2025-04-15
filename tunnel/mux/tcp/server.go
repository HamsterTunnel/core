package tcpmux

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/HamsterTunnel/core/log"
)

type User struct {
	id   uint32
	conn net.Conn
}

type SMultiplexer struct {
	isConnected bool
	tunnelConn  net.Conn
	users       map[uint32]*User
	lock        sync.RWMutex
	nextID      uint32
	writeChan   chan []byte
}

func NewSMultiplexer() *SMultiplexer {
	return &SMultiplexer{
		users:     make(map[uint32]*User),
		writeChan: make(chan []byte, 100),
		nextID:    1,
	}
}

func (m *SMultiplexer) Start(userPort, clientPort string) {
	if tunnelConn, err := listen(clientPort, "client"); err != nil {
		log.Error(fmt.Sprintf("Client fail to connect on port %s", clientPort), err)
	} else {
		log.Success(fmt.Sprintf("Client connected on port %s", clientPort))
		m.lock.Lock()
		m.tunnelConn = tunnelConn
		m.isConnected = true
		m.lock.Unlock()
		go m.readTunnel()
		go m.writeTunnel()
		for {
			if userConn, err := listen(userPort, "user"); err != nil {
				log.Error(fmt.Sprintf("User fail to connect on port %s", userPort), err)
			} else {
				m.lock.Lock()
				user := &User{id: m.nextID, conn: userConn}
				m.users[m.nextID] = user
				log.Success(fmt.Sprintf("User connected with id: %d", m.nextID))
				m.nextID++
				m.lock.Unlock()
				go m.readUserConnection(user)
			}
		}
	}
}

//func (m *SMultiplexer) Stop() {
//
//}

func (m *SMultiplexer) readTunnel() {
	buf := make([]byte, 16*1024)

	for {
		n, err := m.tunnelConn.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Error("Error reading from tunnel connection", err)
			}
			break
		}

		if n < 4 {
			log.Warning("Received packet too small to contain user ID")
			continue
		}

		userID := binary.BigEndian.Uint32(buf[:4])
		payload := make([]byte, n-4)
		copy(payload, buf[4:n])

		m.lock.RLock()
		user, exists := m.users[userID]
		m.lock.RUnlock()

		if !exists {
			log.Warning(fmt.Sprintf("User ID %d not found", userID))
			continue
		}

		_, err = user.conn.Write(payload)
		if err != nil {
			log.Error(fmt.Sprintf("Error writing to user connection ID %d", userID), err)
			_ = user.conn.Close()

			m.lock.Lock()
			delete(m.users, userID)
			m.lock.Unlock()
		}
	}
}

func (m *SMultiplexer) writeTunnel() {
	for {
		select {
		case packet := <-m.writeChan:
			m.lock.RLock()
			conn := m.tunnelConn
			m.lock.RUnlock()

			if conn == nil {
				log.Warning("Tunnel connection is nil. Dropping packet.")
				continue
			}

			_, err := conn.Write(packet)
			if err != nil {
				log.Error("Error writing to tunnel connection", err)

				m.lock.Lock()
				_ = m.tunnelConn.Close()
				m.tunnelConn = nil
				m.isConnected = false
				m.lock.Unlock()
			}
		}
	}
}

func listen(port, subject string) (net.Conn, error) {
	listener, err := net.Listen("tcp", port)

	if err != nil {
		log.Error(fmt.Sprintf("Unable to listen for %s on port %s", subject, port), err)
		return nil, err
	}
	defer listener.Close()

	conn, err := listener.Accept()
	if err != nil {
		log.Error(fmt.Sprintf("Unable to accept %s connection on port %s", subject, port), err)
		return nil, err
	}
	return conn, nil
}

func (m *SMultiplexer) readUserConnection(user *User) {
	buf := make([]byte, 16*1024) // 16 KB buffer

	for {
		n, err := user.conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Error("Error reading from user connection", err)
			}
			break
		}

		packet := make([]byte, 8+n)
		binary.BigEndian.PutUint32(packet[:4], user.id)
		binary.BigEndian.PutUint32(packet[4:8], uint32(n))
		copy(packet[8:], buf[:n])

		m.writeChan <- packet
	}
}
