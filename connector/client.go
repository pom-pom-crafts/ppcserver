package connector

type (
	// Client represents a client connection to any server of the connector.
	Client struct {
		transport Transport
	}
)

// NewClient creates a new Client.
func NewClient(transport Transport) *Client {
	client := &Client{
		transport: transport,
	}

	return client
}

func (c *Client) Write(data []byte) error {
	if err := c.transport.Write(data); err != nil {
		return err
	}
	return nil
}

func (c *Client) Close() {

}
