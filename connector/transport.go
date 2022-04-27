package connector

import "net"

type (
	// TransportProtocolType describes the protocol type name of the connection transport between server and client,
	// currently only supports "websocket".
	TransportProtocolType string

	// Transport abstracts a connection transport between server and client.
	Transport interface {
		// ProtocolType should return the protocol type of the transport.
		ProtocolType() TransportProtocolType
		// NetConn should return the internal net.Conn of the connection.
		NetConn() net.Conn
		Read() ([]byte, error)
		// Write should write single data into a connection.
		Write([]byte) error
		// Close must close transport.
		Close() error
	}
)
