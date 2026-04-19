package kodik

import (
	"encoding/base64"
	"strings"
	"sync"
	"unicode/utf8"
)

const alphabetSize = 26

var (
	shiftCache   uint8
	shiftCachSet bool
	shiftMu      sync.Mutex
)

func getCachedShift() uint8 {
	shiftMu.Lock()
	defer shiftMu.Unlock()
	return shiftCache
}

func setCachedShift(s uint8) {
	shiftMu.Lock()
	defer shiftMu.Unlock()
	shiftCache = s
	shiftCachSet = true
}

func decodeLinks(resp *Response) {
	shift := getCachedShift()
	base360 := decodeLinkSlice(resp.Links.Quality360, shift)

	if shiftCachSet {
		for _, entry := range []struct {
			links   []Link
			quality string
		}{
			{resp.Links.Quality480, "480"},
			{resp.Links.Quality720, "720"},
			{resp.Links.Quality1080, "1080"},
		} {
			deriveOrDecode(entry.links, base360, entry.quality, shift)
		}
		return
	}

	for _, entry := range []struct {
		links   []Link
		quality string
	}{
		{resp.Links.Quality480, "480"},
		{resp.Links.Quality720, "720"},
		{resp.Links.Quality1080, "1080"},
	} {
		deriveOrDecode(entry.links, base360, entry.quality, shift)
	}
	setCachedShift(shift)
}

func decodeLinkSlice(links []Link, shift uint8) string {
	for i := range links {
		links[i].Src = decodeLink(links[i].Src, shift)
	}
	if len(links) > 0 {
		return links[0].Src
	}
	return ""
}

func deriveOrDecode(links []Link, base360, quality string, shift uint8) {
	for i := range links {
		if base360 != "" {
			links[i].Src = strings.Replace(base360, "/360.mp4", "/"+quality+".mp4", 1)
		} else {
			links[i].Src = decodeLink(links[i].Src, shift)
		}
	}
}

func decodeLink(src string, shift uint8) string {
	if shiftCachSet {
		if result, ok := tryDecode(src, shift); ok {
			return ensureScheme(result)
		}
	}
	for s := uint8(0); s <= uint8(alphabetSize); s++ {
		result, ok := tryDecode(src, s)
		if ok {
			setCachedShift(s)
			return ensureScheme(result)
		}
	}
	return src
}

// tryDecode applies a Caesar cipher shift, pads to valid Base64,
// decodes, and validates the result looks like a URL.
func tryDecode(src string, shift uint8) (string, bool) {
	decoded := caesarCipher(src, shift)
	// Pad to valid Base64 length.
	for len(decoded)%4 != 0 {
		decoded += "="
	}
	result, err := base64.StdEncoding.DecodeString(decoded)
	if err != nil {
		return "", false
	}
	if !utf8.Valid(result) {
		return "", false
	}
	s := string(result)
	if !strings.HasPrefix(s, "//") && !strings.HasPrefix(s, "http") {
		return "", false
	}
	return s, true
}

// caesarCipher applies a reverse Caesar shift to text.
func caesarCipher(text string, shift uint8) string {
	var b strings.Builder
	b.Grow(len(text))
	for _, c := range text {
		switch {
		case c >= 'a' && c <= 'z':
			pos := uint8(c-'a') + uint8(alphabetSize) - shift
			b.WriteRune(rune('a' + pos%uint8(alphabetSize)))
		case c >= 'A' && c <= 'Z':
			pos := uint8(c-'A') + uint8(alphabetSize) - shift
			b.WriteRune(rune('A' + pos%uint8(alphabetSize)))
		default:
			b.WriteRune(c)
		}
	}
	return b.String()
}

// ensureScheme prepends "https:" for protocol-relative URLs
// and strips HLS manifest suffix for direct MP4 playback.
func ensureScheme(u string) string {
	if strings.HasPrefix(u, "//") {
		u = "https:" + u
	}
	u = strings.TrimSuffix(u, ":hls:manifest.m3u8")
	return u
}

// EnsureHTTPS prepends "https:" to protocol-relative URLs.
// Exported for use by the TUI layer.
func EnsureHTTPS(u string) string {
	if strings.HasPrefix(u, "//") {
		return "https:" + u
	}
	return u
}
