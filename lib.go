package kittycad

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	client.Constant = &ConstantService{client: client}
	client.Executor = &ExecutorService{client: client}
	client.File = &FileService{client: client}
	client.Hidden = &HiddenService{client: client}
	client.Meta = &MetaService{client: client}
	client.Ml = &MlService{client: client}
	client.Modeling = &ModelingService{client: client}
	client.Oauth2 = &Oauth2Service{client: client}
	client.Org = &OrgService{client: client}
	client.Payment = &PaymentService{client: client}
	client.ServiceAccount = &ServiceAccountService{client: client}
	client.Shortlink = &ShortlinkService{client: client}
	client.Store = &StoreService{client: client}
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

// HTTPError is an error returned by a failed API call.
type HTTPError struct {
	// URL is the URL that was being accessed when the error occurred.
	// It will always be populated.
	URL *url.URL
	// StatusCode is the HTTP response status code and will always be populated.
	StatusCode int
	// Message is the server response message and is only populated when
	// explicitly referenced by the JSON server response.
	Message string
	// Body is the raw response returned by the server.
	// It is often but not always JSON, depending on how the request fails.
	Body string
	// Header contains the response header fields from the server.
	Header http.Header
}

// Error converts the Error type to a readable string.
func (err HTTPError) Error() string {
	if err.Message != "" {
		return fmt.Sprintf("HTTP %d: %s (%s)", err.StatusCode, err.Message, err.URL)
	}

	return fmt.Sprintf("HTTP %d (%s) BODY -> %v", err.StatusCode, err.URL, err.Body)
}

// checkResponse returns an error (of type *HTTPError) if the response
// status code is not 2xx.
func checkResponse(res *http.Response) error {
	if res.StatusCode >= 200 && res.StatusCode <= 299 {
		return nil
	}

	slurp, err := io.ReadAll(res.Body)
	if err == nil {
		var jerr Error

		// Try to decode the body as an ErrorMessage.
		if err := json.Unmarshal(slurp, &jerr); err == nil {
			return &HTTPError{
				URL:        res.Request.URL,
				StatusCode: res.StatusCode,
				Message:    jerr.Message,
				Body:       string(slurp),
				Header:     res.Header,
			}
		}
	}

	return &HTTPError{
		URL:        res.Request.URL,
		StatusCode: res.StatusCode,
		Body:       string(slurp),
		Header:     res.Header,
		Message:    "",
	}
}
