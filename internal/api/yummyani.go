// Package api provides a client for the YummyAnime API.
package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const defaultBaseURL = "https://api.yani.tv"

// Client is a YummyAnime API client.
type Client struct {
	BaseURL    string
	AppToken   string
	UserToken  string
	HTTPClient *http.Client
}

// NewClient creates a new API client with default settings.
func NewClient() *Client {
	return &Client{
		BaseURL: defaultBaseURL,
		HTTPClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// SearchResult represents a single anime search result.
type SearchResult struct {
	AnimeID     int             `json:"anime_id"`
	Title       string          `json:"title"`
	Year        json.Number     `json:"year"`
	Type        AnimeType       `json:"type"`
	AnimeStatus AnimeStatusType `json:"anime_status"`
}

// AnimeType describes the anime format.
type AnimeType struct {
	Shortname string `json:"shortname"`
	Name      string `json:"name"`
}

// AnimeStatusType describes the airing status.
type AnimeStatusType struct {
	Title string `json:"title"`
}

// Search performs an anime search query.
func (c *Client) Search(query string, limit int) ([]SearchResult, error) {
	params := url.Values{}
	params.Set("q", query)
	params.Set("limit", fmt.Sprintf("%d", limit))

	data, err := c.get("/search", params)
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

// VideoEntry represents a single video (episode) entry.
type VideoEntry struct {
	Number    FlexInt   `json:"number"`
	Duration  int       `json:"duration"`
	Views     int       `json:"views"`
	IframeURL string    `json:"iframe_url"`
	VideoID   int       `json:"video_id"`
	Data      VideoData `json:"data"`
}

// FlexInt is an int that can unmarshal from a JSON string or number.
type FlexInt int

func (fi *FlexInt) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), "\"")
	v, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	*fi = FlexInt(v)
	return nil
}

// VideoData holds metadata about the video entry.
type VideoData struct {
	Dubbing string `json:"dubbing"`
	Player  string `json:"player"`
}

// GetVideos returns the list of videos for a given anime ID.
func (c *Client) GetVideos(animeID int) ([]VideoEntry, error) {
	endpoint := fmt.Sprintf("/anime/%d/videos", animeID)
	data, err := c.get(endpoint, nil)
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

// AnimeInfo represents basic anime information.
type AnimeInfo struct {
	Title string `json:"title"`
	Name  string `json:"name"`
}

// GetAnime returns basic info about an anime by ID.
func (c *Client) GetAnime(animeID int) (*AnimeInfo, error) {
	endpoint := fmt.Sprintf("/anime/%d", animeID)
	data, err := c.get(endpoint, nil)
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

// get performs an authenticated GET request.
func (c *Client) get(endpoint string, params url.Values) ([]byte, error) {
	reqURL := c.BaseURL + endpoint
	if len(params) > 0 {
		reqURL += "?" + params.Encode()
	}

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept", "image/avif,image/webp")
	req.Header.Set("Lang", "ru")

	if c.AppToken != "" {
		req.Header.Set("X-Application", c.AppToken)
	}
	if c.UserToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.UserToken)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, err
	}

	return body, nil
}

// DubbingGroup groups video entries by dubbing name.
type DubbingGroup struct {
	Name     string
	Player   string
	Episodes []VideoEntry
}

// GroupByDubbing groups video entries by dubbing and player, filtering only Kodik.
func GroupByDubbing(videos []VideoEntry) []DubbingGroup {
	groups := make(map[string]*DubbingGroup)
	var order []string

	for _, v := range videos {
		// Only support Kodik player.
		if !strings.Contains(v.Data.Player, "Kodik") {
			continue
		}

		key := v.Data.Dubbing
		if _, exists := groups[key]; !exists {
			groups[key] = &DubbingGroup{
				Name:     key,
				Player:   v.Data.Player,
				Episodes: []VideoEntry{},
			}
			order = append(order, key)
		}
		groups[key].Episodes = append(groups[key].Episodes, v)
	}

	result := make([]DubbingGroup, 0, len(order))
	for _, k := range order {
		result = append(result, *groups[k])
	}
	return result
}
