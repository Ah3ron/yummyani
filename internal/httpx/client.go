// Package httpx provides shared HTTP utilities used across the application.
//
// It wraps net/http with common patterns: body size limits,
// status code checking, and structured error messages.
package httpx

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// MaxBodyBytes is the default limit for HTTP response bodies (2 MiB).
const MaxBodyBytes = 2 << 20

// DoGet performs an HTTP GET request and returns the response body as bytes.
//
// It enforces a body size limit and validates the response status code is 2xx.
func DoGet(ctx context.Context, client *http.Client, rawURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build GET %s: %w", rawURL, err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", rawURL, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("unexpected status %d from %s: %s",
			resp.StatusCode, rawURL, string(body))
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, MaxBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("read body from %s: %w", rawURL, err)
	}
	return body, nil
}

// DoGetString performs an HTTP GET and returns the body as a string.
func DoGetString(ctx context.Context, client *http.Client, rawURL string) (string, error) {
	body, err := DoGet(ctx, client, rawURL)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// DoGetJSON performs an HTTP GET and unmarshals JSON into dest.
func DoGetJSON(ctx context.Context, client *http.Client, rawURL string, dest any) error {
	body, err := DoGet(ctx, client, rawURL)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, dest); err != nil {
		return fmt.Errorf("parse JSON from %s: %w", rawURL, err)
	}
	return nil
}

// DoPostForm performs an HTTP POST with form-encoded data.
// Headers map sets additional request headers (e.g. User-Agent, Origin).
func DoPostForm(ctx context.Context, client *http.Client, rawURL string, form url.Values, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rawURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("build POST %s: %w", rawURL, err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("POST %s: %w", rawURL, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, MaxBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("read body from %s: %w", rawURL, err)
	}
	return body, nil
}
