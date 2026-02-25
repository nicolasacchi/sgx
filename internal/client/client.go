package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand/v2"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultBaseURL = "https://statsigapi.net"
	APIVersion     = "20240601"
	MaxRetries     = 3
	MaxPages       = 20
	Timeout        = 30 * time.Second
)

type Client struct {
	http       *http.Client
	apiKey     string
	baseURL    string
	apiVersion string
	verbose    bool
}

type APIResponse struct {
	Message    string          `json:"message"`
	Data       json.RawMessage `json:"data"`
	Pagination *Pagination     `json:"pagination,omitempty"`
}

type Pagination struct {
	ItemsPerPage int     `json:"itemsPerPage"`
	PageNumber   int     `json:"pageNumber"`
	TotalItems   int     `json:"totalItems"`
	NextPage     *string `json:"nextPage"`
	PreviousPage *string `json:"previousPage"`
}

type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%d: %s", e.StatusCode, e.Message)
}

func (e *APIError) ExitCode() int {
	switch {
	case e.StatusCode == 401 || e.StatusCode == 403:
		return 2
	case e.StatusCode == 404:
		return 4
	default:
		return 1
	}
}

func New(apiKey, baseURL string, verbose bool) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return &Client{
		http:       &http.Client{Timeout: Timeout},
		apiKey:     apiKey,
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiVersion: APIVersion,
		verbose:    verbose,
	}
}

func (c *Client) Get(ctx context.Context, path string, params url.Values) (*APIResponse, error) {
	u := c.baseURL + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}
	return c.doGet(ctx, u)
}

func (c *Client) GetAbsolute(ctx context.Context, rawURL string) (*APIResponse, error) {
	if !strings.HasPrefix(rawURL, "http") {
		rawURL = c.baseURL + rawURL
	}
	return c.doGet(ctx, rawURL)
}

func (c *Client) GetAll(ctx context.Context, path string, params url.Values) ([]json.RawMessage, *Pagination, error) {
	var allData []json.RawMessage
	var lastPagination *Pagination

	firstURL := c.baseURL + path
	if len(params) > 0 {
		firstURL += "?" + params.Encode()
	}

	currentURL := firstURL
	for page := 0; page < MaxPages; page++ {
		resp, err := c.doGet(ctx, currentURL)
		if err != nil {
			return allData, lastPagination, err
		}

		var pageData []json.RawMessage
		if err := json.Unmarshal(resp.Data, &pageData); err != nil {
			// Data is not an array — return as single item
			allData = append(allData, resp.Data)
			return allData, resp.Pagination, nil
		}
		allData = append(allData, pageData...)
		lastPagination = resp.Pagination

		if resp.Pagination == nil || resp.Pagination.NextPage == nil || *resp.Pagination.NextPage == "" {
			break
		}
		currentURL = *resp.Pagination.NextPage
		if !strings.HasPrefix(currentURL, "http") {
			currentURL = c.baseURL + currentURL
		}
	}

	return allData, lastPagination, nil
}

func (c *Client) doGet(ctx context.Context, rawURL string) (*APIResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("STATSIG-API-KEY", c.apiKey)
	req.Header.Set("STATSIG-API-VERSION", c.apiVersion)
	req.Header.Set("Content-Type", "application/json")

	if c.verbose {
		fmt.Fprintf(os.Stderr, "> GET %s\n", rawURL)
	}

	start := time.Now()
	resp, err := c.doWithRetry(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	elapsed := time.Since(start)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if c.verbose {
		fmt.Fprintf(os.Stderr, "< %d %s (%s, %s)\n", resp.StatusCode, http.StatusText(resp.StatusCode), elapsed.Round(time.Millisecond), humanBytes(len(body)))
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := extractErrorMessage(body, resp.StatusCode)
		return nil, &APIError{StatusCode: resp.StatusCode, Message: msg}
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("unexpected response format from Statsig API: %w", err)
	}
	return &apiResp, nil
}

func (c *Client) doWithRetry(ctx context.Context, req *http.Request) (*http.Response, error) {
	var lastErr error
	for attempt := 0; attempt <= MaxRetries; attempt++ {
		if attempt > 0 {
			// Clone the request for retry since body may have been consumed
			retryReq := req.Clone(ctx)
			req = retryReq
		}

		resp, err := c.http.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			if attempt == MaxRetries {
				return nil, lastErr
			}
			delay := retryDelay(attempt, "")
			if c.verbose {
				fmt.Fprintf(os.Stderr, "! request error, retrying in %s (attempt %d/%d)\n", delay, attempt+1, MaxRetries)
			}
			select {
			case <-time.After(delay):
				continue
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		if resp.StatusCode != http.StatusTooManyRequests {
			return resp, nil
		}

		resp.Body.Close()
		if attempt == MaxRetries {
			return nil, &APIError{StatusCode: 429, Message: "rate limited after max retries"}
		}

		delay := retryDelay(attempt, resp.Header.Get("Retry-After"))
		if c.verbose {
			fmt.Fprintf(os.Stderr, "! 429 Too Many Requests, retrying in %s (attempt %d/%d)\n", delay, attempt+1, MaxRetries)
		}
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return nil, lastErr
}

func retryDelay(attempt int, retryAfter string) time.Duration {
	if retryAfter != "" {
		if seconds, err := strconv.Atoi(retryAfter); err == nil {
			return time.Duration(seconds) * time.Second
		}
	}
	base := time.Duration(math.Pow(2, float64(attempt))) * time.Second
	jitter := time.Duration(rand.IntN(500)) * time.Millisecond
	return base + jitter
}

func extractErrorMessage(body []byte, statusCode int) string {
	var parsed struct {
		Message string `json:"message"`
		Error   string `json:"error"`
	}
	if json.Unmarshal(body, &parsed) == nil {
		if parsed.Message != "" {
			return parsed.Message
		}
		if parsed.Error != "" {
			return parsed.Error
		}
	}

	switch statusCode {
	case 401:
		return "authentication failed — check STATSIG_API_KEY env var or --api-key flag"
	case 403:
		return "forbidden — your API key may not have read access to this resource"
	case 404:
		return "not found"
	default:
		return http.StatusText(statusCode)
	}
}

func humanBytes(b int) string {
	if b < 1024 {
		return fmt.Sprintf("%dB", b)
	}
	kb := float64(b) / 1024
	if kb < 1024 {
		return fmt.Sprintf("%.1fKB", kb)
	}
	return fmt.Sprintf("%.1fMB", kb/1024)
}
