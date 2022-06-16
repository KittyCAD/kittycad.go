package kittycad

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

//go:generate go run generate/generate.go

// DefaultServerURL is the default server URL for the KittyCad API.
const DefaultServerURL = "https://api.kittycad.io"

// TokenEnvVar is the environment variable that contains the token.
const TokenEnvVar = "KITTYCAD_API_TOKEN"

type service struct {
	client *Client
}

// NewClient creates a new client for the KittyCad API.
// You need to pass in your API token to create the client.
func NewClient(token, userAgent string) (*Client, error) {
	if token == "" {
		return nil, fmt.Errorf("you need to pass in an API token to create the client. Create a token at https://kittycad.io/account")
	}

	client := &Client{
		server: DefaultServerURL,
		token:  token,
	}

	// Ensure the server URL always has a trailing slash.
	if !strings.HasSuffix(client.server, "/") {
		client.server += "/"
	}

	uat := userAgentTransport{
		base:      http.DefaultTransport,
		userAgent: userAgent,
		client:    client,
	}

	client.client = &http.Client{
		Transport: uat,
		// We want a longer timeout since some of the files might take a bit.
		Timeout: 600 * time.Second,
	}

	// Add the services to our client.
	// TODO: this should be part of the generated code.
	client.File = &FileService{client: client}
	client.Meta = &MetaService{client: client}
	client.User = &UserService{client: client}
	client.APICall = &APICallService{client: client}
	client.Payment = &PaymentService{client: client}
	client.APIToken = &APITokenService{client: client}
	client.Session = &SessionService{client: client}
	client.Unit = &UnitService{client: client}

	return client, nil
}

// NewClientFromEnv creates a new client for the KittyCad API, using the token
// stored in the environment variable `KITTYCAD_API_TOKEN`.
func NewClientFromEnv(userAgent string) (*Client, error) {
	token := os.Getenv(TokenEnvVar)
	if token == "" {
		return nil, fmt.Errorf("the environment variable %s must be set with your API token. Create a token at https://kittycad.io/account", TokenEnvVar)
	}

	return NewClient(token, userAgent)
}

// WithBaseURL overrides the baseURL.
func (c *Client) WithBaseURL(baseURL string) error {
	newBaseURL, err := url.Parse(baseURL)
	if err != nil {
		return err
	}

	c.server = newBaseURL.String()

	// Ensure the server URL always has a trailing slash.
	if !strings.HasSuffix(c.server, "/") {
		c.server += "/"
	}

	return nil
}

// WithToken overrides the token used for authentication.
func (c *Client) WithToken(token string) {
	c.token = token
}

type userAgentTransport struct {
	userAgent string
	base      http.RoundTripper
	client    *Client
}

func (t userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.base == nil {
		return nil, errors.New("RoundTrip: no Transport specified")
	}

	newReq := *req
	newReq.Header = make(http.Header)
	for k, vv := range req.Header {
		newReq.Header[k] = vv
	}

	// Add the user agent header.
	newReq.Header["User-Agent"] = []string{t.userAgent}

	// Add the content-type header.
	newReq.Header["Content-Type"] = []string{"application/json"}

	// Add the authorization header.
	newReq.Header["Authorization"] = []string{fmt.Sprintf("Bearer %s", t.client.token)}

	return t.base.RoundTrip(&newReq)
}

// JSONTime is a wrapper around time.Time which marshals to and from empty strings.
type JSONTime struct {
	*time.Time
}

// MarshalJSON implements the json.Marshaler interface.
func (t JSONTime) MarshalJSON() ([]byte, error) {
	return []byte(t.Format("\"" + time.RFC3339 + "\"")), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// The time is expected to be a quoted string in RFC 3339 format.
func (t *JSONTime) UnmarshalJSON(data []byte) (err error) {

	// by convention, unmarshalers implement UnmarshalJSON([]byte("null")) as a no-op.
	if bytes.Equal(data, []byte("null")) {
		return nil
	}

	if bytes.Equal(data, []byte("")) {
		return nil
	}

	if bytes.Equal(data, []byte("\"\"")) {
		return nil
	}

	// Fractional seconds are handled implicitly by Parse.
	tt, err := time.Parse("\""+time.RFC3339+"\"", string(data))
	*t = JSONTime{&tt}
	return
}

// ResponseGetSchema is the response from the GetSchema method.
type ResponseGetSchema interface{}
