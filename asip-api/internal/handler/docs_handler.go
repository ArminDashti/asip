package handler

import (
	"net/http"

	"github.com/ArminDashti/as-ip/server/internal/dto"
	"github.com/gin-gonic/gin"
)

type DocsHandler struct{}

func NewDocsHandler() *DocsHandler {
	return &DocsHandler{}
}

func (h *DocsHandler) GetDocs(c *gin.Context) {
	c.JSON(http.StatusOK, dto.DocsResponse{
		Service:  "as-ip",
		Version:  "v1",
		BasePath: "/api/v1",
		Endpoints: []dto.EndpointDocument{
			{
				Method:      "GET",
				Path:        "/api/v1/docs",
				Summary:     "API documentation catalog",
				Description: "Returns document-like metadata for every HTTP endpoint in this API.",
				Auth:        false,
			},
			{
				Method:      "GET",
				Path:        "/api/v1/health",
				Summary:     "Health check",
				Description: "Reports whether the API process is running.",
				Auth:        false,
			},
			{
				Method:      "GET",
				Path:        "/metrics",
				Summary:     "Prometheus metrics",
				Description: "Exposes Prometheus scrape metrics for request rate and latency.",
				Auth:        false,
			},
			{
				Method:      "GET",
				Path:        "/api/v1/ip/info",
				Summary:     "IP lookup for caller",
				Description: "Resolves geolocation and network details for the request client IP.",
				Auth:        false,
			},
			{
				Method:      "GET",
				Path:        "/api/v1/ip/info/:ip",
				Summary:     "IP lookup by address",
				Description: "Resolves geolocation and network details for the given IPv4 or IPv6 address.",
				Auth:        false,
				Parameters: []dto.EndpointParameter{
					{
						Name:        "ip",
						In:          "path",
						Required:    true,
						Description: "IPv4 or IPv6 address to look up",
					},
				},
			},
			{
				Method:      "GET",
				Path:        "/api/v1/ip/whois",
				Summary:     "IP whois for caller",
				Description: "Returns RDAP whois details for the request client IP.",
				Auth:        false,
			},
			{
				Method:      "GET",
				Path:        "/api/v1/ip/whois/:ip",
				Summary:     "IP whois by address",
				Description: "Returns RDAP whois details for the given IPv4 or IPv6 address.",
				Auth:        false,
				Parameters: []dto.EndpointParameter{
					{
						Name:        "ip",
						In:          "path",
						Required:    true,
						Description: "IPv4 or IPv6 address to look up",
					},
				},
			},
			{
				Method:      "GET",
				Path:        "/api/v1/http/headers",
				Summary:     "Echo request HTTP headers",
				Description: "Returns the client IP and the HTTP headers received by the API (Cookie and Authorization values redacted).",
				Auth:        false,
			},
			{
				Method:      "GET",
				Path:        "/api/v1/asn/list",
				Summary:     "List ASN numbers",
				Description: "Returns every known ASN number with its handle.",
				Auth:        false,
			},
			{
				Method:      "GET",
				Path:        "/api/v1/asn/search/:asn",
				Summary:     "ASN details",
				Description: "Returns ASN details looked up by ASN number or handle.",
				Auth:        false,
				Parameters: []dto.EndpointParameter{
					{
						Name:        "asn",
						In:          "path",
						Required:    true,
						Description: "ASN number (e.g. 15169) or handle",
					},
				},
			},
			{
				Method:      "GET",
				Path:        "/api/v1/asn/to-as/:asn",
				Summary:     "Map ASN to AS",
				Description: "Maps an ASN number or handle to a compact Autonomous System summary.",
				Auth:        false,
				Parameters: []dto.EndpointParameter{
					{
						Name:        "asn",
						In:          "path",
						Required:    true,
						Description: "ASN number (e.g. 15169) or handle",
					},
				},
			},
			{
				Method:      "GET",
				Path:        "/api/v1/as/search/:as",
				Summary:     "AS details",
				Description: "Returns Autonomous System details looked up by handle or number.",
				Auth:        false,
				Parameters: []dto.EndpointParameter{
					{
						Name:        "as",
						In:          "path",
						Required:    true,
						Description: "AS handle or number",
					},
				},
			},
			{
				Method:      "GET",
				Path:        "/api/v1/country/list",
				Summary:     "List countries",
				Description: "Returns every known country with its ISO code and related metadata.",
				Auth:        false,
			},
			{
				Method:      "GET",
				Path:        "/api/v1/country/search/:country",
				Summary:     "ASNs by country",
				Description: "Returns ASN entries associated with the given ISO country code.",
				Auth:        false,
				Parameters: []dto.EndpointParameter{
					{
						Name:        "country",
						In:          "path",
						Required:    true,
						Description: "ISO 3166-1 alpha-2 country code (e.g. US)",
					},
				},
			},
			{
				Method:      "GET",
				Path:        "/api/v1/dns/lookup/*domain",
				Summary:     "DNS lookup for a domain",
				Description: "Resolves DNS records (A, AAAA, NS, MX, TXT, CNAME) for a website or domain and enriches resolved IPs with ASN and country when available.",
				Auth:        false,
				Parameters: []dto.EndpointParameter{
					{
						Name:        "domain",
						In:          "path",
						Required:    true,
						Description: "Hostname or URL (e.g. example.com, https://example.com)",
					},
				},
			},
		},
	})
}
