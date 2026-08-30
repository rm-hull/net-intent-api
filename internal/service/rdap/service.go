package rdap

import (
	"context"
	"time"

	"github.com/kofalt/go-memoize"
	rdap_domain "github.com/openrdap/rdap"
	"github.com/rm-hull/net-intent-api/internal/clients/rdap"
)

type Service struct {
	client rdap.Client
	cache  *memoize.Memoizer
	ttl    time.Duration
}

func NewService(ttl time.Duration) *Service {
	return &Service{
		client: rdap.NewClient(),
		cache:  memoize.NewMemoizer(ttl, ttl),
		ttl:    ttl,
	}
}

func (s *Service) GetDomain(ctx context.Context, domain string) (*rdap_domain.Domain, error) {
	// Check context cancellation first
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if s.ttl > 0 {
		result, err, _ := memoize.Call(s.cache, domain, func() (*rdap_domain.Domain, error) {
			return s.fetch(ctx, domain)
		})

		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		if err != nil {
			return nil, err
		}

		return result, nil
	}

	return s.fetch(ctx, domain)
}

func (s *Service) fetch(ctx context.Context, domain string) (*rdap_domain.Domain, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return s.client.QueryDomain(domain)
}
