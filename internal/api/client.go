package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// AnimeProvider defines the interface for anime data operations.
// Implementations may be the real HTTP client or a test mock.
type AnimeProvider interface {
	Search(ctx context.Context, query string, limit int) ([]SearchResult, error)
	GetVideos(ctx context.Context, animeID int) ([]VideoEntry, error)
	GetAnime(ctx context.Context, animeID int) (*AnimeInfo, error)
}

// Client is an [AnimeProvider] that communicates with the YummyAnime API.
//
// Zero-value is NOT usable; use [NewClient] to create one.
type Client struct {
	BaseURL    string
	AppToken   string
	UserToken  string
	HTTPClient *http.Client
}

// NewClient creates a Client with the given base URL and HTTP timeout.
func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Search performs an anime search query and returns up to limit results.
func (c *Client) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	params := url.Values{}
	params.Set("q", query)
	params.Set("limit", strconv.Itoa(limit))

	data, err := c.get(ctx, "/search", params)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}

	var resp struct {
		Response []SearchResult `json:"response"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parse search results: %w", err)
	}
	return resp.Response, nil
}

// GetVideos returns the list of videos for a given anime ID.
func (c *Client) GetVideos(ctx context.Context, animeID int) ([]VideoEntry, error) {
	endpoint := fmt.Sprintf("/anime/%d/videos", animeID)
	data, err := c.get(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("get videos: %w", err)
	}

	var resp struct {
		Response []VideoEntry `json:"response"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parse videos: %w", err)
	}
	return resp.Response, nil
}

// GetAnime returns basic info about an anime by ID.
func (c *Client) GetAnime(ctx context.Context, animeID int) (*AnimeInfo, error) {
	endpoint := fmt.Sprintf("/anime/%d", animeID)
	data, err := c.get(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("get anime info: %w", err)
	}

	var resp struct {
		Response AnimeInfo `json:"response"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parse anime info: %w", err)
	}
	return &resp.Response, nil
}

// get performs an authenticated GET request and returns the raw response body.
func (c *Client) get(ctx context.Context, endpoint string, params url.Values) ([]byte, error) {
	reqURL := c.BaseURL + endpoint
	if len(params) > 0 {
		reqURL += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request %s: %w", reqURL, err)
	}

	// Set common headers.
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Lang", "ru")
	if c.AppToken != "" {
		req.Header.Set("X-Application", c.AppToken)
	}
	if c.UserToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.UserToken)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request %s: %w", reqURL, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("unexpected status %d from %s: %s",
			resp.StatusCode, reqURL, string(body))
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, fmt.Errorf("read body from %s: %w", reqURL, err)
	}
	return body, nil
}
