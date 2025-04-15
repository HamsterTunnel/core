package tunnel

import (
	"io"
	"log"
	"net"
)

func forwarData(clientConn, userConn net.Conn, done chan<- error) {
	_, err := io.Copy(clientConn, userConn)
	if err != nil {
		if err.Error() == "use of closed network connection" {
			log.Println("Connection closed by user while forwarding data to client.")
		} else {
			log.Printf("Error forwarding data from user to client: %v", err)
		}
		done <- err // Invia l'errore (o nil) al canale
		return
	}
	log.Println("Data forwarding from user to client completed.")
	done <- nil // Indica che tutto è andato a buon fine
}

func handleClientConnection(clientConn net.Conn, userListener net.Listener) {
	defer clientConn.Close()

	for {
		userConn, err := userListener.Accept()
		if err != nil {
			log.Printf("Error accepting connection from user: %v", err)
			continue
		}

		log.Println("Connection established with user.")
		done := make(chan error, 2) // Canale per segnare la fine di entrambe le operazioni di forwarding

		// Forward dati in entrambe le direzioni
		go forwarData(clientConn, userConn, done)
		go forwarData(userConn, clientConn, done)

		// Attendi la chiusura di entrambe le direzioni
		var finalErr error
		for i := 0; i < 2; i++ {
			err := <-done
			if err != nil && finalErr == nil {
				finalErr = err // Se c'è un errore, lo memorizziamo
			}
		}

		// Dopo che entrambe le operazioni sono terminate, riavvia l'ascolto
		if finalErr != nil {
			log.Printf("Forwarding terminated with error: %v", finalErr)
		} else {
			log.Println("Forwarding completed successfully.")
		}
	}
}

func StartRemoteTCPTunnel(clientPort, userPort string) error {
	clientListener, err := net.Listen("tcp", ":"+clientPort)
	if err != nil {
		log.Fatalf("Unable to start listener on public port %s: %v", clientPort, err)
	}
	defer clientListener.Close()

	userListener, err := net.Listen("tcp", ":"+userPort)
	if err != nil {
		log.Fatalf("Unable to start listener on user port %s: %v", userPort, err)
	}
	defer userListener.Close()

	log.Printf("Remote proxy listening on ports %s (client) and %s (user)", clientPort, userPort)

	for {
		clientConn, err := clientListener.Accept()
		if err != nil {
			log.Printf("Error accepting connection from client: %v", err)
			continue
		}

		log.Println("Connection established with client.")

		// Gestisci la connessione del client
		go handleClientConnection(clientConn, userListener)
	}
}
