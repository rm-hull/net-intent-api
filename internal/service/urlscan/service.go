package urlscan

import (
	"errors"
	"fmt"

	"github.com/rm-hull/net-intent-api/internal/clients/urlscan"
)

type Service struct {
	client *urlscan.Client
}

func NewUrlScanService(apiKey string) *Service {
	return &Service{client: urlscan.NewClient(apiKey)}
}

func (s *Service) GetLatestResult(domain string) (*urlscan.Result, error) {
	resp, err := s.client.Search(domain, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to scan domain %s: %w", domain, err)
	}

	if len(resp.Results) == 0 || resp.Total == 0 {
		return nil, errors.New("no results found")
	}

	return &resp.Results[0], nil
}
