package kodik

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"regexp"
)

// Compiled regex patterns for endpoint discovery.
var (
	// playerPathRe finds the player JS script path (exact Rust pattern).
	playerPathRe = regexp.MustCompile(
		`<script\s+type="text/javascript"\s+src="/(assets/js/app\.player_single[^"]*)"`,
	)

	// playerPathFallbackRe is a broader fallback for player JS path.
	playerPathFallbackRe = regexp.MustCompile(
		`<script[^>]+src="(/assets/js/app\.player[^"]*)"[^>]*>`,
	)

	// endpointRe finds the base64-encoded API endpoint in player JS:
	// $.ajax(...,url: atob("..."))
	endpointRe = regexp.MustCompile(
		`\$\s*\.\s*ajax\([^>]+,url:\s*atob\(['"]([\w=+/]+)['"]\)`,
	)

	// atobRe is a broad fallback: any atob() call in player JS.
	atobRe = regexp.MustCompile(`atob\(['"]([\w=+/]+)['"]\)`)

	// domainRe extracts domain from URL.
	domainRe = regexp.MustCompile(
		`(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9]`,
	)
)

// extractDomain extracts the domain from a URL using regex (matches Rust impl).
func extractDomain(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	host := u.Hostname()
	if m := domainRe.FindString(host); m != "" {
		return m, nil
	}
	return host, nil
}

// discoverEndpoint finds the player JS URL, fetches it, and extracts
// the base64-encoded API endpoint path.
func discoverEndpoint(getPageFn HTTPGetFunc, ctx context.Context, domain, pageHTML string) (string, error) {
	playerPath := findPlayerPath(pageHTML)
	if playerPath == "" {
		return "", fmt.Errorf("player JS path not found in HTML")
	}

	playerURL := fmt.Sprintf("https://%s/%s", domain, playerPath)
	playerJS, err := getPageFn(ctx, playerURL)
	if err != nil {
		return "", fmt.Errorf("fetch player JS: %w", err)
	}

	return extractEndpointFromJS(playerJS)
}

// findPlayerPath locates the player script path in the page HTML.
func findPlayerPath(html string) string {
	if m := playerPathRe.FindStringSubmatch(html); len(m) >= 2 {
		return m[1]
	}
	if m := playerPathFallbackRe.FindStringSubmatch(html); len(m) >= 2 {
		return m[1]
	}
	return ""
}

// extractEndpointFromJS decodes the base64 API endpoint from player JS.
func extractEndpointFromJS(js string) (string, error) {
	matches := endpointRe.FindStringSubmatch(js)
	if len(matches) < 2 {
		matches = atobRe.FindStringSubmatch(js)
		if len(matches) < 2 {
			return "", fmt.Errorf("API endpoint not found in player JS")
		}
	}

	decoded, err := base64.StdEncoding.DecodeString(matches[1])
	if err != nil {
		decoded, err = base64.RawStdEncoding.DecodeString(matches[1])
		if err != nil {
			return "", fmt.Errorf("decode endpoint: %w", err)
		}
	}
	return string(decoded), nil
}
