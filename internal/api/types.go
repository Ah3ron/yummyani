// Package api provides a client for the YummyAnime API (https://api.yani.tv).
//
// It supports searching for anime, retrieving video entries, and grouping
// videos by dubbing studio. The [AnimeProvider] interface allows consumers
// to swap the real implementation for mocks in tests.
package api

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// SearchResult represents a single anime search result.
type SearchResult struct {
	AnimeID     int             `json:"anime_id"`
	Title       string          `json:"title"`
	Year        json.Number     `json:"year"`
	Type        AnimeType       `json:"type"`
	AnimeStatus AnimeStatusType `json:"anime_status"`
}

// AnimeType describes the anime format (TV, Movie, OVA, etc.).
type AnimeType struct {
	Shortname string `json:"shortname"`
	Name      string `json:"name"`
}

// DisplayName returns the shortname if available, otherwise the full name.
func (t AnimeType) DisplayName() string {
	if t.Shortname != "" {
		return t.Shortname
	}
	return t.Name
}

// AnimeStatusType describes the airing status of an anime.
type AnimeStatusType struct {
	Title string `json:"title"`
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

// FlexInt is an int that can unmarshal from either a JSON string ("42")
// or a JSON number (42). The YummyAnime API inconsistently encodes
// episode numbers as either type.
type FlexInt int

// UnmarshalJSON implements json.Unmarshaler for [FlexInt].
// It tries decoding as a JSON number first, then as a quoted string.
func (fi *FlexInt) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("FlexInt: empty input")
	}

	// Try raw integer first (no surrounding quotes).
	if data[0] != '"' {
		var v int
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("FlexInt: cannot parse %q as int: %w", data, err)
		}
		*fi = FlexInt(v)
		return nil
	}

	// Quoted string — unmarshal then parse.
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("FlexInt: cannot unmarshal string: %w", err)
	}
	v, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return fmt.Errorf("FlexInt: cannot parse %q as int: %w", s, err)
	}
	*fi = FlexInt(v)
	return nil
}

// VideoData holds metadata about the video entry.
type VideoData struct {
	Dubbing string `json:"dubbing"`
	Player  string `json:"player"`
}

// AnimeInfo represents basic anime information.
type AnimeInfo struct {
	Title string `json:"title"`
	Name  string `json:"name"`
}

// DisplayName returns Title if non-empty, otherwise Name.
func (a AnimeInfo) DisplayName() string {
	if a.Title != "" {
		return a.Title
	}
	return a.Name
}

// DubbingGroup groups video entries by dubbing studio name.
type DubbingGroup struct {
	Name     string
	Player   string
	Episodes []VideoEntry
}

// GroupByDubbing groups video entries by dubbing name, filtering only
// entries whose Player field contains "Kodik". Insertion order is preserved.
func GroupByDubbing(videos []VideoEntry) []DubbingGroup {
	groups := make(map[string]*DubbingGroup, len(videos))
	var order []string

	for _, v := range videos {
		if !strings.Contains(v.Data.Player, "Kodik") {
			continue
		}

		key := v.Data.Dubbing
		if _, exists := groups[key]; !exists {
			groups[key] = &DubbingGroup{
				Name:     key,
				Player:   v.Data.Player,
				Episodes: make([]VideoEntry, 0, 8),
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
