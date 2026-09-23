package service

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/ArminDashti/as-ip/server/internal/dto"
	"github.com/ArminDashti/as-ip/server/internal/mapper"
)

// GetIPInfo resolves the AS/country attribution of an IP address.
//
// A valid IP always yields a response: when the imported dataset has no prefix
// mapping for it, the lookup falls back to the latest stored attribution and
// then to a live ASN lookup, so AS/Country are still reported. Only malformed
// or missing input is rejected.
func (s *LookupService) GetIPInfo(ctx context.Context, ip string) (dto.IpInfoResponse, error) {
	normalizedIP := strings.TrimSpace(ip)
	if normalizedIP == "" {
		return dto.IpInfoResponse{}, fmt.Errorf("ip is required: %w", ErrBadRequest)
	}

	parsed := net.ParseIP(normalizedIP)
	if parsed == nil {
		return dto.IpInfoResponse{}, fmt.Errorf("invalid ip address: %w", ErrBadRequest)
	}
	normalizedIP = parsed.String()

	attribution := s.resolveIpAttribution(ctx, normalizedIP, true)
	return mapper.ToIpInfoResponse(normalizedIP, attribution), nil
}
