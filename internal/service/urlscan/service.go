package urlscan

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kofalt/go-memoize"
	"github.com/rm-hull/net-intent-api/internal/clients/urlscan"
)

type Service struct {
	client urlscan.Client
	cache  *memoize.Memoizer
	ttl    time.Duration
}

func NewUrlScanService(apiKey string, ttl time.Duration) *Service {
	return &Service{
		client: urlscan.NewClient(apiKey),
		cache:  memoize.NewMemoizer(ttl, ttl),
		ttl:    ttl,
	}
}

func (s *Service) GetLatestResult(ctx context.Context, domain string) (*urlscan.Result, error) {
	cacheKey := domain

	if s.ttl > 0 {
		result, err, cached := memoize.Call(s.cache, cacheKey, func() (*urlscan.Result, error) {
			return s.fetch(ctx, domain)
		})

		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		if err != nil {
			return nil, err
		}

		// Return cached result or fresh result (cached == true means it was found in cache)
		_ = cached
		return result, nil
	}

	return s.fetch(ctx, domain)
}

func (s *Service) fetch(ctx context.Context, domain string) (*urlscan.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	resp, err := s.client.Search(domain, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to scan domain %s: %w", domain, err)
	}

	if len(resp.Results) == 0 || resp.Total == 0 {
		return nil, errors.New("no results found")
	}

	return &resp.Results[0], nil
}
