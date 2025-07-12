package frrproxy

// Http frr-proxy client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"time"
)

// ClientResponse is a json key value mapping
type ClientResponse map[string]interface{}

// A Client uses the http client to talk
// to the frr-proxy API.
type Client struct {
	api        string
	apiKey     string
	httpClient *http.Client
	afi        string
}

// NewClient creates a new client instance
func NewClient(conf Config) *Client {
	tr := &http.Transport{}

	if conf.TLS {
		tr.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: !conf.TLSVerify,
		}
	}

	// Set default timeout if none is provided
	timeout := conf.Timeout
	if timeout == 0 {
		timeout = 15 * time.Second
	}

	return &Client{
		api:        conf.API,
		apiKey:     conf.APIKey,
		httpClient: &http.Client{Transport: tr, Timeout: timeout},
		afi:        conf.Afi,
	}
}

// PostCommand makes an API request and returns the
// response. The response body will be parsed further
// downstream.
func (c *Client) PostCommand(
	ctx context.Context,
	commandPayload map[string]string,
) (*http.Response, error) {
	// Marshal the payload to JSON
	bodyBytes, err := json.Marshal(commandPayload)
	if err != nil {
		return nil, err
	}

	// Create full URL request
	url := c.api + "/frr"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// RunCommand wraps post command to simplify the command calls further
func (c *Client) RunCommand(ctx context.Context, daemon, command string) (*http.Response, error) {
	return c.PostCommand(ctx, map[string]string{
		"daemon":  daemon,
		"command": command,
	})
}
