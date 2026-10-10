// Package gorest is a small, dependency-free client for the GoRest v2 API
// (https://gorest.co.in).
package gorest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// DefaultBaseURL is the public GoRest v2 endpoint.
const DefaultBaseURL = "https://gorest.co.in/public/v2"

// User is a GoRest user resource.
type User struct {
	ID     int    `json:"id,omitempty"`
	Name   string `json:"name,omitempty"`
	Email  string `json:"email,omitempty"`
	Gender string `json:"gender,omitempty"` // "male" | "female"
	Status string `json:"status,omitempty"` // "active" | "inactive"
}

// FieldError is a single validation error returned by the API (HTTP 422).
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// APIError is returned for any non-2xx response.
type APIError struct {
	StatusCode int
	Message    string
	Fields     []FieldError
}

func (e *APIError) Error() string {
	if len(e.Fields) > 0 {
		parts := make([]string, len(e.Fields))
		for i, f := range e.Fields {
			parts[i] = f.Field + " " + f.Message
		}
		return fmt.Sprintf("gorest: %d: %s", e.StatusCode, strings.Join(parts, "; "))
	}
	msg := e.Message
	if msg == "" {
		msg = http.StatusText(e.StatusCode)
	}
	return fmt.Sprintf("gorest: %d: %s", e.StatusCode, msg)
}

// Client talks to the GoRest API. The zero value is not usable; use New.
type Client struct {
	BaseURL    string
	Token      string // required for create, update and delete
	HTTPClient *http.Client
}

// New returns a Client for the public API. token may be empty for read-only use.
func New(token string) *Client {
	return &Client{
		BaseURL:    DefaultBaseURL,
		Token:      token,
		HTTPClient: &http.Client{Timeout: 15 * time.Second},
	}
}

// ListUsers returns one page of users (page starts at 1).
func (c *Client) ListUsers(ctx context.Context, page, perPage int) ([]User, error) {
	q := url.Values{}
	if page > 0 {
		q.Set("page", strconv.Itoa(page))
	}
	if perPage > 0 {
		q.Set("per_page", strconv.Itoa(perPage))
	}
	var users []User
	err := c.do(ctx, http.MethodGet, "/users?"+q.Encode(), nil, &users)
	return users, err
}

// GetUser fetches a single user by ID.
func (c *Client) GetUser(ctx context.Context, id int) (*User, error) {
	var u User
	if err := c.do(ctx, http.MethodGet, userPath(id), nil, &u); err != nil {
		return nil, err
	}
	return &u, nil
}

// CreateUser creates a user and returns it with its assigned ID.
func (c *Client) CreateUser(ctx context.Context, u User) (*User, error) {
	var out User
	if err := c.do(ctx, http.MethodPost, "/users", u, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateUser patches the non-empty fields of u on the user with the given ID.
func (c *Client) UpdateUser(ctx context.Context, id int, u User) (*User, error) {
	u.ID = 0
	var out User
	if err := c.do(ctx, http.MethodPatch, userPath(id), u, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteUser deletes the user with the given ID.
func (c *Client) DeleteUser(ctx context.Context, id int) error {
	return c.do(ctx, http.MethodDelete, userPath(id), nil, nil)
}

func userPath(id int) string { return "/users/" + strconv.Itoa(id) }

func (c *Client) do(ctx context.Context, method, path string, in, out any) error {
	var body io.Reader
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return fmt.Errorf("gorest: encode request: %w", err)
		}
		body = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, body)
	if err != nil {
		return fmt.Errorf("gorest: build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("gorest: %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("gorest: read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return parseError(resp.StatusCode, data)
	}
	if out == nil || len(data) == 0 {
		return nil
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("gorest: decode response: %w", err)
	}
	return nil
}

func parseError(status int, data []byte) error {
	apiErr := &APIError{StatusCode: status}
	if json.Unmarshal(data, &apiErr.Fields) == nil && len(apiErr.Fields) > 0 {
		return apiErr
	}
	var msg struct {
		Message string `json:"message"`
	}
	if json.Unmarshal(data, &msg) == nil {
		apiErr.Message = msg.Message
	}
	return apiErr
}
