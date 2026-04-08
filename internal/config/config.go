// Package config provides application-wide configuration for the YummyAnime player.
//
// All values have sensible defaults. Use [Default] to get a pre-populated config,
// or construct [Config] manually to override specific fields.
package config

import "time"

// Config holds all configurable parameters for the application.
type Config struct {
	// API settings.
	APIBaseURL string        // Base URL for the YummyAnime API.
	APITimeout time.Duration // HTTP timeout for API requests.

	// Kodik parser settings.
	KodikTimeout     time.Duration // HTTP timeout for Kodik page/JS fetching.
	KodikMaxAttempts int           // Number of retries for endpoint discovery + POST.
	KodikUserAgent   string        // User-Agent header for Kodik requests.

	// Player settings.
	PlayerCommand string // External player binary name (default: "mpv").

	// UI settings.
	MaxVisibleLines int // Maximum visible items in filter lists.
	SearchLimit     int // Max anime search results.
}

// Default returns a Config with sensible defaults.
func Default() Config {
	return Config{
		APIBaseURL:       "https://api.yani.tv",
		APITimeout:       15 * time.Second,
		KodikTimeout:     30 * time.Second,
		KodikMaxAttempts: 3,
		KodikUserAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
		PlayerCommand:    "mpv",
		MaxVisibleLines:  20,
		SearchLimit:      10,
	}
}
