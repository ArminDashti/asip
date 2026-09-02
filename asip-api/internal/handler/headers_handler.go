package handler

import (
	"net/http"
	"sort"
	"strings"

	"github.com/ArminDashti/as-ip/server/internal/dto"
	"github.com/gin-gonic/gin"
)

var omittedRequestHeaders = map[string]struct{}{
	"Connection":        {},
	"Content-Length":    {},
	"Transfer-Encoding": {},
	"Keep-Alive":        {},
	"Proxy-Connection":  {},
	"Upgrade":           {},
}

var redactedRequestHeaders = map[string]struct{}{
	"Cookie":        {},
	"Authorization": {},
}

type HeadersHandler struct{}

func NewHeadersHandler() *HeadersHandler {
	return &HeadersHandler{}
}

func (h *HeadersHandler) GetHttpHeaders(c *gin.Context) {
	names := make([]string, 0, len(c.Request.Header))
	for name := range c.Request.Header {
		canonical := http.CanonicalHeaderKey(name)
		if _, omit := omittedRequestHeaders[canonical]; omit {
			continue
		}
		names = append(names, canonical)
	}
	sort.Strings(names)

	entries := make([]dto.HttpHeaderEntry, 0, len(names))
	for _, name := range names {
		value := strings.Join(c.Request.Header.Values(name), ", ")
		if _, redact := redactedRequestHeaders[name]; redact {
			value = "[redacted]"
		}
		entries = append(entries, dto.HttpHeaderEntry{
			Name:  name,
			Value: value,
		})
	}

	c.JSON(http.StatusOK, dto.HttpHeadersResponse{
		Ip:      c.ClientIP(),
		Headers: entries,
	})
}
