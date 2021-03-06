package connector

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/sync/errgroup"
	"log"
	"sync"
)

const (
	// ClientStateConnected represents a new connection that is waiting for the auth message from the peer.
	// A Client instance begins at this state and then transition to either ClientStateAuthorized or ClientStateClosed.
	ClientStateConnected ClientState = iota
	ClientStateAuthorized
	// ClientStateClosed represents a closed connection. This is a terminal state.
	// After entering this state, a Client instance will not receive any message and can not send any message.
	ClientStateClosed
)

var (
	ErrExceedMaxClients = errors.New("ppcserver: exceed maximum number of clients")
)

type (
	// ClientState represents the state of a Client instance, uint8 is used for save memory usage.
	ClientState uint8

	// Client represents a Client connection to a server.
	Client struct {
		transport Transport
		mu        sync.Mutex         // mu guards state.
		state     ClientState        // state is guarded by mu.
		cancelCtx context.CancelFunc // cancelCtx cancels the Client-level context that creates inside StartClient and result in Client.Close() being called.
		readCh    chan []byte
		writeCh   chan []byte // writeCh is the buffered channel of messages waiting to write to the transport.
	}
)

// StartClient creates a new Client with ClientStateConnected as the initial state,
//
func StartClient(ctx context.Context, transport Transport) error {
	if ExceedMaxClients() {
		return ErrExceedMaxClients
	}
	incrNumClients()
	defer decrNumClients()

	// The ctx.Done channel returns from context.WithCancel() is closed when the cancelCtx() function is called
	// or when the parent context's Done channel is closed, whichever happens first.
	ctx, cancelCtx := context.WithCancel(ctx)
	defer cancelCtx() // Call cancelCtx when StartClient exits to ensure the current Client's resources are fully released.

	c := &Client{
		transport: transport,
		state:     ClientStateConnected,
		cancelCtx: cancelCtx,
		readCh:    make(chan []byte),      // TODO, what is the buffer size?
		writeCh:   make(chan []byte, 256), // TODO, buffer size is configurable
	}

	// if !allowToConnect() {
	// 	return
	// }
	// if err := handshake(); err != nil {
	// 	return
	// }

	// The ctx.Done channel returns from errgroup.WithContext() will be closed
	// when the first time either writeLoop or readLoop passed to g.Go() returns a non-nil error,
	// or g.Wait() returns, whichever occurs first.
	g, ctx := errgroup.WithContext(ctx)
	// Since per connection support only one concurrent reader and one concurrent writer,
	// we execute all writes from the `writeLoop` goroutine and all reads from the `readLoop` goroutine.
	// Reference https://pkg.go.dev/github.com/gorilla/websocket#hdr-Concurrency for the concurrency usage details.
	g.Go(
		func() error {
			return c.writeLoop(ctx)
		},
	)
	g.Go(
		func() error {
			return c.readLoop()
		},
	)

	// Actively close the connection when ctx.Done channel is closed to force readLoop exits.
	<-ctx.Done()
	_ = c.Close()

	// Block until both readLoop and writeLoop exit to achieve a graceful shutdown of the Client.
	// The g.Wait() will return the first error that causes the blocking exits.
	return g.Wait()
}

// Close first mutates Client to the ClientStateClosed state,
// then closes the underlying transport connection with the peer.
// Close does nothing if the Client's state is already ClientStateClosed.
func (c *Client) Close() (err error) {
	defer func() {
		if err != nil {
			log.Println("ppcserver: Client.Close() error:", err)
			return
		}
		log.Println("ppcserver: Client.Close() complete")
	}()

	// Change to the closed state should be guarded by mu. Skip if already in the closed state.
	c.mu.Lock()
	if c.state == ClientStateClosed {
		c.mu.Unlock()
		return nil
	}
	c.state = ClientStateClosed
	c.mu.Unlock()

	// TODO, should send close message

	// c.cancelCtx()

	// transport.Close() closes the underlying network connection.
	// It can be called concurrently, and it's OK to call Close more than once.
	return c.transport.Close()
}

// readLoop keep reading from the transport until transport.Read() errored.
// The only reason readLoop exits is an error returns from transport.Read().
// readLoop must execute by a single goroutine to ensure that there is at most one concurrent reader on a connection.
func (c *Client) readLoop() error {
	// The readLoop method is the only sender on readCh,
	// so we Close the readCh here to ensure not sending on the closed readCh channel.
	defer close(c.readCh)

	for {
		// TODO, here we actually use read timeout to break the loop

		message, err := c.transport.Read()

		// The connection must be closed once Read returns any error.
		if err != nil {
			return fmt.Errorf("ppcserver: Client.transport.Read() error: %w", err)
		}

		log.Printf("ppcserver: Client.transport.Read() receive: %s", message)

		// TODO, send to readCh, block when readCh is full
		// case c.readCh <- message:
	}

	// TODO, wait auth request from the peer.
}

func (c *Client) writeLoop(ctx context.Context) error {
	// ticker := time.NewTicker(pingPeriod)
	// defer ticker.Stop()

	return nil
}

// State returns the current state of the Client.
func (c *Client) State() ClientState {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.state
}

func (c *Client) Write(data []byte) error {
	select {
	case c.writeCh <- data:
	default:
		// TODO, do we need to close the writeCh when its buffer is full?
		close(c.writeCh)
	}
	return nil
}

func (c *Client) heartbeat() {

}
