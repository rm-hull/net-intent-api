package urlscan

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kofalt/go-memoize"
	api "github.com/rm-hull/net-intent-api/internal/clients/urlscan"
)

// mockClient is a test double for api.Client
type mockClient struct {
	calls    int
	result   *api.Result
	err      error
	response *api.Response
}

func (m *mockClient) Search(domain string, size int) (*api.Response, error) {
	m.calls++
	if m.err != nil {
		return nil, m.err
	}
	if m.response != nil {
		return m.response, nil
	}
	return &api.Response{Results: []api.Result{*m.result}, Total: 1}, nil
}

func makeResult(id string) *api.Result {
	return &api.Result{ID: id}
}

func TestGetLatestResult_CacheHit(t *testing.T) {
	mock := &mockClient{result: makeResult("abc123")}
	svc := &Service{
		client: mock,
		cache:  memoize.NewMemoizer(5*time.Minute, 1*time.Minute),
		ttl:    5 * time.Minute,
	}

	ctx := context.Background()

	// First call hits upstream
	r1, err := svc.GetLatestResult(ctx, "example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r1.ID != "abc123" {
		t.Fatalf("expected ID abc123, got %s", r1.ID)
	}
	if mock.calls != 1 {
		t.Fatalf("expected 1 upstream call, got %d", mock.calls)
	}

	// Second call should be served from cache
	r2, err := svc.GetLatestResult(ctx, "example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r2.ID != "abc123" {
		t.Fatalf("expected ID abc123, got %s", r2.ID)
	}
	if mock.calls != 1 {
		t.Fatalf("expected no additional upstream calls (cache hit), got %d", mock.calls)
	}
}

func TestGetLatestResult_CacheExpiration(t *testing.T) {
	mock := &mockClient{result: makeResult("abc123")}
	// Use a very short TTL so the cache expires quickly
	svc := &Service{
		client: mock,
		cache:  memoize.NewMemoizer(50*time.Millisecond, 50*time.Millisecond),
		ttl:    50 * time.Millisecond,
	}

	ctx := context.Background()

	// First call hits upstream
	_, err := svc.GetLatestResult(ctx, "example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.calls != 1 {
		t.Fatalf("expected 1 upstream call, got %d", mock.calls)
	}

	// Wait for cache to expire
	time.Sleep(100 * time.Millisecond)

	// Second call should hit upstream again
	_, err = svc.GetLatestResult(ctx, "example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.calls != 2 {
		t.Fatalf("expected 2 upstream calls (cache miss after expiry), got %d", mock.calls)
	}
}

func TestGetLatestResult_DifferentDomainsCachedSeparately(t *testing.T) {
	mock := &mockClient{result: makeResult("abc")}
	svc := &Service{
		client: mock,
		cache:  memoize.NewMemoizer(5*time.Minute, 1*time.Minute),
		ttl:    5 * time.Minute,
	}

	ctx := context.Background()

	_, _ = svc.GetLatestResult(ctx, "example.com")
	_, _ = svc.GetLatestResult(ctx, "example.org")

	if mock.calls != 2 {
		t.Fatalf("expected 2 upstream calls for different domains, got %d", mock.calls)
	}
}

func TestGetLatestResult_UpstreamError(t *testing.T) {
	mock := &mockClient{err: errors.New("upstream failure")}
	svc := &Service{
		client: mock,
		cache:  memoize.NewMemoizer(5*time.Minute, 1*time.Minute),
		ttl:    5 * time.Minute,
	}

	ctx := context.Background()

	_, err := svc.GetLatestResult(ctx, "example.com")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if mock.calls != 1 {
		t.Fatalf("expected 1 upstream call, got %d", mock.calls)
	}
}

func TestGetLatestResult_NoResults(t *testing.T) {
	mock := &mockClient{response: &api.Response{Results: nil, Total: 0}}
	svc := &Service{
		client: mock,
		cache:  memoize.NewMemoizer(5*time.Minute, 1*time.Minute),
		ttl:    5 * time.Minute,
	}

	ctx := context.Background()

	_, err := svc.GetLatestResult(ctx, "example.com")
	if err == nil {
		t.Fatal("expected error for no results, got nil")
	}
}

func TestGetLatestResult_ContextCancellation(t *testing.T) {
	mock := &mockClient{result: makeResult("abc123")}
	svc := &Service{
		client: mock,
		cache:  memoize.NewMemoizer(5*time.Minute, 1*time.Minute),
		ttl:    5 * time.Minute,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := svc.GetLatestResult(ctx, "example.com")
	if err == nil {
		t.Fatal("expected context cancelled error, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled error, got: %v", err)
	}
}
