package fuzzy_test

import (
	"testing"

	"yummyani/pkg/fuzzy"
)

func TestMatch(t *testing.T) {
	tests := []struct {
		name  string
		query string
		text  string
		want  bool
	}{
		{"empty query", "", "anything", true},
		{"exact match", "hello", "hello", true},
		{"subsequence", "hlo", "hello", true},
		{"not subsequence", "ohl", "hello", false},
		{"cyrillic", "нтр", "интернет", true},
		{"cyrillic not found", "абв", "гдеж", false},
		{"single char", "n", "naruto", true},
		{"case sensitive lower", "Naruto", "naruto", false},
		{"case sensitive upper", "naruto", "Naruto", false},
		{"mixed case match", "nT", "naRuto", true},
		{"longer query", "abcdefghij", "abc", false},
		{"numbers", "12", "Episode 123", true},
		{"spaces in query", "a b", "a b", true},
		{"partial", "anime", "anime search query", true},
		{"reverse order", "olleh", "hello", false},
		{"special chars", "a.b", "a.b.c", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fuzzy.Match(tt.query, tt.text)
			if got != tt.want {
				t.Errorf("Match(%q, %q) = %v, want %v", tt.query, tt.text, got, tt.want)
			}
		})
	}
}
