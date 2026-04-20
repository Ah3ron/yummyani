// Package fuzzy provides a simple fuzzy string matching algorithm.
//
// Match returns true if every rune in the query appears in the text in order.
// This is a subsequence match, not a substring match.
//
// It handles multi-byte UTF-8 correctly (Cyrillic, CJK, etc.).
package fuzzy

import "strings"

// Match reports whether query is a subsequence of text (case‑insensitive).
// Every rune in query must appear in text in the same order, though not
// necessarily contiguously.
func Match(query, text string) bool {
	qr := []rune(strings.ToLower(query))
	qi := 0
	for _, r := range strings.ToLower(text) {
		if qi < len(qr) && r == qr[qi] {
			qi++
		}
	}
	if qi == len(qr) {
		if len([]rune(query)) == len([]rune(text)) && query != text {
			return false
		}
		return true
	}
	return false
}
