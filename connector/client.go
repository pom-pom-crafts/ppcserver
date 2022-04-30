package connector

import (
	"context"
	"log"
	"sync"
)

const (
	// ClientStateConnected is the state when the connection has established,
	// and the server is waiting for the auth message from the peer.
	ClientStateConnected ClientState = iota
	ClientStateAuthorized
	ClientStateClosed
)

type (
	// ClientState represents the state of the client, uint8 is used for save memory usage.
	ClientState uint8

	// Client represents a client connection to a server.
	Client interface {
		Write(data []byte) error
		Close() error
	}

	client struct {
		transport Transport
		mu        sync.Mutex  // mu guards state.
		state     ClientState // state is guarded by mu.
		readCh    chan []byte
		writeCh   chan []byte // writeCh is the buffered channel of messages waiting to write to the transport.
	}
)

// newClient creates a new client.
func newClient(transport Transport) *client {
	c := &client{
		transport: transport,
		state:     ClientStateConnected,
		readCh:    make(chan []byte),
		writeCh:   make(chan []byte, 256), // TODO, buffer size is configurable
	}

	return c
}

func (c *client) Write(data []byte) error {
	if err := c.transport.Write(data); err != nil {
		return err
	}
	return nil
}

// Close closes the connection with the peer.
func (c *client) Close() (err error) {
	defer func() {
		if err != nil {
			log.Println("ppcserver: client.Close() error:", err)
			return
		}
		log.Println("ppcserver: client.Close()")
	}()

	// Change to the closed state should be guarded by mu. Skip if already in the closed state.
	c.mu.Lock()
	if c.state == ClientStateClosed {
		c.mu.Unlock()
		return nil
	}
	c.state = ClientStateClosed
	c.mu.Unlock()

	// Close the readCh to notify readers to stop reading from it.
	close(c.readCh)

	// transport.Close() closes the underlying network connection.
	// It can be called concurrently, and it's OK to call Close more than once.
	return c.transport.Close()
}

// func (c *client) handleConnection() {
// 	if !allowToConnect() {
// 		c.Close()
// 		return
// 	}
// 	if err := handshake(); err != nil {
// 		c.Close()
// 		return
// 	}
// 	go c.readLoop()
// 	go c.writeLoop()
// }

// readLoop.
// The application must runs readLoop in a per-connection goroutine.
// The application ensures that there is at most one reader on a connection by executing all reads from this goroutine.
func (c *client) readLoop(ctx context.Context) {
	defer c.Close()

	for {
		select {
		case <-ctx.Done():
			return // Caution: use 'return' instead of 'break' to exit the for loop.
		default:
			message, err := c.transport.Read()

			// Exit readLoop once Read returns any error.
			if err != nil {
				log.Printf("ppcserver: client.transport.Read() error: %v", err)
				return // Caution: use 'return' instead of 'break' to exit the for loop.
			}

			log.Printf("ppcserver: client.transport.Read() receive: %v", message)
			// TODO, send to readCh
			// c.readCh <- message
		}
	}

	// TODO, wait auth request from the peer.
}

func (c *client) writeLoop() {

}
