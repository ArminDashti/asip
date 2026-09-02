package router

import (
	"github.com/ArminDashti/as-ip/server/internal/handler"
	"github.com/ArminDashti/as-ip/server/internal/repository"
	"github.com/gin-gonic/gin"
)

func Register(engine *gin.Engine, healthHandler *handler.HealthHandler, docsHandler *handler.DocsHandler, lookupHandler *handler.LookupHandler, headersHandler *handler.HeadersHandler, repo *repository.AsRepository) {
	engine.Use(requestLogger(repo))

	api := engine.Group("/api/v1")
	{
		api.GET("/docs", docsHandler.GetDocs)
		api.GET("/health", healthHandler.GetHealth)

		api.GET("/ip/info", lookupHandler.GetIPInfo)
		api.GET("/ip/info/:ip", lookupHandler.GetIPInfo)
		api.GET("/ip/whois", lookupHandler.GetIPWhois)
		api.GET("/ip/whois/:ip", lookupHandler.GetIPWhois)
		api.GET("/dns/lookup/*domain", lookupHandler.LookupDns)
		api.GET("/http/headers", headersHandler.GetHttpHeaders)

		asnGroup := api.Group("/asn")
		{
			asnGroup.GET("/list", lookupHandler.ListAsns)
			asnGroup.GET("/search/:asn", lookupHandler.SearchAsn)
			asnGroup.GET("/to-as/:asn", lookupHandler.MapAsnToAs)
		}

		asGroup := api.Group("/as")
		{
			asGroup.GET("/search/:as", lookupHandler.SearchAs)
		}

		countryGroup := api.Group("/country")
		{
			countryGroup.GET("/list", lookupHandler.ListCountries)
			countryGroup.GET("/search/:country", lookupHandler.SearchCountry)
		}
	}
}

func requestLogger(repo *repository.AsRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		path := c.Request.URL.Path
		if path == "/api/v1/health" || path == "/api/v1/docs" || path == "/metrics" {
			return
		}
		_ = repo.LogRequest(c.Request.Context(), c.ClientIP())
	}
}
