package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rm-hull/net-intent-api/internal/urlscan"
)

func UrlScan(client *urlscan.Client) gin.HandlerFunc {
	return func(c *gin.Context) {

		var query QueryParams
		if err := c.ShouldBindQuery(&query); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid query parameters", "details": err.Error()})
			return
		}

		resp, err := client.Search(query.Domain, 1)
		if err != nil {
			_ = c.Error(err)
			c.AbortWithStatusJSON(http.StatusBadGateway, gin.H{"error": "failed to scan: " + query.Domain, "details": err.Error()})
			return
		}

		if len(resp.Results) == 0 || resp.Total == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "no results"})
		}

		c.JSON(http.StatusOK, resp.Results[0])
	}
}
