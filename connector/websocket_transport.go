package connector

import (
	"github.com/gorilla/websocket"
	"log"
	"net"
	"sync"
	"time"
)

const (
	TransportProtocolTypeWebsocket TransportProtocolType = "websocket"
)

type (
	websocketTransportOptions struct {
		encodingType   EncodingType
		writeTimeout   time.Duration
		maxMessageSize int64
	}

	// websocketTransport is a wrapper struct over websocket connection to fit Transport
	// interface so client will accept it.
	websocketTransport struct {
		conn     *websocket.Conn
		opts     *websocketTransportOptions
		mu       sync.RWMutex
		isClosed bool
	}
)

func newWebsocketTransport(conn *websocket.Conn, opts *websocketTransportOptions) *websocketTransport {
	transport := &websocketTransport{
		conn: conn,
		opts: opts,
	}

	return transport
}

// ProtocolType returns the protocol type of the transport.
func (t *websocketTransport) ProtocolType() TransportProtocolType {
	return TransportProtocolTypeWebsocket
}

// NetConn returns the internal net.Conn of the connection.
func (t *websocketTransport) NetConn() net.Conn {
	return t.conn.UnderlyingConn()
}

// Write data to websocket.Conn.
func (t *websocketTransport) Write(data []byte) error {
	messageType := websocket.TextMessage
	if t.opts.encodingType == EncodingTypeProtobuf {
		messageType = websocket.BinaryMessage
	}

	// SetWriteDeadline should be set per WriteMessage call.
	if t.opts.writeTimeout > 0 {
		_ = t.conn.SetWriteDeadline(time.Now().Add(t.opts.writeTimeout))
	}

	if err := t.conn.WriteMessage(messageType, data); err != nil {
		return err
	}

	return nil
}

func (t *websocketTransport) Close() error {
	t.mu.Lock()

	if t.isClosed {
		t.mu.Unlock()
		return nil
	}

	t.isClosed = true
	t.mu.Unlock()

	return nil
}

func (t *websocketTransport) writeLoop() {

}

func (t *websocketTransport) readLoop() {
	defer func() {
		// TODO, notify Client to remove it from WebSocketServer.
		// c.hub.unregister <- c
		t.conn.Close()
	}()

	if t.opts.maxMessageSize > 0 {
		t.conn.SetReadLimit(t.opts.maxMessageSize)
	}

	// TODO, wait pong
	// c.conn.SetReadDeadline(time.Now().Add(pongWait))
	// c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := t.conn.ReadMessage()

		// Exit readLoop on ReadMessage returns any error.
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("ppcserver: websocket.Conn.ReadMessage() error: %v", err)
			}
			break
		}

		log.Println("ppcserver: websocket.Conn.ReadMessage() receive:", message)

		// TODO, notify client that new message has arrived.
		// c.hub.broadcast <- message
	}
}
