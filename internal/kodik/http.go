package kodik

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// HTTPGetFunc is a function type for making HTTP GET requests.
// Allows injecting mock implementations in tests.
type HTTPGetFunc func(ctx context.Context, rawURL string) (string, error)

// defaultHTTPGet performs a GET request with User-Agent header.
func defaultHTTPGet(ctx context.Context, client *http.Client, userAgent string, rawURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return "", fmt.Errorf("build GET %s: %w", rawURL, err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("GET %s: %w", rawURL, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return "", fmt.Errorf("read body %s: %w", rawURL, err)
	}
	return string(body), nil
}

// postVideoInfo sends video info to the Kodik API endpoint.
func postVideoInfo(ctx context.Context, client *http.Client, userAgent, domain, endpoint string, vi videoInfo) ([]byte, error) {
	apiURL := fmt.Sprintf("https://%s%s", domain, endpoint)

	form := url.Values{}
	for k, v := range vi.toFormValues() {
		form.Set(k, v)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("build POST %s: %w", apiURL, err)
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Origin", "https://"+domain)
	req.Header.Set("Referer", "https://"+domain)
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("POST %s: %w", apiURL, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, fmt.Errorf("read body %s: %w", apiURL, err)
	}

	if len(body) == 0 {
		return nil, fmt.Errorf("empty response (status %d)", resp.StatusCode)
	}
	if body[0] != '{' && body[0] != '[' {
		return nil, fmt.Errorf("unexpected response (status %d): %.200s", resp.StatusCode, string(body))
	}
	return body, nil
}
