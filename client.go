package kittycad

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// Client which conforms to the OpenAPI3 specification for this service.
type Client struct {
	// The endpoint of the server conforming to this interface, with scheme,
	// https://api.kittycad.io for example.
	Server string

	// Client is the *http.Client for performing requests.
	Client *http.Client
}

// NewClient creates a new client for the KittyCad API.
// You need to pass in your API token to create the client.
func NewClient(token string) (*Client, error) {
	if token == "" {
		return nil, fmt.Errorf("you need to pass in an API token to create the client. Create a token at https://kittycad.io/account")
	}

	client := &Client{
		Server: DefaultServerURL,
		Client: &http.Client{},
	}

	// Ensure the server URL always has a trailing slash.
	if !strings.HasSuffix(client.Server, "/") {
		client.Server += "/"
	}

	return client, nil
}

// NewClientFromEnv creates a new client for the KittyCad API, using the token
// stored in the environment variable `KITTYCAD_API_TOKEN`.
func NewClientFromEnv() (*Client, error) {
	token := os.Getenv(TokenEnvVar)
	if token == "" {
		return nil, fmt.Errorf("the environment variable %s must be set with your API token. Create a token at https://kittycad.io/account", TokenEnvVar)
	}

	return NewClient(token)
}

// WithHTTPClient allows overriding the default http.Client, which is
// automatically created using http.Client. This is useful for tests.
func (c *Client) WithHTTPClient(client *http.Client) {
	c.Client = client
}

// WithBaseURL overrides the baseURL.
func (c *Client) WithBaseURL(baseURL string) error {
	newBaseURL, err := url.Parse(baseURL)
	if err != nil {
		return err
	}

	c.Server = newBaseURL.String()

	// Ensure the server URL always has a trailing slash.
	if !strings.HasSuffix(c.Server, "/") {
		c.Server += "/"
	}

	return nil
}
