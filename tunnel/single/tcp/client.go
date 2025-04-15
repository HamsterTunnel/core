package tunnel

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/HamsterTunnel/core/log"
)

type SingleClient struct {
	tunnelConn net.Conn
	localConn  net.Conn
	lock       sync.RWMutex
}

func NewSingleClient() *SingleClient {
	return &SingleClient{
		tunnelConn: nil,
		localConn:  nil,
	}
}

func (sc *SingleClient) connectToServer(serverAddr string) error {
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		log.Error(fmt.Sprintf("Unable to connect to tunnel at %s", serverAddr), err)
		return err
	}
	sc.lock.Lock()
	sc.tunnelConn = conn
	sc.lock.Unlock()
	log.Success(fmt.Sprintf("Connected to tunnel at %s", serverAddr))
	return nil
}

func (sc *SingleClient) connectToLocal(localPort string) error {
	conn, err := net.Dial("tcp", localPort)
	if err != nil {
		log.Error(fmt.Sprintf("Unable to connect to local service at %s", localPort), err)
		return err
	}
	sc.lock.Lock()
	sc.localConn = conn
	sc.lock.Unlock()
	log.Success(fmt.Sprintf("Connected to local service at %s", localPort))
	return nil
}

func (sc *SingleClient) forwardData(dst net.Conn, src net.Conn, done chan<- error) {
	_, err := io.Copy(dst, src)
	if err != nil && err.Error() != "use of closed network connection" {
		log.Error(fmt.Sprintf("Error forwarding from %s to %s", src.RemoteAddr(), dst.RemoteAddr()), err)
	} else if err == nil {
		log.Message(fmt.Sprintf("Finished forwarding from %s to %s", src.RemoteAddr(), dst.RemoteAddr()))
	}
	done <- err
}

func (sc *SingleClient) tunnelData() {
	done := make(chan error, 2)

	go sc.forwardData(sc.tunnelConn, sc.localConn, done) // local ➝ tunnel
	go sc.forwardData(sc.localConn, sc.tunnelConn, done) // tunnel ➝ local

	err := <-done
	if err != nil {
		log.Error("Tunnel closed with error", err)
	} else {
		log.Message("Tunnel closed gracefully.")
	}

	sc.tunnelConn.Close()
	sc.localConn.Close()
}

func (sc *SingleClient) Start(tunnelAddr, localAddr string) {
	for {
		if err := sc.connectToServer(tunnelAddr); err != nil {
			time.Sleep(2 * time.Second)
			continue
		}
		if err := sc.connectToLocal(localAddr); err != nil {
			sc.tunnelConn.Close()
			time.Sleep(2 * time.Second)
			continue
		}
		sc.tunnelData()
	}
}

