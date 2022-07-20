package kittycad

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

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
	client.APICall = &APICallService{client: client}
	client.APIToken = &APITokenService{client: client}
	client.App = &AppService{client: client}
	client.Beta = &BetaService{client: client}
	client.File = &FileService{client: client}
	client.Hidden = &HiddenService{client: client}
	client.Meta = &MetaService{client: client}
	client.Oauth2 = &Oauth2Service{client: client}
	client.Payment = &PaymentService{client: client}
	client.Session = &SessionService{client: client}
	client.Unit = &UnitService{client: client}
	client.User = &UserService{client: client}
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

// ResponseGetSchema is the response from the GetSchema method.
type ResponseGetSchema interface{}
