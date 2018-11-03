package ucp

type Logger interface {
	Printf(format string, v ...interface{})
}

// Printf logs debug information to the underlying logger
func (c *Client) Printf(format string, v ...interface{}) {
	if c.logger != nil {
		c.logger.Printf(format, v...)
	}
}
