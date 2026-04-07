// Package kodik extracts direct video URLs from Kodik player pages.
// Ported from https://github.com/BatogiX/kodik (Rust).
package kodik

import (
        "encoding/base64"
        "encoding/json"
        "fmt"
        "io"
        "net/http"
        "net/http/cookiejar"
        "net/url"
        "regexp"
        "strings"
        "sync"
        "time"
        "unicode/utf8"
)

const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"

// --- Compiled regex patterns (matching Rust exactly) ---

var (
        // fromURLRe extracts type, id, hash from URL path.
        // Rust: /([^/]+)/(\d+)/([a-z0-9]+)
        fromURLRe = regexp.MustCompile(`/([^/]+)/(\d+)/([a-z0-9]+)`)

        // fromHTMLRe extracts type, hash, id from HTML: vInfo.type = 'value';
        // Rust: \.(?P<field>type|hash|id) = '(?P<value>.*?)';
        fromHTMLRe = regexp.MustCompile(`\.(type|hash|id)\s*=\s*'(.*?)';`)

        // fromHTMLVarRe extracts type/hash/id from var declarations with double quotes:
        // var type = "seria"; var videoId = "1517500";
        fromHTMLVarRe = regexp.MustCompile(`var\s+(type|video_id|videoId|hash|id)\s*=\s*"(.*?)";?`)

        // domainRe extracts domain from URL.
        // Rust: (?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9]
        domainRe = regexp.MustCompile(`(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9]`)

        // playerPathRe finds the player JS script path.
        // Rust: <script\s*type="text/javascript"\s*src="/(assets/js/app\.player_single[^"]*)"
        playerPathRe = regexp.MustCompile(`<script\s+type="text/javascript"\s+src="/(assets/js/app\.player_single[^"]*)"`)

        // endpointRe finds the base64-encoded API endpoint in player JS.
        // Rust: \$\.ajax\([^>]+,url:\s*atob\(["\']([\w=]+)["\']\)
        endpointRe = regexp.MustCompile(`\$\s*\.\s*ajax\([^>]+,url:\s*atob\(['"]([\w=+/]+)['"]\)`)

        // playerPathFallbackRe is a broader fallback for player JS path.
        playerPathFallbackRe = regexp.MustCompile(`<script[^>]+src="(/assets/js/app\.player[^"]*)"[^>]*>`)
)

// --- Data types ---

// Response represents the parsed Kodik player data with video links.
type Response struct {
        Links Links `json:"links"`
}

// Links holds video links grouped by quality.
type Links struct {
        Quality360 []Link `json:"360"`
        Quality480 []Link `json:"480"`
        Quality720 []Link `json:"720"`
}

// Link represents a single video stream URL.
type Link struct {
        Src  string `json:"src"`
        Type string `json:"type"`
}

// videoInfo holds type, hash, id for the POST request.
// Matches Rust VideoInfo struct fields exactly.
type videoInfo struct {
        Type          string `url:"type"`
        Hash          string `url:"hash"`
        ID            string `url:"id"`
        BadUser       string `url:"bad_user"`
        Info          string `url:"info"`
        CDNIsWorking  string `url:"cdn_is_working"`
}

// BestLink returns the highest quality direct link from the response.
func BestLink(resp *Response) string {
        for _, links := range [][]Link{resp.Links.Quality720, resp.Links.Quality480, resp.Links.Quality360} {
                if len(links) > 0 {
                        return links[0].Src
                }
        }
        return ""
}

// --- Main Parse function ---

// Parse extracts a direct video URL from a Kodik player page.
//
// Algorithm (matching Rust kodik-parser exactly):
//  1. Extract domain from URL
//  2. GET the page and extract video info from HTML
//     (URL path values can differ from actual video info for season URLs)
//  3. Discover the API endpoint from player JS (with retry)
//  4. POST video info (6 fields only) to the endpoint
//  5. Decode links using Caesar cipher + base64
func Parse(client *http.Client, rawURL string) (*Response, error) {
        if !strings.HasPrefix(rawURL, "http") {
                rawURL = "https://" + rawURL
        }

        // Step 1: extract domain from URL.
        domain, err := extractDomain(rawURL)
        if err != nil {
                return nil, fmt.Errorf("extract domain: %w", err)
        }

        // Step 2: always fetch the page and extract video info from HTML.
        // The URL path may contain season/series info that differs from the
        // actual embedded video info (type, hash, id) for the specific episode.
        pageHTML, err := httpGet(client, rawURL)
        if err != nil {
                return nil, fmt.Errorf("fetch page: %w", err)
        }
        if len(pageHTML) < 500 {
                return nil, fmt.Errorf("page too small (%d bytes), possible block", len(pageHTML))
        }
        vi, err := videoInfoFromHTML(pageHTML)
        if err != nil {
                return nil, fmt.Errorf("extract video info: %w", err)
        }

        // Step 3-4: discover endpoint and POST (with retry on failure).
        for attempt := 0; attempt < 3; attempt++ {
                // Try cached endpoint first.
                endpoint := globalState.getEndpoint()
                if endpoint != "" {
                        resp, err := postToEndpoint(client, domain, endpoint, &vi)
                        if err == nil {
                                decodeLinks(resp)
                                return resp, nil
                        }
                        // POST failed — clear cached endpoint and try to rediscover.
                        globalState.clearEndpoint()
                        continue
                }

                // No cached endpoint — discover it from player JS.
                endpoint, err = discoverEndpoint(client, domain, pageHTML)
                if err != nil {
                        return nil, fmt.Errorf("discover endpoint (attempt %d): %w", attempt+1, err)
                }
                globalState.setEndpoint(endpoint)
                // Loop back to try POST with the new endpoint.
        }

        return nil, fmt.Errorf("failed after 3 attempts to get video links")
}

// --- Extraction helpers ---

// videoInfoFromURL extracts type, id, hash from URL path.
// Rust: /([^/]+)/(\d+)/([a-z0-9]+) — group 1=type, group 2=id, group 3=hash
func videoInfoFromURL(rawURL string) (videoInfo, bool) {
        matches := fromURLRe.FindStringSubmatch(rawURL)
        if len(matches) < 4 {
                return videoInfo{}, false
        }
        return videoInfo{
                Type:         matches[1],
                ID:           matches[2],
                Hash:         matches[3],
                BadUser:      "True",
                Info:         "{}",
                CDNIsWorking: "True",
        }, true
}

// videoInfoFromHTML extracts type, hash, id from HTML.
// Tries two patterns:
//   1. Rust: \.(type|hash|id) = '(?P<value>.*?)';  (dot-prefix, single quotes)
//   2. Fallback: var type = "..."; var videoId = "...";  (var, double quotes)
func videoInfoFromHTML(html string) (videoInfo, error) {
        vi := extractFromDotNotation(html)
        if vi.Type == "" || vi.Hash == "" || vi.ID == "" {
                vi = mergeWithVarNotation(vi, html)
        }
        if vi.Type == "" || vi.Hash == "" || vi.ID == "" {
                return vi, fmt.Errorf("incomplete video info: type=%q hash=%q id=%q", vi.Type, vi.Hash, vi.ID)
        }
        vi.BadUser = "True"
        vi.Info = "{}"
        vi.CDNIsWorking = "True"
        return vi, nil
}

// extractFromDotNotation matches vInfo.type = 'seria'; patterns.
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

// mergeWithVarNotation fills missing fields from var type = "..."; var videoId = "..."; patterns.
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

// extractDomain extracts the domain from a URL.
// Rust uses the same regex pattern.
func extractDomain(rawURL string) (string, error) {
        u, err := url.Parse(rawURL)
        if err != nil {
                return "", err
        }
        host := u.Hostname()
        matches := domainRe.FindString(host)
        if matches == "" {
                return host, nil
        }
        return matches, nil
}

// discoverEndpoint finds the player JS URL, fetches it, and extracts
// the base64-encoded API endpoint path.
func discoverEndpoint(client *http.Client, domain, pageHTML string) (string, error) {
        // Try the exact Rust pattern first (player_single).
        playerPath := ""
        matches := playerPathRe.FindStringSubmatch(pageHTML)
        if len(matches) >= 2 {
                playerPath = matches[1]
        } else {
                // Fallback: broader pattern.
                matches = playerPathFallbackRe.FindStringSubmatch(pageHTML)
                if len(matches) >= 2 {
                        playerPath = matches[1]
                }
        }
        if playerPath == "" {
                return "", fmt.Errorf("player JS path not found in HTML")
        }

        playerURL := fmt.Sprintf("https://%s/%s", domain, playerPath)
        playerJS, err := httpGet(client, playerURL)
        if err != nil {
                return "", fmt.Errorf("fetch player JS: %w", err)
        }

        // Try the exact Rust pattern first ($.ajax with atob).
        matches = endpointRe.FindStringSubmatch(playerJS)
        if len(matches) < 2 {
                // Fallback: look for any atob() call.
                matches = regexp.MustCompile(`atob\(['"]([\w=+/]+)['"]\)`).FindStringSubmatch(playerJS)
                if len(matches) < 2 {
                        return "", fmt.Errorf("API endpoint not found in player JS")
                }
        }

        decoded, err := base64.StdEncoding.DecodeString(matches[1])
        if err != nil {
                // Try with RawStdEncoding (no padding).
                decoded, err = base64.RawStdEncoding.DecodeString(matches[1])
                if err != nil {
                        return "", fmt.Errorf("decode endpoint: %w", err)
                }
        }
        return string(decoded), nil
}

// --- HTTP helpers ---

func httpGet(client *http.Client, rawURL string) (string, error) {
        req, err := http.NewRequest(http.MethodGet, rawURL, nil)
        if err != nil {
                return "", err
        }
        req.Header.Set("User-Agent", userAgent)

        resp, err := client.Do(req)
        if err != nil {
                return "", fmt.Errorf("get %s: %w", rawURL, err)
        }
        defer resp.Body.Close()

        body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
        if err != nil {
                return "", err
        }
        return string(body), nil
}

// postToEndpoint sends video info (6 fields only) to the Kodik API endpoint.
// This matches the Rust implementation exactly.
func postToEndpoint(client *http.Client, domain, endpoint string, vi *videoInfo) (*Response, error) {
        apiURL := fmt.Sprintf("https://%s%s", domain, endpoint)

        form := url.Values{}
        form.Set("type", vi.Type)
        form.Set("hash", vi.Hash)
        form.Set("id", vi.ID)
        form.Set("bad_user", vi.BadUser)       // "True" (capital T)
        form.Set("info", vi.Info)              // "{}"
        form.Set("cdn_is_working", vi.CDNIsWorking) // "True" (capital T)

        req, err := http.NewRequest(http.MethodPost, apiURL, strings.NewReader(form.Encode()))
        if err != nil {
                return nil, err
        }

        // Headers matching Rust exactly.
        req.Header.Set("User-Agent", userAgent)
        req.Header.Set("Origin", fmt.Sprintf("https://%s", domain))
        req.Header.Set("Referer", fmt.Sprintf("https://%s", domain))
        req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
        req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
        req.Header.Set("X-Requested-With", "XMLHttpRequest")

        resp, err := client.Do(req)
        if err != nil {
                return nil, fmt.Errorf("post %s: %w", apiURL, err)
        }
        defer resp.Body.Close()

        body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
        if err != nil {
                return nil, err
        }

        if len(body) == 0 {
                return nil, fmt.Errorf("empty response (status %d)", resp.StatusCode)
        }
        if body[0] != '{' && body[0] != '[' {
                return nil, fmt.Errorf("unexpected response (status %d): %.200s", resp.StatusCode, string(body))
        }

        var kodikResp Response
        if err := unmarshalLinks(body, &kodikResp); err != nil {
                return nil, fmt.Errorf("parse response: %w", err)
        }
        return &kodikResp, nil
}

// unmarshalLinks parses the JSON response with integer quality keys.
func unmarshalLinks(data []byte, resp *Response) error {
        var raw map[string]json.RawMessage
        if err := json.Unmarshal(data, &raw); err != nil {
                return err
        }
        linksRaw, ok := raw["links"]
        if !ok {
                return fmt.Errorf("no 'links' field in response")
        }
        var linksMap map[string][]Link
        if err := json.Unmarshal(linksRaw, &linksMap); err != nil {
                return err
        }
        resp.Links.Quality360 = linksMap["360"]
        resp.Links.Quality480 = linksMap["480"]
        resp.Links.Quality720 = linksMap["720"]
        return nil
}

// --- Decoder (matching Rust decoder.rs exactly) ---

const maxShift = 26

// decodeLinks decodes all links using Caesar cipher + base64.
func decodeLinks(resp *Response) {
        // Decode 360p links first.
        for i := range resp.Links.Quality360 {
                resp.Links.Quality360[i].Src = decodeLink(resp.Links.Quality360[i].Src)
        }
        base360 := ""
        if len(resp.Links.Quality360) > 0 {
                base360 = resp.Links.Quality360[0].Src
        }
        // 480p: derive from 360p if available, otherwise decode independently.
        for i := range resp.Links.Quality480 {
                if base360 != "" {
                        resp.Links.Quality480[i].Src = strings.Replace(base360, "/360.mp4", "/480.mp4", 1)
                } else {
                        resp.Links.Quality480[i].Src = decodeLink(resp.Links.Quality480[i].Src)
                }
        }
        // 720p: derive from 360p if available, otherwise decode independently.
        for i := range resp.Links.Quality720 {
                if base360 != "" {
                        resp.Links.Quality720[i].Src = strings.Replace(base360, "/360.mp4", "/720.mp4", 1)
                } else {
                        resp.Links.Quality720[i].Src = decodeLink(resp.Links.Quality720[i].Src)
                }
        }
}

func decodeLink(src string) string {
        // Try cached shift first.
        shift := globalState.getShift()
        if result, ok := tryDecode(src, shift); ok {
                return ensureScheme(result)
        }
        // Brute-force all shifts.
        for s := uint8(0); s <= uint8(maxShift); s++ {
                if result, ok := tryDecode(src, s); ok {
                        globalState.setShift(s)
                        return ensureScheme(result)
                }
        }
        return src
}

func tryDecode(src string, shift uint8) (string, bool) {
        decoded := caesarCipher(src, shift)
        for len(decoded)%4 != 0 {
                decoded += "="
        }
        result, err := base64.StdEncoding.DecodeString(decoded)
        if err != nil {
                return "", false
        }
        // Validate UTF-8 (matches Rust's String::from_utf8).
        if !utf8.Valid(result) {
                return "", false
        }
        s := string(result)
        // The decoded result must look like a URL.
        if !strings.HasPrefix(s, "//") && !strings.HasPrefix(s, "http") {
                return "", false
        }
        return s, true
}

func caesarCipher(text string, shift uint8) string {
        var b strings.Builder
        b.Grow(len(text))
        for _, c := range text {
                if c >= 'a' && c <= 'z' {
                        pos := uint8(c - 'a')
                        newPos := (pos + uint8(maxShift) - shift) % uint8(maxShift)
                        b.WriteRune(rune('a' + newPos))
                } else if c >= 'A' && c <= 'Z' {
                        pos := uint8(c - 'A')
                        newPos := (pos + uint8(maxShift) - shift) % uint8(maxShift)
                        b.WriteRune(rune('A' + newPos))
                } else {
                        b.WriteRune(c)
                }
        }
        return b.String()
}

func ensureScheme(u string) string {
        if strings.HasPrefix(u, "//") {
                u = "https:" + u
        }
        // Strip HLS manifest suffix — use direct MP4 for mpv.
        // The :hls:manifest.m3u8 suffix causes mpv to interpret the URL as
        // an HLS playlist, showing each .ts segment as a separate "episode".
        u = strings.TrimSuffix(u, ":hls:manifest.m3u8")
        return u
}

// --- State (matching Rust KodikState with endpoint caching) ---

type state struct {
        mu         sync.RWMutex
        shift      uint8
        cachedPath string
}

func (s *state) getShift() uint8 {
        s.mu.RLock()
        defer s.mu.RUnlock()
        return s.shift
}

func (s *state) setShift(v uint8) {
        s.mu.Lock()
        defer s.mu.Unlock()
        s.shift = v
}

func (s *state) getEndpoint() string {
        s.mu.RLock()
        defer s.mu.RUnlock()
        return s.cachedPath
}

func (s *state) setEndpoint(v string) {
        s.mu.Lock()
        defer s.mu.Unlock()
        s.cachedPath = v
}

func (s *state) clearEndpoint() {
        s.mu.Lock()
        defer s.mu.Unlock()
        s.cachedPath = ""
}

var globalState = &state{}

// --- Client factory ---

// NewHTTPClient creates an HTTP client with a cookie jar.
func NewHTTPClient() *http.Client {
        jar, _ := cookiejar.New(nil)
        return &http.Client{
                Jar:     jar,
                Timeout: 30 * time.Second,
                CheckRedirect: func(req *http.Request, via []*http.Request) error {
                        if len(via) >= 10 {
                                return fmt.Errorf("too many redirects")
                        }
                        return nil
                },
        }
}
