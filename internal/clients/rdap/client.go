package rdap

import (
	"fmt"
	"strings"

	"github.com/openrdap/rdap"
)

type Client interface {
	QueryDomain(domain string) (*rdap.Domain, error)
}

type clientImpl struct {
	client *rdap.Client
}

// Ensure clientImpl satisfies Client
var _ Client = (*clientImpl)(nil)

func NewClient() Client {
	return &clientImpl{
		client: &rdap.Client{},
	}
}

func (c *clientImpl) QueryDomain(domain string) (*rdap.Domain, error) {
	// Trim trailing dot if present
	if before, found := strings.CutSuffix(domain, "."); found {
		domain = before
	}

	return c.client.QueryDomain(domain)
}

func ClassifyError(err error) (notFound bool) {
	if err == nil {
		return false
	}

	clientErr, ok := err.(*rdap.ClientError)
	if !ok {
		return false
	}

	switch clientErr.Type {
	case rdap.NoWorkingServers, rdap.ObjectDoesNotExist:
		return true
	}

	return false
}

func ErrorMessage(domain string, err error) string {
	return fmt.Sprintf("failed to lookup: %s: %s", domain, err.Error())
}
