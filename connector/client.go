package connector

import (
	"github.com/gorilla/websocket"
	"log"
)

const (
	// ClientReadyStateConnected is the readyState when the connection has established,
	// and the server is waiting for the auth message from the peer.
	ClientReadyStateConnected clientReadyState = iota
	ClientReadyStateAuthorized
	ClientReadyStateClosed
)

type (
	// clientReadyState represents different readyStates of the Client, uint8 is used for save memory usage.
	clientReadyState uint8

	// Client represents a client connection to any server of the connector.
	Client struct {
		transport  Transport
		readyState clientReadyState
	}
)

// NewClient creates a new Client.
func NewClient(transport Transport) *Client {
	client := &Client{
		transport:  transport,
		readyState: ClientReadyStateConnected,
	}

	return client
}

func (c *Client) Write(data []byte) error {
	if err := c.transport.Write(data); err != nil {
		return err
	}
	return nil
}

func (c *Client) Close() error {

	// TODO, notify that client is closed.

	return nil
}
