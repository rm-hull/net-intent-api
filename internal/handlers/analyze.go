package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	llmprovider "github.com/rm-hull/net-intent-api/internal/clients/llm_provider"
	"github.com/rm-hull/net-intent-api/internal/config"
)

type QueryParams struct {
	Domain string `form:"domain" binding:"required,fqdn"`
}

func Analyze(cfg *config.Config, provider llmprovider.Provider) gin.HandlerFunc {
	return func(c *gin.Context) {

		var query QueryParams
		if err := c.ShouldBindQuery(&query); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid query parameters", "details": err.Error()})
			return
		}

		result, err := provider.Call(c.Request.Context(), cfg.Gemini.Prompt, query.Domain)
		if err != nil {
			_ = c.Error(err)
			c.AbortWithStatusJSON(http.StatusBadGateway, gin.H{"error": "failed to analyze: " + query.Domain, "details": err.Error()})
			return
		}

		if !json.Valid([]byte(result)) {
			c.AbortWithStatusJSON(http.StatusBadGateway, gin.H{"error": "provider returned invalid JSON"})
			return
		}

		c.Data(http.StatusOK, "application/json", []byte(result))
	}
}
