// Local test API without PostgreSQL: headers, whois, and stub IP/DNS for UI testing.
package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ArminDashti/as-ip/server/internal/handler"
	"github.com/ArminDashti/as-ip/server/internal/service"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())
	engine.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:5173",
			"http://127.0.0.1:5173",
		},
		AllowMethods:     []string{"GET", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	lookupService := service.NewLookupService(nil)
	lookupHandler := handler.NewLookupHandler(lookupService)
	headersHandler := handler.NewHeadersHandler()
	docsHandler := handler.NewDocsHandler()

	api := engine.Group("/api/v1")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "as-ip-localtest"})
		})
		api.GET("/docs", docsHandler.GetDocs)
		api.GET("/http/headers", headersHandler.GetHttpHeaders)
		api.GET("/ip/whois", lookupHandler.GetIPWhois)
		api.GET("/ip/whois/:ip", lookupHandler.GetIPWhois)
		api.GET("/ip/info", stubIPInfo)
		api.GET("/ip/info/:ip", stubIPInfo)
		api.GET("/dns/lookup/*domain", stubDNS)
	}

	addr := ":" + port
	log.Printf("as-ip localtest listening on %s (no database)", addr)
	if err := engine.Run(addr); err != nil {
		log.Fatalf("listen: %v", err)
	}
}

func stubIPInfo(c *gin.Context) {
	ip := strings.TrimSpace(c.Param("ip"))
	if ip == "" {
		ip = c.ClientIP()
	}
	if ip == "" || ip == "::1" {
		ip = "127.0.0.1"
	}
	c.JSON(http.StatusOK, gin.H{
		"ip":      ip,
		"asn":     64512,
		"as":      "LOCALTEST-AS",
		"country": "United States",
	})
}

func stubDNS(c *gin.Context) {
	domain := strings.TrimPrefix(c.Param("domain"), "/")
	if domain == "" {
		domain = "example.com"
	}
	c.JSON(http.StatusOK, gin.H{
		"domain":  domain,
		"a":       []string{"93.184.216.34"},
		"aaaa":    []string{},
		"ns":      []string{"a.iana-servers.net."},
		"mx":      []any{},
		"txt":     []string{},
		"cname":   "",
		"asn":     15133,
		"as":      "EDGECAST",
		"country": "United States",
		"addresses": []gin.H{
			{
				"ip":      "93.184.216.34",
				"asn":     15133,
				"as":      "EDGECAST",
				"country": "United States",
			},
		},
	})
}
