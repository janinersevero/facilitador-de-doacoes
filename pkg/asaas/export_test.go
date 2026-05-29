package asaas

// SetBaseURL exposes the internal baseURL for use in unit tests.
func SetBaseURL(c *Client, url string) {
	c.baseURL = url
}
