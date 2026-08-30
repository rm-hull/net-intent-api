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

func (s *Service) GetDomain(ctx context.Context, domain string) (*DomainResponse, error) {
	// Extract apex domain before lookup/caching
	apex := extractApexDomain(domain)

	// Check context cancellation first
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if s.ttl > 0 {
		result, err, _ := memoize.Call(s.cache, apex, func() (*DomainResponse, error) {
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

func (s *Service) fetch(ctx context.Context, domain string) (*DomainResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	result, err := s.client.QueryDomain(domain)
	if err != nil {
		return nil, err
	}

	return toDomainResponse(result), nil
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

type DomainResponse struct {
	ObjectClassName string           `json:"objectClassName"`
	Handle          string           `json:"handle,omitempty"`
	LDHName         string           `json:"ldhName,omitempty"`
	UnicodeName     string           `json:"unicodeName,omitempty"`
	Status          []string         `json:"status,omitempty"`
	Events          []eventResponse  `json:"events,omitempty"`
	Entities        []entityResponse `json:"entities,omitempty"`
	Nameservers     []string         `json:"nameservers,omitempty"`
	SecureDNS       *secureDNSResp   `json:"secureDNS,omitempty"`
	PublicIDs       []publicIDResp   `json:"publicIds,omitempty"`
	Notices         []noticeResponse `json:"notices,omitempty"`

	// Derived fields — not in the RDAP spec, but useful for your own scoring.
	RegistrationAgeDays *int `json:"registrationAgeDays,omitempty"`
}

type eventResponse struct {
	Action string `json:"eventAction"`
	Date   string `json:"eventDate"`
	Actor  string `json:"eventActor,omitempty"`
}

type entityResponse struct {
	Handle string   `json:"handle,omitempty"`
	Roles  []string `json:"roles,omitempty"`
}

type secureDNSResp struct {
	DelegationSigned bool `json:"delegationSigned"`
}

type publicIDResp struct {
	Type       string `json:"type"`
	Identifier string `json:"identifier"`
}

type noticeResponse struct {
	Title       string   `json:"title,omitempty"`
	Description []string `json:"description,omitempty"`
}

func toDomainResponse(d *rdap_domain.Domain) *DomainResponse {
	out := DomainResponse{
		ObjectClassName: "domain",
		Handle:          d.Handle,
		LDHName:         d.LDHName,
		UnicodeName:     d.UnicodeName,
		Status:          d.Status,
	}

	for _, e := range d.Events {
		out.Events = append(out.Events, eventResponse{
			Action: e.Action,
			Date:   e.Date,
			Actor:  e.Actor,
		})

		if e.Action == "registration" {
			if t, err := time.Parse(time.RFC3339, e.Date); err == nil {
				days := int(time.Since(t).Hours() / 24)
				out.RegistrationAgeDays = &days
			}
		}
	}

	for _, ent := range d.Entities {
		out.Entities = append(out.Entities, entityResponse{
			Handle: ent.Handle,
			Roles:  ent.Roles,
		})
	}

	for _, ns := range d.Nameservers {
		out.Nameservers = append(out.Nameservers, ns.LDHName)
	}

	if d.SecureDNS != nil && d.SecureDNS.DelegationSigned != nil {
		out.SecureDNS = &secureDNSResp{
			DelegationSigned: *d.SecureDNS.DelegationSigned,
		}
	}

	for _, pid := range d.PublicIDs {
		out.PublicIDs = append(out.PublicIDs, publicIDResp{
			Type:       pid.Type,
			Identifier: pid.Identifier,
		})
	}

	for _, n := range d.Notices {
		out.Notices = append(out.Notices, noticeResponse{
			Title:       n.Title,
			Description: n.Description,
		})
	}

	return &out
}
