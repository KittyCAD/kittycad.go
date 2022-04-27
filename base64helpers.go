package kittycad

import (
	"bytes"
	"encoding/base64"
	"fmt"
)

// GetConversionWithBase64Helper returns the status of a file conversion.
// This function will automatically base64 decode the contents of the result output.
//
// This function is a wrapper around the GetConversion function.
func (c *FileService) GetConversionWithBase64Helper(id string) (*FileConversionWithOutput, []byte, error) {
	resp, err := c.GetConversion(id)
	if err != nil {
		return nil, nil, err
	}

	if resp.Output == "" {
		return resp, nil, nil
	}

	// Decode the base64 encoded body.
	output, err := base64.StdEncoding.DecodeString(resp.Output)
	if err != nil {
		return nil, nil, fmt.Errorf("base64 decoding output from API failed: %v", err)
	}

	return resp, output, nil
}

// CreateConversionWithBase64Helper converts a file.
// This function will automatically base64 encode and decode the contents of the
// src file and output file.
//
// This function is a wrapper around the CreateConversion function.
func (c *FileService) CreateConversionWithBase64Helper(srcFormat FileConversionSourceFormat, outputFormat FileConversionOutputFormat, body []byte) (*FileConversionWithOutput, []byte, error) {
	var b bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &b)
	// Encode the body as base64.
	encoder.Write(body)
	// Must close the encoder when finished to flush any partial blocks.
	// If you comment out the following line, the last partial block "r"
	// won't be encoded.
	encoder.Close()
	resp, err := c.CreateConversion(outputFormat, srcFormat, &b)
	if err != nil {
		return nil, nil, err
	}

	if resp.Output == "" {
		return resp, nil, nil
	}

	// Decode the base64 encoded body.
	output, err := base64.StdEncoding.DecodeString(resp.Output)
	if err != nil {
		return nil, nil, fmt.Errorf("base64 decoding output from API failed: %v", err)
	}

	return resp, output, nil
}
