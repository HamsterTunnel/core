package tunnel

import (
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/HamsterTunnel/core/log"
)

type SingleServer struct {
	tunnelConn net.Conn
	userConn   net.Conn
	lock       sync.RWMutex
}

func NewSingleServer() *SingleServer {
	return &SingleServer{
		tunnelConn: nil,
		userConn:   nil,
	}
}

func (su *SingleServer) listenForClient(clientPort string) {
	clientListener, err := net.Listen("tcp", clientPort)
	if err != nil {
		log.Error(fmt.Sprintf("Unable to start listener on %s", clientPort), err)
	} else {
		log.Success(fmt.Sprintf("Listening for Client on %s", clientPort))
	}
	defer clientListener.Close()

	if clientConn, err := clientListener.Accept(); err != nil {
		log.Error(fmt.Sprintf("Unable to Accpet on %s", clientPort), err)
	} else {
		su.lock.Lock()
		su.tunnelConn = clientConn
		su.lock.Unlock()
		log.Success(fmt.Sprintf("Client connected on %s", clientPort))
	}
}

func (su *SingleServer) listenForUser(userPort string) error {
	userListener, err := net.Listen("tcp", userPort)
	if err != nil {
		log.Error(fmt.Sprintf("Unable to start listener on %s", userPort), err)
		return err
	} else {
		log.Success(fmt.Sprintf("Listening for User on %s", userPort))
	}
	defer userListener.Close()

	if userConn, err := userListener.Accept(); err != nil {
		log.Error(fmt.Sprintf("Unable to Accpet on %s", userPort), err)
		return err
	} else {
		su.lock.Lock()
		su.userConn = userConn
		su.lock.Unlock()
		log.Success(fmt.Sprintf("User connected on %s", userPort))
		return nil
	}
}

func (su *SingleServer) forwardData(dst net.Conn, src net.Conn, done chan<- error) {
	_, err := io.Copy(dst, src)
	if err != nil && err.Error() != "use of closed network connection" {
		log.Error(fmt.Sprintf("Error forwarding from %s to %s", src.RemoteAddr(), dst.RemoteAddr()), err)
	} else if err == nil {
		log.Message(fmt.Sprintf("Finished forwarding from %s to %s", src.RemoteAddr(), dst.RemoteAddr()))
	}
	done <- err
}

func (su *SingleServer) tunnelData() {
	done := make(chan error, 2)

	go su.forwardData(su.tunnelConn, su.userConn, done) // user ➝ client
	go su.forwardData(su.userConn, su.tunnelConn, done) // client ➝ user

	err := <-done
	if err != nil {
		log.Error("Tunnel closed with error", err)
	} else {
		log.Message("Tunnel closed gracefully.")
	}

	su.tunnelConn.Close()
	su.userConn.Close()
}

func (su *SingleServer) Start(userPort, clientPort string) {
	for {
		su.listenForClient(clientPort)
		su.listenForUser(userPort)
		su.tunnelData()
	}
}
