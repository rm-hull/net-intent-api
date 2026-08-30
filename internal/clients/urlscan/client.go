package urlscan

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Client interface {
	Search(domain string, size int) (*Response, error)
}

type clientImpl struct {
	APIKey     string
	HTTPClient *http.Client
}

// Ensure clientImpl satisfies Clietn
var _ Client = (*clientImpl)(nil)

func NewClient(apiKey string) Client {
	return &clientImpl{
		APIKey: apiKey,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *clientImpl) Search(domain string, size int) (*Response, error) {
	domain = strings.TrimSuffix(domain, ".")

	req, err := http.NewRequest("GET", "https://urlscan.io/api/v1/search", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("q", "page.domain:"+domain)
	q.Add("size", fmt.Sprintf("%d", size))
	req.URL.RawQuery = q.Encode()
	req.Header.Set("API-Key", c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

type Response struct {
	Results []Result `json:"results"`
	Total   int      `json:"total"`
	Took    int      `json:"took"`
	HasMore bool     `json:"has_more"`
}

type Result struct {
	ID         string        `json:"_id"`
	Score      *float64      `json:"_score"`
	Sort       []interface{} `json:"sort"`
	Result     string        `json:"result"`
	Screenshot string        `json:"screenshot"`
	Submitter  struct{}      `json:"submitter"`
	Canonical  Canonical     `json:"canonical"`
	Task       Task          `json:"task"`
	Stats      Stats         `json:"stats"`
	Files      []File        `json:"files,omitempty"`
	Page       Page          `json:"page"`
}

type File struct {
	Filename        string `json:"filename"`
	SHA256          string `json:"sha256"`
	Filesize        int64  `json:"filesize"`
	State           string `json:"state"`
	MimeType        string `json:"mimeType"`
	MimeDescription string `json:"mimeDescription"`
	URL             string `json:"url"`
}

type Canonical struct {
	Task struct {
		URL string `json:"url"`
	} `json:"task"`
	Page struct {
		URL string `json:"url"`
	} `json:"page"`
}

type Task struct {
	Visibility string    `json:"visibility"`
	Method     string    `json:"method"`
	Domain     string    `json:"domain"`
	ApexDomain string    `json:"apexDomain"`
	Time       time.Time `json:"time"`
	UUID       string    `json:"uuid"`
	URL        string    `json:"url"`
}

type Stats struct {
	UniqIPs           int `json:"uniqIPs"`
	UniqCountries     int `json:"uniqCountries"`
	DataLength        int `json:"dataLength"`
	EncodedDataLength int `json:"encodedDataLength"`
	Requests          int `json:"requests"`
}

type Page struct {
	Country           string    `json:"country,omitempty"`
	Server            string    `json:"server"`
	Redirected        string    `json:"redirected,omitempty"`
	IP                string    `json:"ip"`
	ApexDomainAgeDays int       `json:"apexDomainAgeDays"`
	MimeType          string    `json:"mimeType"`
	URL               string    `json:"url"`
	TLSValidDays      int       `json:"tlsValidDays"`
	TLSAgeDays        int       `json:"tlsAgeDays"`
	PTR               string    `json:"ptr,omitempty"`
	DomainAgeDays     int       `json:"domainAgeDays"`
	TLSValidFrom      time.Time `json:"tlsValidFrom"`
	Domain            string    `json:"domain"`
	UmbrellaRank      int       `json:"umbrellaRank"`
	ApexDomain        string    `json:"apexDomain"`
	ASNName           string    `json:"asnname"`
	ASN               string    `json:"asn"`
	TLSIssuer         string    `json:"tlsIssuer"`
	Status            string    `json:"status"`
}
