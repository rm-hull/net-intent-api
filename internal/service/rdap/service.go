package rdap

import (
	"context"
	"strings"
	"time"

	"github.com/kofalt/go-memoize"
	rdap_domain "github.com/openrdap/rdap"
	"github.com/rm-hull/net-intent-api/internal/clients/rdap"
	"golang.org/x/net/publicsuffix"
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
	// Extract apex domain before lookup/caching
	apex := extractApexDomain(domain)

	// Check context cancellation first
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if s.ttl > 0 {
		result, err, _ := memoize.Call(s.cache, apex, func() (*rdap_domain.Domain, error) {
			return s.fetch(ctx, apex)
		})

		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		if err != nil {
			return nil, err
		}

		return result, nil
	}

	return s.fetch(ctx, apex)
}

func (s *Service) fetch(ctx context.Context, domain string) (*rdap_domain.Domain, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return s.client.QueryDomain(domain)
}

// extractApexDomain extracts the registrable apex domain from the input.
// It handles trailing dots and gracefully falls back to the original
// domain if apex extraction fails.
func extractApexDomain(domain string) string {
	// Trim trailing dots (FQDN form)
	domain = strings.TrimSuffix(domain, ".")

	apex, err := publicsuffix.EffectiveTLDPlusOne(domain)
	if err != nil || apex == "" {
		// Fall back to the original domain if we can't extract apex
		return domain
	}

	return apex
}
