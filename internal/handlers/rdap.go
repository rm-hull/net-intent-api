package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	domain_rdap "github.com/openrdap/rdap"
	client "github.com/rm-hull/net-intent-api/internal/clients/rdap"
	"github.com/rm-hull/net-intent-api/internal/service/rdap"
)

func RdapHandler(svc *rdap.Service) gin.HandlerFunc {
	return func(c *gin.Context) {

		var query QueryParams
		if err := c.ShouldBindQuery(&query); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid query parameters", "details": err.Error()})
			return
		}

		resp, err := svc.GetDomain(c.Request.Context(), query.Domain)
		if err != nil {
			if client.ClassifyError(err) {
				c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": client.ErrorMessage(query.Domain, err)})
				return
			}
			_ = c.Error(err)
			c.AbortWithStatusJSON(http.StatusBadGateway, gin.H{"error": "failed to lookup: " + query.Domain, "details": err.Error()})
			return
		}

		c.Header("Content-Type", "application/rdap+json")
		c.JSON(http.StatusOK, toDomainResponse(resp))
	}
}

type domainResponse struct {
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

func toDomainResponse(d *domain_rdap.Domain) domainResponse {
	out := domainResponse{
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

	return out
}
