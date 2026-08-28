package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// APIError is returned by Client.Do for any non-2xx response.
type APIError struct {
	Method     string
	Path       string
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("ohdear API %s %s: %d: %s", e.Method, e.Path, e.StatusCode, e.Body)
}

// apiStatus returns the HTTP status carried by an APIError, or 0.
func apiStatus(err error) int {
	var ae *APIError
	if errors.As(err, &ae) {
		return ae.StatusCode
	}
	return 0
}

const defaultBaseURL = "https://ohdear.app/api"

// Client is a thin Oh Dear API client. Every request carries
// `Authorization: Bearer <token>`.
type Client struct {
	http    *http.Client
	baseURL string
	token   string
}

func NewClient(token string) *Client {
	return &Client{http: http.DefaultClient, baseURL: defaultBaseURL, token: token}
}

// Do sends an API request. body is JSON-encoded when non-nil; out is
// JSON-decoded from the response when non-nil.
func (c *Client) Do(ctx context.Context, method, path string, body, out any) error {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(resp.Body)
		return &APIError{Method: method, Path: path, StatusCode: resp.StatusCode, Body: string(bytes.TrimSpace(msg))}
	}

	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

// List GETs path and returns its items, tolerating both a bare JSON array and a
// `{"data": [...]}` envelope (the Oh Dear API uses both across endpoints).
func List[T any](ctx context.Context, c *Client, path string) ([]T, error) {
	var raw json.RawMessage
	if err := c.Do(ctx, http.MethodGet, path, nil, &raw); err != nil {
		return nil, err
	}
	var list []T
	if json.Unmarshal(raw, &list) == nil {
		return list, nil
	}
	var env struct {
		Data []T `json:"data"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, err
	}
	return env.Data, nil
}
