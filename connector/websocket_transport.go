package connector

import (
	"github.com/gorilla/websocket"
	"net"
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
	// interface so Client will accept it.
	websocketTransport struct {
		conn *websocket.Conn
		opts *websocketTransportOptions
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

func (t *websocketTransport) Read() ([]byte, error) {
	_, message, err := t.conn.ReadMessage()
	return message, err
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

// Close closes the underlying network connection.
// It can be called concurrently, and it's OK to call Close more than once.
func (t *websocketTransport) Close() error {
	return t.conn.Close()
}
