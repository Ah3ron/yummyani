package kodik

import (
	"context"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"sync"
	"time"
)

type LinkExtractor interface {
	Parse(ctx context.Context, rawURL string) (*Response, error)
	Prefetch(ctx context.Context, urls []string)
}

type Parser struct {
	client      *http.Client
	userAgent   string
	maxAttempts int
	httpGet     HTTPGetFunc
	cachedPath  string
	prefetch    sync.Map
}

// ParserOption configures a [Parser].
type ParserOption func(*Parser)

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(c *http.Client) ParserOption {
	return func(p *Parser) { p.client = c }
}

// WithUserAgent sets the User-Agent header.
func WithUserAgent(ua string) ParserOption {
	return func(p *Parser) { p.userAgent = ua }
}

// WithMaxAttempts sets the number of retries.
func WithMaxAttempts(n int) ParserOption {
	return func(p *Parser) { p.maxAttempts = n }
}

// WithHTTPGet sets a custom GET function (for testing).
func WithHTTPGet(fn HTTPGetFunc) ParserOption {
	return func(p *Parser) { p.httpGet = fn }
}

// NewParser creates a Parser with sensible defaults.
// Default: cookie jar, 30s timeout, Chrome-like User-Agent, 3 attempts.
func NewParser(opts ...ParserOption) *Parser {
	jar, _ := cookiejar.New(nil)
	p := &Parser{
		client: &http.Client{
			Jar:     jar,
			Timeout: 30 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return fmt.Errorf("too many redirects")
				}
				return nil
			},
		},
		userAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
		maxAttempts: 3,
		httpGet:     nil, // will use defaultHTTPGet
	}
	for _, opt := range opts {
		opt(p)
	}
	if p.httpGet == nil {
		ua := p.userAgent
		cli := p.client
		p.httpGet = func(ctx context.Context, rawURL string) (string, error) {
			return defaultHTTPGet(ctx, cli, ua, rawURL)
		}
	}
	return p
}

func (p *Parser) Parse(ctx context.Context, rawURL string) (*Response, error) {
	if !strings.HasPrefix(rawURL, "http") {
		rawURL = "https://" + rawURL
	}

	if cached, ok := p.prefetch.Load(rawURL); ok {
		return cached.(*Response), nil
	}

	domain, err := extractDomain(rawURL)
	if err != nil {
		return nil, fmt.Errorf("extract domain: %w", err)
	}

	pageHTML, err := p.httpGet(ctx, rawURL)
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

	var resp *Response
	for attempt := 0; attempt < p.maxAttempts; attempt++ {
		if p.cachedPath != "" {
			body, err := postVideoInfo(ctx, p.client, p.userAgent, domain, p.cachedPath, vi)
			if err == nil {
				resp, err = parseKodikResponse(body)
				if err == nil {
					decodeLinks(resp)
					return resp, nil
				}
			}
			p.cachedPath = ""
			continue
		}

		endpoint, err := discoverEndpoint(p.httpGet, ctx, domain, pageHTML)
		if err != nil {
			return nil, fmt.Errorf("discover endpoint (attempt %d): %w", attempt+1, err)
		}
		p.cachedPath = endpoint
	}

	return nil, fmt.Errorf("failed after %d attempts to get video links", p.maxAttempts)
}

func (p *Parser) Prefetch(ctx context.Context, urls []string) {
	var wg sync.WaitGroup
	for _, url := range urls {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			if resp, err := p.Parse(ctx, u); err == nil {
				p.prefetch.Store(u, resp)
			}
		}(url)
	}
	wg.Wait()
}

func parseKodikResponse(body []byte) (*Response, error) {
	var resp Response
	if err := unmarshalLinks(body, &resp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return &resp, nil
}
