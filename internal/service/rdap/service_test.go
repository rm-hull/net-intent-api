package rdap

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kofalt/go-memoize"
	"github.com/openrdap/rdap"
)

// mockClient is a test double for clientsrdap.Client
type mockClient struct {
	calls  int
	result *rdap.Domain
	err    error
}

func (m *mockClient) QueryDomain(domain string) (*rdap.Domain, error) {
	m.calls++
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}

func TestGetDomain_CacheHit(t *testing.T) {
	mock := &mockClient{result: &rdap.Domain{Handle: "abc123"}}
	svc := &Service{
		client: mock,
		cache:  memoize.NewMemoizer(5*time.Minute, 1*time.Minute),
		ttl:    5 * time.Minute,
	}

	ctx := context.Background()

	// First call hits upstream
	r1, err := svc.GetDomain(ctx, "example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r1.Handle != "abc123" {
		t.Fatalf("expected Handle abc123, got %s", r1.Handle)
	}
	if mock.calls != 1 {
		t.Fatalf("expected 1 upstream call, got %d", mock.calls)
	}

	// Second call should be served from cache
	r2, err := svc.GetDomain(ctx, "example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r2.Handle != "abc123" {
		t.Fatalf("expected Handle abc123, got %s", r2.Handle)
	}
	if mock.calls != 1 {
		t.Fatalf("expected no additional upstream calls (cache hit), got %d", mock.calls)
	}
}

func TestGetDomain_CacheExpiration(t *testing.T) {
	mock := &mockClient{result: &rdap.Domain{Handle: "abc123"}}
	svc := &Service{
		client: mock,
		cache:  memoize.NewMemoizer(50*time.Millisecond, 50*time.Millisecond),
		ttl:    50 * time.Millisecond,
	}

	ctx := context.Background()

	// First call hits upstream
	_, err := svc.GetDomain(ctx, "example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.calls != 1 {
		t.Fatalf("expected 1 upstream call, got %d", mock.calls)
	}

	// Wait for cache to expire
	time.Sleep(100 * time.Millisecond)

	// Second call should hit upstream again
	_, err = svc.GetDomain(ctx, "example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.calls != 2 {
		t.Fatalf("expected 2 upstream calls (cache miss after expiry), got %d", mock.calls)
	}
}

func TestGetDomain_DifferentDomainsCachedSeparately(t *testing.T) {
	mock := &mockClient{result: &rdap.Domain{Handle: "abc"}}
	svc := &Service{
		client: mock,
		cache:  memoize.NewMemoizer(5*time.Minute, 1*time.Minute),
		ttl:    5 * time.Minute,
	}

	ctx := context.Background()

	_, _ = svc.GetDomain(ctx, "example.com")
	_, _ = svc.GetDomain(ctx, "example.org")

	if mock.calls != 2 {
		t.Fatalf("expected 2 upstream calls for different domains, got %d", mock.calls)
	}
}

func TestGetDomain_UpstreamError(t *testing.T) {
	mock := &mockClient{err: errors.New("upstream failure")}
	svc := &Service{
		client: mock,
		cache:  memoize.NewMemoizer(5*time.Minute, 1*time.Minute),
		ttl:    5 * time.Minute,
	}

	ctx := context.Background()

	_, err := svc.GetDomain(ctx, "example.com")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if mock.calls != 1 {
		t.Fatalf("expected 1 upstream call, got %d", mock.calls)
	}
}

func TestGetDomain_ContextCancellation(t *testing.T) {
	mock := &mockClient{result: &rdap.Domain{Handle: "abc"}}
	svc := &Service{
		client: mock,
		cache:  memoize.NewMemoizer(5*time.Minute, 1*time.Minute),
		ttl:    5 * time.Minute,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := svc.GetDomain(ctx, "example.com")
	if err == nil {
		t.Fatal("expected context cancelled error, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled error, got: %v", err)
	}
}
