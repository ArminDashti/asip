package service

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/ArminDashti/as-ip/server/internal/dto"
	"github.com/ArminDashti/as-ip/server/internal/mapper"
	"github.com/ArminDashti/as-ip/server/internal/repository"
)

func (s *LookupService) GetIPInfo(ctx context.Context, ip string) (dto.IpInfoResponse, error) {
	normalizedIP := strings.TrimSpace(ip)
	if normalizedIP == "" {
		return dto.IpInfoResponse{}, fmt.Errorf("ip is required: %w", ErrBadRequest)
	}
	if net.ParseIP(normalizedIP) == nil {
		return dto.IpInfoResponse{}, fmt.Errorf("invalid ip address: %w", ErrBadRequest)
	}

	record, err := s.repository.FindByIP(ctx, normalizedIP)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return dto.IpInfoResponse{}, fmt.Errorf("no ASN mapping found for IPv4 address %s: %w", normalizedIP, ErrNotFound)
		}
		return dto.IpInfoResponse{}, err
	}

	countryName := ""
	if _, geoCountry, geoErr := s.repository.FindCountryByIP(ctx, normalizedIP); geoErr == nil {
		countryName = geoCountry
	} else if !errors.Is(geoErr, repository.ErrNotFound) {
		return dto.IpInfoResponse{}, geoErr
	} else if record.CountryName != nil {
		countryName = *record.CountryName
	}

	return mapper.ToIpInfoResponse(normalizedIP, record, countryName), nil
}
