package service

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/ArminDashti/as-ip/server/internal/model"
)

// Team Cymru's public DNS service answers "which ASN originates this IP?" from
// BGP data, which covers announced ranges that the RIR-derived prefix dataset
// (ipverse/as-ip-blocks) misses — Cloudflare anycast/WARP blocks for example.
const (
	cymruOriginV4Zone = "origin.asn.cymru.com"
	cymruOriginV6Zone = "origin6.asn.cymru.com"
	cymruAsnZone      = "asn.cymru.com"
	cymruQueryBudget  = 3 * time.Second
	cymruSource       = "cymru-dns"
)

// cymruOriginName builds the reversed-octet (or reversed-nibble) lookup name
// for an IP address.
func cymruOriginName(ipStr string) (string, error) {
	ip := net.ParseIP(strings.TrimSpace(ipStr))
	if ip == nil {
		return "", fmt.Errorf("invalid ip: %s", ipStr)
	}

	if v4 := ip.To4(); v4 != nil {
		return fmt.Sprintf("%d.%d.%d.%d.%s", v4[3], v4[2], v4[1], v4[0], cymruOriginV4Zone), nil
	}

	expanded := ip.To16()
	if expanded == nil {
		return "", fmt.Errorf("invalid ip: %s", ipStr)
	}
	nibbles := make([]string, 0, 32)
	for i := len(expanded) - 1; i >= 0; i-- {
		nibbles = append(nibbles, fmt.Sprintf("%x", expanded[i]&0x0f))
		nibbles = append(nibbles, fmt.Sprintf("%x", expanded[i]>>4))
	}
	return strings.Join(nibbles, ".") + "." + cymruOriginV6Zone, nil
}

// parseCymruOriginRecords extracts the origin ASN and country code from a
// Cymru origin TXT answer: "13335 | 172.68.205.0/24 | US | arin | 2015-02-25".
func parseCymruOriginRecords(records []string) (int, string, bool) {
	for _, record := range records {
		fields := splitCymruFields(record)
		if len(fields) < 3 {
			continue
		}

		asnField := fields[0]
		if idx := strings.IndexAny(asnField, " \t"); idx >= 0 {
			asnField = asnField[:idx]
		}
		asn, err := strconv.Atoi(strings.TrimSpace(asnField))
		if err != nil || asn <= 0 {
			continue
		}

		return asn, strings.ToUpper(strings.TrimSpace(fields[2])), true
	}
	return 0, "", false
}

// parseCymruAsnRecords extracts the AS name and country code from a Cymru ASN
// TXT answer: "13335 | US | arin | 2010-07-14 | CLOUDFLARENET - Cloudflare, Inc., US".
func parseCymruAsnRecords(records []string) (string, string, bool) {
	for _, record := range records {
		fields := splitCymruFields(record)
		if len(fields) < 5 {
			continue
		}

		name := strings.TrimSpace(fields[4])
		countryCode := strings.ToUpper(strings.TrimSpace(fields[1]))
		if countryCode != "" {
			// The registry name is suffixed with the country code, e.g. "..., US".
			suffix := ", " + countryCode
			name = strings.TrimSuffix(name, suffix)
		}
		if name == "" {
			continue
		}
		return name, countryCode, true
	}
	return "", "", false
}

func splitCymruFields(record string) []string {
	raw := strings.Split(record, "|")
	fields := make([]string, 0, len(raw))
	for _, field := range raw {
		fields = append(fields, strings.TrimSpace(field))
	}
	return fields
}

// lookupCymru resolves an IP to its origin ASN/AS/country through Cymru DNS.
// Any failure is a soft miss: callers fall back to whatever the local data has.
func (s *LookupService) lookupCymru(ctx context.Context, ip string) (model.IpAttribution, bool) {
	name, err := cymruOriginName(ip)
	if err != nil {
		return model.IpAttribution{}, false
	}

	resolver := s.resolverOrDefault()
	lookupCtx, cancel := context.WithTimeout(ctx, cymruQueryBudget)
	defer cancel()

	records, err := resolver.LookupTXT(lookupCtx, name)
	if err != nil {
		return model.IpAttribution{}, false
	}

	asn, countryCode, ok := parseCymruOriginRecords(records)
	if !ok {
		return model.IpAttribution{}, false
	}

	attribution := model.IpAttribution{
		Ip:          ip,
		Asn:         asn,
		CountryCode: countryCode,
		Source:      cymruSource,
		ResolvedAt:  time.Now().UTC(),
	}

	if s.repository != nil {
		if asName, nameErr := s.repository.FindAsnName(lookupCtx, asn); nameErr == nil && asName != "" {
			attribution.AsName = asName
		} else if asName, _, nameOK := s.lookupCymruAsName(lookupCtx, resolver, asn); nameOK {
			attribution.AsName = asName
		}
		if countryName, countryErr := s.repository.FindCountryNameByCode(lookupCtx, countryCode); countryErr == nil {
			attribution.Country = countryName
		}
	}

	return attribution, true
}

func (s *LookupService) lookupCymruAsName(ctx context.Context, resolver dnsResolver, asn int) (string, string, bool) {
	records, err := resolver.LookupTXT(ctx, fmt.Sprintf("AS%d.%s", asn, cymruAsnZone))
	if err != nil {
		return "", "", false
	}
	return parseCymruAsnRecords(records)
}
