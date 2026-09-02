package mapper

import (
	"time"

	"github.com/ArminDashti/as-ip/server/internal/dto"
	"github.com/ArminDashti/as-ip/server/internal/model"
)

func formatDate(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.Format("2006-01-02")
	return &formatted
}

func ToAsnResponse(record model.AsRecord) dto.AsnResponse {
	metadata := dto.AsnMetadata{
		Handle:       record.Name,
		Description:  record.Name,
		Origin:       record.Origin,
		Category:     record.Category,
		Registered:   formatDate(record.Registered),
		LastModified: formatDate(record.LastModified),
	}

	if record.CountryCode != nil {
		metadata.CountryCode = *record.CountryCode
	}
	if record.CountryName != nil {
		metadata.Country = *record.CountryName
	}

	providerAsns := record.ProviderAsns
	if providerAsns == nil {
		providerAsns = []int{}
	}

	return dto.AsnResponse{
		Asn: record.AsnNumber,
		Metadata: metadata,
		Stats: dto.AsnStats{
			Ipv4: dto.Ipv4Stats{
				Prefixes:           record.PrefixCount,
				PrefixesAggregated: record.PrefixesAgg,
				LargestPrefix:      record.LargestPrefix,
				TotalAddresses:     record.TotalAddresses,
			},
			Connectivity: dto.ConnectivityStats{
				Providers:    len(providerAsns),
				ProviderAsns: providerAsns,
			},
		},
		LastAnnounced: formatDate(record.LastAnnounced),
	}
}

func ToAsSummary(record model.AsRecord) dto.AsSummary {
	return dto.AsSummary{
		ID:           record.ID,
		Handle:       record.Name,
		Organization: record.Name,
		CountryCode:  record.CountryCode,
		Country:      record.CountryName,
	}
}

func ToIpInfoResponse(ip string, record model.AsRecord, countryName string) dto.IpInfoResponse {
	return dto.IpInfoResponse{
		Ip:      ip,
		Asn:     record.AsnNumber,
		As:      record.Name,
		Country: countryName,
	}
}

func ToIpPrefixCollection(record model.AsRecord) dto.IpPrefixCollection {
	countryCode := ""
	countryName := ""
	if record.CountryCode != nil {
		countryCode = *record.CountryCode
	}
	if record.CountryName != nil {
		countryName = *record.CountryName
	}

	prefixes := record.Prefixes
	if prefixes == nil {
		prefixes = []string{}
	}

	return dto.IpPrefixCollection{
		Asn:         record.AsnNumber,
		Handle:      record.Name,
		CountryCode: countryCode,
		Country:     countryName,
		Prefixes:    prefixes,
	}
}
