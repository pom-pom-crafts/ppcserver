package connector

const (
	// ClientReadyStateNew is the readyState when the connection has established,
	// and the server is waiting for auth or reconnect message from the peer.
	ClientReadyStateNew clientReadyState = iota
	ClientReadyStateAuthorized
	ClientReadyStateOffline
	ClientReadyStateClosed
)

type (
	clientReadyState int

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
		readyState: ClientReadyStateNew,
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
