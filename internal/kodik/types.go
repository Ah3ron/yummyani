// Package kodik extracts direct video URLs from Kodik player pages.
//
// Algorithm:
//  1. Fetch the player page HTML
//  2. Extract video info (type, hash, id) from the HTML
//  3. Discover the API endpoint from the player JS
//  4. POST video info to the endpoint
//  5. Decode the resulting links using Caesar cipher + Base64
//
// For testability, use [Parser] with [NewParser]. The package-level
// convenience functions have been removed to eliminate global state.
package kodik

import (
	"encoding/json"
	"fmt"
)

// Response represents the parsed Kodik player data with video links.
type Response struct {
	Links Links `json:"links"`
}

// Links holds video links grouped by quality.
type Links struct {
	Quality360  []Link `json:"360"`
	Quality480  []Link `json:"480"`
	Quality720  []Link `json:"720"`
	Quality1080 []Link `json:"1080"`
}

// Link represents a single video stream URL.
type Link struct {
	Src  string `json:"src"`
	Type string `json:"type"`
}

// Quality represents a video quality level with its available links.
type Quality struct {
	Label string // e.g. "1080p"
	Links []Link
}

// qualityEntry maps a quality label to its links for iteration.
type qualityEntry struct {
	label string
	links []Link
}

// allQualities returns qualities in descending order (1080p → 360p).
func allQualities(l *Links) []qualityEntry {
	return []qualityEntry{
		{"1080p", l.Quality1080},
		{"720p", l.Quality720},
		{"480p", l.Quality480},
		{"360p", l.Quality360},
	}
}

// AvailableQualities returns qualities that have at least one link,
// ordered from highest to lowest (1080p → 360p).
func AvailableQualities(resp *Response) []Quality {
	var result []Quality
	for _, q := range allQualities(&resp.Links) {
		if len(q.links) > 0 {
			result = append(result, Quality{Label: q.label, Links: q.links})
		}
	}
	return result
}

// BestLink returns the highest quality direct link from the response,
// or an empty string if no links are available.
func BestLink(resp *Response) string {
	for _, q := range allQualities(&resp.Links) {
		if len(q.links) > 0 {
			return q.links[0].Src
		}
	}
	return ""
}

// unmarshalLinks parses the JSON response with integer quality keys.
func unmarshalLinks(data []byte, resp *Response) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	linksRaw, ok := raw["links"]
	if !ok {
		return fmt.Errorf("missing \"links\" field")
	}

	var linksMap map[string][]Link
	if err := json.Unmarshal(linksRaw, &linksMap); err != nil {
		return err
	}
	resp.Links.Quality360 = linksMap["360"]
	resp.Links.Quality480 = linksMap["480"]
	resp.Links.Quality720 = linksMap["720"]
	resp.Links.Quality1080 = linksMap["1080"]
	return nil
}
