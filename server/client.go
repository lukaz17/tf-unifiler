// Copyright (C) 2025  T-Force I/O
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/tforceaio/tf-unifiler/task"
)

const clientTimeout = 5 * time.Second

// Client talks to a running TFunifiler server over HTTP.
type Client struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

// Return a new Client targeting the server at baseURL (e.g. "http://127.0.0.1:8080"), authenticating with apiKey unless it is empty.
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		http:    &http.Client{Timeout: clientTimeout},
	}
}

// Enqueue args as a task on the server and return its snapshot.
func (c *Client) Enqueue(ctx context.Context, args []string) (task.Task, error) {
	body, err := json.Marshal(enqueueRequest{Args: args})
	if err != nil {
		return task.Task{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/tasks", bytes.NewReader(body))
	if err != nil {
		return task.Task{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.do(req)
	if err != nil {
		return task.Task{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return task.Task{}, fmt.Errorf("server rejected task (%s): %s", resp.Status, serverError(resp.Body))
	}

	var out task.Task
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return task.Task{}, fmt.Errorf("decoding enqueue response: %w", err)
	}
	return out, nil
}

// Fetch the server's status document.
func (c *Client) Status(ctx context.Context) (StatusResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/status", nil)
	if err != nil {
		return StatusResponse{}, err
	}

	resp, err := c.do(req)
	if err != nil {
		return StatusResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return StatusResponse{}, fmt.Errorf("status endpoint returned %s", resp.Status)
	}

	var out StatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return StatusResponse{}, fmt.Errorf("decoding status response: %w", err)
	}
	return out, nil
}

// Send req with the API key attached, reporting a rejected key as ErrUnauthorized.
func (c *Client) do(req *http.Request) (*http.Response, error) {
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		return nil, ErrUnauthorized
	}
	return resp, nil
}

// Extract the server's error message from a JSON body of the form {"error": "..."}, falling back to a generic note when it cannot be parsed.
func serverError(body io.Reader) string {
	var e struct {
		Error string `json:"error"`
	}
	if err := json.NewDecoder(body).Decode(&e); err != nil || e.Error == "" {
		return "unknown error"
	}
	return e.Error
}
