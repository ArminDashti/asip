package service

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/ArminDashti/as-ip/server/internal/dto"
	"github.com/ArminDashti/as-ip/server/internal/repository"
)

const (
	dnsLookupBudget      = 8 * time.Second
	dnsRecordQueryBudget = 2 * time.Second
)

type dnsResolver interface {
	LookupIP(ctx context.Context, network, host string) ([]net.IP, error)
	LookupNS(ctx context.Context, name string) ([]*net.NS, error)
	LookupMX(ctx context.Context, name string) ([]*net.MX, error)
	LookupTXT(ctx context.Context, name string) ([]string, error)
	LookupCNAME(ctx context.Context, name string) (string, error)
}

type stdDNSResolver struct{}

func (stdDNSResolver) LookupIP(ctx context.Context, network, host string) ([]net.IP, error) {
	return net.DefaultResolver.LookupIP(ctx, network, host)
}

func (stdDNSResolver) LookupNS(ctx context.Context, name string) ([]*net.NS, error) {
	return net.DefaultResolver.LookupNS(ctx, name)
}

func (stdDNSResolver) LookupMX(ctx context.Context, name string) ([]*net.MX, error) {
	return net.DefaultResolver.LookupMX(ctx, name)
}

func (stdDNSResolver) LookupTXT(ctx context.Context, name string) ([]string, error) {
	return net.DefaultResolver.LookupTXT(ctx, name)
}

func (stdDNSResolver) LookupCNAME(ctx context.Context, name string) (string, error) {
	return net.DefaultResolver.LookupCNAME(ctx, name)
}

func (s *LookupService) resolverOrDefault() dnsResolver {
	if s.resolver != nil {
		return s.resolver
	}
	return stdDNSResolver{}
}

func (s *LookupService) LookupDns(ctx context.Context, domain string) (dto.DnsLookupResponse, error) {
	normalizedDomain, err := normalizeDomain(domain)
	if err != nil {
		return dto.DnsLookupResponse{}, err
	}

	lookupCtx, cancel := context.WithTimeout(ctx, dnsLookupBudget)
	defer cancel()

	resolver := s.resolverOrDefault()
	response := dto.DnsLookupResponse{
		Domain:    normalizedDomain,
		A:         []string{},
		AAAA:      []string{},
		NS:        []string{},
		MX:        []dto.DnsMxRecord{},
		TXT:       []string{},
		CNAME:     "",
		Addresses: []dto.DnsAddressInfo{},
	}

	if ips, lookupErr := lookupWithTimeout(lookupCtx, resolver.LookupIP, "ip", normalizedDomain); lookupErr == nil {
		for _, ip := range ips {
			if v4 := ip.To4(); v4 != nil {
				response.A = append(response.A, v4.String())
			} else {
				response.AAAA = append(response.AAAA, ip.String())
			}
		}
	} else if !isDNSRecordMiss(lookupErr) {
		return dto.DnsLookupResponse{}, lookupErr
	}

	if nsRecords, lookupErr := lookupNameWithTimeout(lookupCtx, resolver.LookupNS, normalizedDomain); lookupErr == nil {
		for _, ns := range nsRecords {
			response.NS = append(response.NS, ns.Host)
		}
	} else if !isDNSRecordMiss(lookupErr) {
		return dto.DnsLookupResponse{}, lookupErr
	}

	if mxRecords, lookupErr := lookupNameWithTimeout(lookupCtx, resolver.LookupMX, normalizedDomain); lookupErr == nil {
		for _, mx := range mxRecords {
			response.MX = append(response.MX, dto.DnsMxRecord{Host: mx.Host, Pref: mx.Pref})
		}
	} else if !isDNSRecordMiss(lookupErr) {
		return dto.DnsLookupResponse{}, lookupErr
	}

	if txtRecords, lookupErr := lookupNameWithTimeout(lookupCtx, resolver.LookupTXT, normalizedDomain); lookupErr == nil {
		response.TXT = append(response.TXT, txtRecords...)
	} else if !isDNSRecordMiss(lookupErr) {
		return dto.DnsLookupResponse{}, lookupErr
	}

	if cname, lookupErr := lookupNameWithTimeout(lookupCtx, resolver.LookupCNAME, normalizedDomain); lookupErr == nil {
		if cname != normalizedDomain+"." && cname != normalizedDomain {
			response.CNAME = cname
		}
	} else if !isDNSRecordMiss(lookupErr) {
		return dto.DnsLookupResponse{}, lookupErr
	}

	if len(response.A) == 0 && len(response.AAAA) == 0 &&
		len(response.NS) == 0 && len(response.MX) == 0 &&
		len(response.TXT) == 0 && response.CNAME == "" {
		return dto.DnsLookupResponse{}, fmt.Errorf("no DNS records found for %s: %w", normalizedDomain, ErrNotFound)
	}

	orderedIPs := append(append([]string{}, response.A...), response.AAAA...)
	for _, ip := range orderedIPs {
		response.Addresses = append(response.Addresses, s.enrichDnsAddress(lookupCtx, ip))
	}

	for _, address := range response.Addresses {
		if address.Asn != 0 || address.As != "" || address.Country != "" {
			response.Asn = address.Asn
			response.As = address.As
			response.Country = address.Country
			break
		}
	}
	if response.Asn == 0 && response.As == "" && response.Country == "" && len(response.Addresses) > 0 {
		primary := response.Addresses[0]
		response.Asn = primary.Asn
		response.As = primary.As
		response.Country = primary.Country
	}

	return response, nil
}

func (s *LookupService) enrichDnsAddress(ctx context.Context, ip string) dto.DnsAddressInfo {
	info := dto.DnsAddressInfo{Ip: ip}
	if s.repository == nil || net.ParseIP(ip).To4() == nil {
		return info
	}

	record, err := s.repository.FindByIP(ctx, ip)
	if err != nil {
		return info
	}

	countryName := ""
	if _, geoCountry, geoErr := s.repository.FindCountryByIP(ctx, ip); geoErr == nil {
		countryName = geoCountry
	} else if errors.Is(geoErr, repository.ErrNotFound) && record.CountryName != nil {
		countryName = *record.CountryName
	} else if geoErr != nil && !errors.Is(geoErr, repository.ErrNotFound) {
		return info
	} else if record.CountryName != nil {
		countryName = *record.CountryName
	}

	info.Asn = record.AsnNumber
	info.As = record.Name
	info.Country = countryName
	return info
}

func isDNSRecordMiss(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return true
	}
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return dnsErr.IsNotFound || dnsErr.IsTimeout || dnsErr.IsTemporary
	}
	return false
}

func lookupWithTimeout(
	parent context.Context,
	fn func(ctx context.Context, network, host string) ([]net.IP, error),
	network, host string,
) ([]net.IP, error) {
	ctx, cancel := context.WithTimeout(parent, dnsRecordQueryBudget)
	defer cancel()
	return fn(ctx, network, host)
}

func lookupNameWithTimeout[T any](
	parent context.Context,
	fn func(ctx context.Context, name string) (T, error),
	name string,
) (T, error) {
	ctx, cancel := context.WithTimeout(parent, dnsRecordQueryBudget)
	defer cancel()
	return fn(ctx, name)
}
