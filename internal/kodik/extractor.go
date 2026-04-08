package kodik

import (
	"fmt"
	"regexp"
)

// Compiled regex patterns for HTML parsing (matching Rust kodik-parser).
var (
	// fromHTMLRe extracts type, hash, id from dot-notation: .type = 'value';
	fromHTMLRe = regexp.MustCompile(`\.(type|hash|id)\s*=\s*'(.*?)';`)

	// fromHTMLVarRe extracts from var declarations: var type = "value";
	fromHTMLVarRe = regexp.MustCompile(`var\s+(type|video_id|videoId|hash|id)\s*=\s*"(.*?)";?`)
)

// videoInfo holds the fields needed for the POST request to the Kodik API.
type videoInfo struct {
	Type         string
	Hash         string
	ID           string
	BadUser      string
	Info         string
	CDNIsWorking string
}

// toFormValues converts videoInfo to a map for POST form encoding.
func (vi videoInfo) toFormValues() map[string]string {
	return map[string]string{
		"type":           vi.Type,
		"hash":           vi.Hash,
		"id":             vi.ID,
		"bad_user":       vi.BadUser,
		"info":           vi.Info,
		"cdn_is_working": vi.CDNIsWorking,
	}
}

// videoInfoFromHTML extracts type, hash, id from HTML using two patterns:
//  1. Dot-notation: .(type|hash|id) = 'value';
//  2. Var declaration: var type = "value";
func videoInfoFromHTML(html string) (videoInfo, error) {
	vi := extractFromDotNotation(html)
	if vi.Type == "" || vi.Hash == "" || vi.ID == "" {
		vi = mergeWithVarNotation(vi, html)
	}
	if vi.Type == "" || vi.Hash == "" || vi.ID == "" {
		return vi, fmt.Errorf("incomplete video info: type=%q hash=%q id=%q",
			vi.Type, vi.Hash, vi.ID)
	}
	vi.BadUser = "True"
	vi.Info = "{}"
	vi.CDNIsWorking = "True"
	return vi, nil
}

// extractFromDotNotation matches .type = 'seria'; patterns.
func extractFromDotNotation(html string) videoInfo {
	var vi videoInfo
	matches := fromHTMLRe.FindAllStringSubmatch(html, -1)
	for _, m := range matches {
		switch m[1] {
		case "type":
			vi.Type = m[2]
		case "hash":
			vi.Hash = m[2]
		case "id":
			vi.ID = m[2]
		}
	}
	return vi
}

// mergeWithVarNotation fills missing fields from var declarations.
func mergeWithVarNotation(vi videoInfo, html string) videoInfo {
	matches := fromHTMLVarRe.FindAllStringSubmatch(html, -1)
	for _, m := range matches {
		switch m[1] {
		case "type":
			if vi.Type == "" {
				vi.Type = m[2]
			}
		case "hash":
			if vi.Hash == "" {
				vi.Hash = m[2]
			}
		case "video_id", "videoId":
			if vi.ID == "" {
				vi.ID = m[2]
			}
		case "id":
			if vi.ID == "" {
				vi.ID = m[2]
			}
		}
	}
	return vi
}
