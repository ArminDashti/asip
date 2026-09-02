package handler

import (
	"net/http"
	"strings"

	"github.com/ArminDashti/as-ip/server/internal/service"
	"github.com/gin-gonic/gin"
)

type LookupHandler struct {
	service *service.LookupService
}

func NewLookupHandler(service *service.LookupService) *LookupHandler {
	return &LookupHandler{service: service}
}

func (h *LookupHandler) GetIPInfo(c *gin.Context) {
	ip := strings.TrimSpace(c.Param("ip"))
	if ip == "" {
		ip = c.ClientIP()
	}

	response, err := h.service.GetIPInfo(c.Request.Context(), ip)
	if err != nil {
		respondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, response)
}

func (h *LookupHandler) ListAsns(c *gin.Context) {
	response, err := h.service.ListAsns(c.Request.Context())
	if err != nil {
		respondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, response)
}

func (h *LookupHandler) SearchAsn(c *gin.Context) {
	response, err := h.service.SearchAsn(c.Request.Context(), c.Param("asn"))
	if err != nil {
		respondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, response)
}

func (h *LookupHandler) MapAsnToAs(c *gin.Context) {
	response, err := h.service.MapAsnToAs(c.Request.Context(), c.Param("asn"))
	if err != nil {
		respondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, response)
}

func (h *LookupHandler) SearchAs(c *gin.Context) {
	response, err := h.service.SearchAs(c.Request.Context(), c.Param("as"))
	if err != nil {
		respondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, response)
}

func (h *LookupHandler) ListCountries(c *gin.Context) {
	response, err := h.service.ListCountries(c.Request.Context())
	if err != nil {
		respondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, response)
}

func (h *LookupHandler) SearchCountry(c *gin.Context) {
	response, err := h.service.SearchCountry(c.Request.Context(), c.Param("country"))
	if err != nil {
		respondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, response)
}

func (h *LookupHandler) LookupDns(c *gin.Context) {
	domain := strings.TrimPrefix(c.Param("domain"), "/")
	response, err := h.service.LookupDns(c.Request.Context(), domain)
	if err != nil {
		respondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, response)
}

func (h *LookupHandler) GetIPWhois(c *gin.Context) {
	ip := strings.TrimSpace(c.Param("ip"))
	if ip == "" {
		ip = c.ClientIP()
	}

	response, err := h.service.GetIPWhois(c.Request.Context(), ip)
	if err != nil {
		respondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, response)
}
