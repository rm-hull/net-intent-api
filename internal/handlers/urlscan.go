package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rm-hull/net-intent-api/internal/service/urlscan"
)

func UrlScan(svc *urlscan.Service) gin.HandlerFunc {
	return func(c *gin.Context) {

		var query QueryParams
		if err := c.ShouldBindQuery(&query); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid query parameters", "details": err.Error()})
			return
		}

		result, err := svc.GetLatestResult(c.Request.Context(), query.Domain)
		if err != nil {
			_ = c.Error(err)
			c.AbortWithStatusJSON(http.StatusBadGateway, gin.H{"error": "failed to scan: " + query.Domain, "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}
