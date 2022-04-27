package connector

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
	}

	client struct {
		transport Transport
		state     ClientState
		writeCh   chan []byte // writeCh is the buffered channel of messages waiting to write to the transport.
	}
)

// NewClient creates a new Client.
func NewClient(transport Transport) Client {
	c := &client{
		transport: transport,
		state:     ClientStateConnected,
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

func (c *client) Close() error {

	// TODO, notify that client is closed.

	return nil
}

func (c *client) readLoop() {
	// TODO, notify client to remove it from WebSocketServer.
	// c.hub.unregister <- c
	defer c.Close()

	// for {
	// 	_, message, err := t.conn.ReadMessage()
	//
	// 	// Exit readLoop on ReadMessage returns any error.
	// 	if err != nil {
	// 		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
	// 			log.Printf("ppcserver: websocket.Conn.ReadMessage() error: %v", err)
	// 		}
	// 		break
	// 	}
	//
	// 	log.Println("ppcserver: websocket.Conn.ReadMessage() receive:", message)
	//
	// 	// TODO, notify client that new message has arrived.
	// 	// c.hub.broadcast <- message
	// }
}

func (c *client) writeLoop() {

}
