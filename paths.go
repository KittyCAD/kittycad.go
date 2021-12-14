// Code generated by `generate`. DO NOT EDIT.

package kittycad

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// FileConversionByID: Get a file conversion
//
// Get the status of a file conversion.
//
// Parameters:
//	`id`: The id of the file conversion.
func (c *Client) FileConversionByID(id string) (*FileConversion, error) {
	// Create the url.
	path := "/file/conversion/{{.id}}"
	uri := resolveRelative(c.server, path)
	// Create the request.
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	// Add the parameters to the url.
	if err := expandURL(req.URL, map[string]string{
		"id": string(id),
	}); err != nil {
		return nil, fmt.Errorf("expanding URL with parameters failed: %v", err)
	}
	// Send the request.
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()
	// Check the response.
	if err := checkResponse(resp); err != nil {
		return nil, err
	}
	// Decode the body from the response.
	if resp.Body == nil {
		return nil, errors.New("request returned an empty body in the response")
	}
	var body FileConversion
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("error decoding response body: %v", err)
	}
	// Return the response.
	return &body, nil
}

// FileConvert: Convert CAD file
//
// Convert a CAD file from one format to another. If the file being converted is larger than a certain size it will be performed asynchronously.
//
// Parameters:
//	`sourceFormat`: The format of the file to convert.
//	`outputFormat`: The format the file should be converted to.
func (c *Client) FileConvert(sourceFormat ValidFileType, outputFormat ValidFileType, b io.Reader) (*FileConversion, error) {
	// Create the url.
	path := "/file/conversion/{{.sourceFormat}}/{{.outputFormat}}"
	uri := resolveRelative(c.server, path)
	// Create the request.
	req, err := http.NewRequest("POST", uri, b)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	// Add the parameters to the url.
	if err := expandURL(req.URL, map[string]string{
		"sourceFormat": string(sourceFormat),
		"outputFormat": string(outputFormat),
	}); err != nil {
		return nil, fmt.Errorf("expanding URL with parameters failed: %v", err)
	}
	// Send the request.
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()
	// Check the response.
	if err := checkResponse(resp); err != nil {
		return nil, err
	}
	// Decode the body from the response.
	if resp.Body == nil {
		return nil, errors.New("request returned an empty body in the response")
	}
	var body FileConversion
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("error decoding response body: %v", err)
	}
	// Return the response.
	return &body, nil
}

// Ping: Ping
//
// Simple ping to the server.
func (c *Client) Ping() (*Message, error) {
	// Create the url.
	path := "/ping"
	uri := resolveRelative(c.server, path)
	// Create the request.
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	// Send the request.
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()
	// Check the response.
	if err := checkResponse(resp); err != nil {
		return nil, err
	}
	// Decode the body from the response.
	if resp.Body == nil {
		return nil, errors.New("request returned an empty body in the response")
	}
	var body Message
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("error decoding response body: %v", err)
	}
	// Return the response.
	return &body, nil
}

// MetaDebugInstance: Get instance metadata
//
// Get information about this specific API server instance. This is primarily used for debugging.
func (c *Client) MetaDebugInstance() (*InstanceMetadata, error) {
	// Create the url.
	path := "/_meta/debug/instance"
	uri := resolveRelative(c.server, path)
	// Create the request.
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	// Send the request.
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()
	// Check the response.
	if err := checkResponse(resp); err != nil {
		return nil, err
	}
	// Decode the body from the response.
	if resp.Body == nil {
		return nil, errors.New("request returned an empty body in the response")
	}
	var body InstanceMetadata
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("error decoding response body: %v", err)
	}
	// Return the response.
	return &body, nil
}

// MetaDebugSession: Get auth session
//
// Get information about your API request session. This is primarily used for debugging.
func (c *Client) MetaDebugSession() (*AuthSession, error) {
	// Create the url.
	path := "/_meta/debug/session"
	uri := resolveRelative(c.server, path)
	// Create the request.
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	// Send the request.
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()
	// Check the response.
	if err := checkResponse(resp); err != nil {
		return nil, err
	}
	// Decode the body from the response.
	if resp.Body == nil {
		return nil, errors.New("request returned an empty body in the response")
	}
	var body AuthSession
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("error decoding response body: %v", err)
	}
	// Return the response.
	return &body, nil
}
