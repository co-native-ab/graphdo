// Package graph provides a lightweight HTTP client for the Microsoft Graph API.
package graph

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

// TokenSource provides bearer tokens for Graph API requests.
// Implementations should handle caching and refresh internally.
type TokenSource interface {
	Token(ctx context.Context) (string, error)
}

// StaticToken returns a TokenSource that always provides the given token.
func StaticToken(token string) TokenSource {
	return staticToken(token)
}

type staticToken string

func (s staticToken) Token(_ context.Context) (string, error) {
	return string(s), nil
}

// Client is a lightweight HTTP client for the Microsoft Graph API.
type Client struct {
	BaseURL     string
	HTTPClient  *http.Client
	TokenSource TokenSource
}

// APIError represents an error response from the Microsoft Graph API.
type APIError struct {
	StatusCode int
	Code       string `json:"code"`
	Message    string `json:"message"`
}

// Error returns a human-readable description of the Graph API error.
func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %s (HTTP %d)", e.Code, e.Message, e.StatusCode)
}

// NewClient creates a new Graph API client with the given base URL and token source.
func NewClient(baseURL string, ts TokenSource) *Client {
	return &Client{
		BaseURL:     strings.TrimRight(baseURL, "/"),
		HTTPClient:  &http.Client{},
		TokenSource: ts,
	}
}

func (c *Client) do(ctx context.Context, method, path string, body any) (*http.Response, error) {
	url := c.BaseURL + path

	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("graph %s %s: %w", method, path, err)
		}
		reqBody = bytes.NewReader(data)
	}

	slog.Debug("graph request", "method", method, "path", path)

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("graph %s %s: %w", method, path, err)
	}

	token, err := c.TokenSource.Token(ctx)
	if err != nil {
		return nil, fmt.Errorf("graph %s %s: acquiring token: %w", method, path, err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("graph %s %s: %w", method, path, err)
	}

	slog.Debug("graph response", "method", method, "path", path, "status", resp.StatusCode)

	if resp.StatusCode >= 400 {
		defer func() { _ = resp.Body.Close() }()
		graphErr := parseGraphError(resp)
		return nil, fmt.Errorf("graph %s %s: %w", method, path, graphErr)
	}

	return resp, nil
}

func parseGraphError(resp *http.Response) error {
	var envelope struct {
		Error APIError `json:"error"`
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return &APIError{
			StatusCode: resp.StatusCode,
			Code:       "UnknownError",
			Message:    fmt.Sprintf("failed to read error response: %v", err),
		}
	}

	if err := json.Unmarshal(data, &envelope); err != nil {
		return &APIError{
			StatusCode: resp.StatusCode,
			Code:       "UnknownError",
			Message:    string(data),
		}
	}

	envelope.Error.StatusCode = resp.StatusCode
	return &envelope.Error
}
