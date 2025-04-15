package tunnel

import (
	"errors"
	"io"
	"log"
	"net"
)

func forwardData(dst, src net.Conn, direction string, done chan<- error) {
	_, err := io.Copy(dst, src)
	if err != nil && !errors.Is(err, net.ErrClosed) {
		log.Printf("Error forwarding data %s: %v", direction, err)
		done <- err
	} else {
		log.Printf("Connection closed normally %s", direction)
		done <- nil
	}
}

func StartLocalTCPTunnel(remotePort, servicePort string) {
	for {
		log.Println("Attempting to connect to remote and local services...")

		// Try to connect to remote service
		remoteConn, err := net.Dial("tcp", remotePort)
		if err != nil {
			log.Printf("Unable to connect to remote proxy on port %s: %v", remotePort, err)
			// Retry immediately if connection fails
			continue
		}
		log.Printf("Connected to remote proxy: %s", remotePort)

		// Try to connect to local service
		serviceConn, err := net.Dial("tcp", servicePort)
		if err != nil {
			log.Printf("Error connecting to local service on port %s: %v", servicePort, err)
			remoteConn.Close() // Close remoteConn as we failed to connect to the service
			// Retry immediately if connection fails
			continue
		}
		log.Printf("Connected to local service: %s", servicePort)

		// Channel to receive the results of the two forwarding directions
		errChan := make(chan error, 2)

		// Start forwarding data in both directions
		go forwardData(serviceConn, remoteConn, "remote -> service", errChan)
		go forwardData(remoteConn, serviceConn, "service -> remote", errChan)

		// Wait for one of the connections to close and handle reconnection
		var finalErr error
		for {
			if err := <-errChan; err != nil {
				finalErr = err
				break // Exit the loop as one of the connections has failed or closed
			}
		}

		// Close connections after forwarding is done
		serviceConn.Close()
		remoteConn.Close()

		// Log the final outcome and retry if needed
		if finalErr != nil {
			log.Printf("Forwarding terminated with error: %v", finalErr)
		} else {
			log.Println("Forwarding terminated normally")
		}

		// After the forwarding has finished, attempt to reconnect
		log.Println("Attempting to reconnect...")
	}
}
