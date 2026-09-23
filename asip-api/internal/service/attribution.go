package service

import (
	"context"
	"time"

	"github.com/ArminDashti/as-ip/server/internal/model"
)

// attributionTTL bounds how long a cached (non-dataset) attribution is trusted
// before it is resolved again, so the stored AS/country stays the latest one.
const attributionTTL = 30 * 24 * time.Hour

// resolveIpAttribution answers "which AS and country owns this IP?" for a valid
// IP address, in this order:
//
//  1. the imported dataset (asn/asn_ipv4 + country_ipv4),
//  2. the latest stored attribution row for the IP,
//  3. a live Team Cymru DNS lookup (only when allowLive is true).
//
// Missing pieces never fail the request: the caller receives whatever is known.
// Resolved values are upserted, so a single row per IP is kept — there is no
// archive/history table.
func (s *LookupService) resolveIpAttribution(ctx context.Context, ip string, allowLive bool) model.IpAttribution {
	attribution := model.IpAttribution{Ip: ip}
	if s.repository == nil {
		return attribution
	}

	if record, err := s.repository.FindByIP(ctx, ip); err == nil {
		attribution.Asn = record.AsnNumber
		attribution.AsName = record.Name
		if record.CountryCode != nil {
			attribution.CountryCode = *record.CountryCode
		}
		if record.CountryName != nil {
			attribution.Country = *record.CountryName
		}
		attribution.Source = "dataset"
	}

	// The geo dataset is prefix-level, so its country beats the ASN's registered country.
	if code, name, geoErr := s.repository.FindCountryByIP(ctx, ip); geoErr == nil {
		if code != "" {
			attribution.CountryCode = code
		}
		if name != "" {
			attribution.Country = name
		}
		if attribution.Source == "" {
			attribution.Source = "dataset"
		}
	}

	if attribution.IsComplete() {
		return attribution
	}

	if cached, err := s.repository.FindAttribution(ctx, ip); err == nil {
		if time.Since(cached.ResolvedAt) < attributionTTL {
			attribution = mergeIpAttribution(attribution, cached)
			if attribution.IsComplete() {
				return attribution
			}
		}
	}

	if !allowLive {
		return attribution
	}

	live, ok := s.lookupCymru(ctx, ip)
	if !ok {
		return attribution
	}

	attribution = mergeIpAttribution(attribution, live)
	_ = s.repository.UpsertAttribution(ctx, attribution)
	return attribution
}

// mergeIpAttribution fills the empty fields of base from extra. Values already
// present in base always win. Source tracks where the ASN came from, and
// ResolvedAt is refreshed whenever extra contributed anything.
func mergeIpAttribution(base, extra model.IpAttribution) model.IpAttribution {
	merged := base
	asnFromExtra := false
	anyFromExtra := false

	if merged.Asn == 0 && extra.Asn != 0 {
		merged.Asn = extra.Asn
		asnFromExtra = true
		anyFromExtra = true
	}
	if merged.AsName == "" && extra.AsName != "" {
		merged.AsName = extra.AsName
		anyFromExtra = true
	}
	if merged.CountryCode == "" && extra.CountryCode != "" {
		merged.CountryCode = extra.CountryCode
		anyFromExtra = true
	}
	if merged.Country == "" && extra.Country != "" {
		merged.Country = extra.Country
		anyFromExtra = true
	}

	if asnFromExtra || merged.Source == "" {
		merged.Source = extra.Source
	}
	if anyFromExtra {
		merged.ResolvedAt = extra.ResolvedAt
	}

	return merged
}
