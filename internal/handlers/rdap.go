package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
		c.JSON(http.StatusOK, resp)
	}
}
