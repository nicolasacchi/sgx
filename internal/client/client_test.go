package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
)

func TestGet_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("STATSIG-API-KEY") != "test-key" {
			t.Error("missing API key header")
		}
		if r.Header.Get("STATSIG-API-VERSION") != APIVersion {
			t.Error("missing API version header")
		}
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]any{
			"message": "ok",
			"data":    []string{"a", "b"},
		})
	}))
	defer srv.Close()

	c := New("test-key", srv.URL, false)
	resp, err := c.Get(context.Background(), "/console/v1/test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Message != "ok" {
		t.Errorf("expected message 'ok', got '%s'", resp.Message)
	}
}

func TestGet_404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		json.NewEncoder(w).Encode(map[string]string{"message": "Not found"})
	}))
	defer srv.Close()

	c := New("test-key", srv.URL, false)
	_, err := c.Get(context.Background(), "/console/v1/missing", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.StatusCode != 404 {
		t.Errorf("expected status 404, got %d", apiErr.StatusCode)
	}
	if apiErr.ExitCode() != 4 {
		t.Errorf("expected exit code 4, got %d", apiErr.ExitCode())
	}
}

func TestGet_401(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		w.Write([]byte(`{"message":"Unauthorized"}`))
	}))
	defer srv.Close()

	c := New("bad-key", srv.URL, false)
	_, err := c.Get(context.Background(), "/console/v1/test", nil)
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.ExitCode() != 2 {
		t.Errorf("expected exit code 2 for auth error, got %d", apiErr.ExitCode())
	}
}

func TestGet_429Retry(t *testing.T) {
	var attempts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := attempts.Add(1)
		if n <= 2 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(429)
			return
		}
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]any{
			"message": "ok",
			"data":    "success",
		})
	}))
	defer srv.Close()

	c := New("test-key", srv.URL, false)
	resp, err := c.Get(context.Background(), "/console/v1/test", nil)
	if err != nil {
		t.Fatalf("unexpected error after retries: %v", err)
	}
	if resp.Message != "ok" {
		t.Errorf("expected message 'ok', got '%s'", resp.Message)
	}
	if attempts.Load() != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts.Load())
	}
}

func TestGet_QueryParams(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("status") != "active" {
			t.Errorf("expected status=active, got %s", r.URL.Query().Get("status"))
		}
		if r.URL.Query().Get("limit") != "10" {
			t.Errorf("expected limit=10, got %s", r.URL.Query().Get("limit"))
		}
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]any{"message": "ok", "data": []string{}})
	}))
	defer srv.Close()

	c := New("test-key", srv.URL, false)
	params := url.Values{"status": {"active"}, "limit": {"10"}}
	_, err := c.Get(context.Background(), "/console/v1/experiments", params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetAll_Pagination(t *testing.T) {
	page := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page++
		var nextPage *string
		if page == 1 {
			// Server-side r.URL has no Scheme/Host (the host is in r.Host), so
			// build the absolute next-page URL from r.Host — otherwise nextPage
			// is ":///console/..." and GetAll produces an invalid host:port URL.
			np := "http://" + r.Host + "/console/v1/test?page=2&limit=2"
			nextPage = &np
		}
		resp := map[string]any{
			"message": "ok",
			"data":    []string{"item" + string(rune('0'+page)) + "a", "item" + string(rune('0'+page)) + "b"},
			"pagination": map[string]any{
				"itemsPerPage": 2,
				"pageNumber":   page,
				"totalItems":   4,
				"nextPage":     nextPage,
			},
		}
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := New("test-key", srv.URL, false)
	data, pagination, err := c.GetAll(context.Background(), "/console/v1/test", url.Values{"limit": {"2"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data) != 4 {
		t.Errorf("expected 4 items merged, got %d", len(data))
	}
	if pagination == nil {
		t.Fatal("expected pagination, got nil")
	}
	if pagination.TotalItems != 4 {
		t.Errorf("expected totalItems=4, got %d", pagination.TotalItems)
	}
}

func TestGet_ContextCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response — but context should cancel first
		select {
		case <-r.Context().Done():
			return
		}
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	c := New("test-key", srv.URL, false)
	_, err := c.Get(ctx, "/console/v1/test", nil)
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

func TestGet_MalformedJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("not json"))
	}))
	defer srv.Close()

	c := New("test-key", srv.URL, false)
	_, err := c.Get(context.Background(), "/console/v1/test", nil)
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}
