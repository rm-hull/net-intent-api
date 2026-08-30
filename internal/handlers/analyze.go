package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	llmprovider "github.com/rm-hull/net-intent-api/internal/clients/llm_provider"
	urlscan_client "github.com/rm-hull/net-intent-api/internal/clients/urlscan"
	"github.com/rm-hull/net-intent-api/internal/config"
	"github.com/rm-hull/net-intent-api/internal/service/rdap"
	"github.com/rm-hull/net-intent-api/internal/service/urlscan"
)

type QueryParams struct {
	Domain string `form:"domain" binding:"required,fqdn"`
}

func Analyze(cfg *config.Config, provider llmprovider.Provider, urlscanService *urlscan.Service, rdapService *rdap.Service) gin.HandlerFunc {
	return func(c *gin.Context) {

		var query QueryParams
		if err := c.ShouldBindQuery(&query); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid query parameters", "details": err.Error()})
			return
		}

		urlScanResult, err := urlscanService.GetLatestResult(c.Request.Context(), query.Domain)
		if err != nil {
			urlScanResult = &urlscan_client.Result{}
		}

		rdapResult, err := rdapService.GetDomain(c.Request.Context(), query.Domain)
		if err != nil {
			rdapResult = &rdap.DomainResponse{}
		}

		userPrompt := gin.H{
			"fqdn":          query.Domain,
			"country":       urlScanResult.Page.Country,
			"server":        urlScanResult.Page.Server,
			"ip":            urlScanResult.Page.IP,
			"asn":           urlScanResult.Page.ASN,
			"umbrella_rank": urlScanResult.Page.UmbrellaRank,
			"tls_issuer":    urlScanResult.Page.TLSIssuer,
			"rdap_events":   rdapResult.Events,
		}

		data, err := json.Marshal(userPrompt)
		if err != nil {
			_ = c.Error(err)
			c.AbortWithStatusJSON(http.StatusBadGateway, gin.H{"error": "failed to marshall JSON", "details": err.Error()})
			return
		}

		llmResult, err := provider.Call(c.Request.Context(), cfg.Gemini.SystemPrompt, string(data))
		if err != nil {
			_ = c.Error(err)
			c.AbortWithStatusJSON(http.StatusBadGateway, gin.H{"error": "failed to analyze: " + query.Domain, "details": err.Error()})
			return
		}

		if !json.Valid([]byte(llmResult)) {
			c.AbortWithStatusJSON(http.StatusBadGateway, gin.H{"error": "LLM provider returned invalid JSON"})
			return
		}

		c.Data(http.StatusOK, "application/json", []byte(llmResult))
	}
}
