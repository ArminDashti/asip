package sync

import "time"

type metadataEntry struct {
	Asn           int              `json:"asn"`
	Metadata      metadataFields   `json:"metadata"`
	Stats         *statsBlock      `json:"stats"`
	LastAnnounced *string          `json:"lastAnnounced"`
}

type metadataFields struct {
	Handle       string  `json:"handle"`
	Description  string  `json:"description"`
	CountryCode  string  `json:"countryCode"`
	Country      string  `json:"country"`
	Origin       *string `json:"origin"`
	Category     *string `json:"category"`
	Registered   *string `json:"registered"`
	LastModified *string `json:"lastModified"`
}

type statsBlock struct {
	Ipv4         *ipv4StatsBlock         `json:"ipv4"`
	Ipv6         *ipv6StatsBlock         `json:"ipv6"`
	Connectivity *connectivityStatsBlock `json:"connectivity"`
}

type ipv4StatsBlock struct {
	Prefixes           int    `json:"prefixes"`
	PrefixesAggregated int    `json:"prefixesAggregated"`
	LargestPrefix      *int   `json:"largestPrefix"`
	TotalAddresses     *int64 `json:"totalAddresses"`
}

type ipv6StatsBlock struct {
	Prefixes           int    `json:"prefixes"`
	PrefixesAggregated int    `json:"prefixesAggregated"`
	LargestPrefix      *int   `json:"largestPrefix"`
	TotalAddresses     *int64 `json:"totalAddresses"`
}

type connectivityStatsBlock struct {
	ProviderAsns []int `json:"providerAsns"`
}

type asPrefixEntry struct {
	Asn      int              `json:"asn"`
	Prefixes asPrefixFamilies `json:"prefixes"`
}

type asPrefixFamilies struct {
	Ipv4 []string `json:"ipv4"`
	Ipv6 []string `json:"ipv6"`
}

type geoCountryEntry struct {
	CountryCode string            `json:"countryCode"`
	Country     string            `json:"country"`
	Prefixes    geoPrefixFamilies `json:"prefixes"`
}

type geoPrefixFamilies struct {
	Ipv4 []string `json:"ipv4"`
	Ipv6 []string `json:"ipv6"`
}

type countryRow struct {
	Code string
	Name string
}

type asnRow struct {
	Number              int
	Name                string
	Category            *string
	Origin              *string
	Registered          *time.Time
	LastModified        *time.Time
	LastAnnounced       *time.Time
	CountryCode         string
	Ipv4Prefixes        int
	Ipv4PrefixesAgg     int
	Ipv4LargestPrefix   *int
	Ipv4TotalAddresses  int64
	Ipv6Prefixes        int
	Ipv6PrefixesAgg     int
	Ipv6LargestPrefix   *int
	Ipv6TotalAddresses  int64
}

type connectivityRelation struct {
	AsnNumber         int
	ProviderAsnNumber int
}

func parseDate(value *string) *time.Time {
	if value == nil || *value == "" {
		return nil
	}
	parsed, err := time.Parse("2006-01-02", *value)
	if err != nil {
		return nil
	}
	return &parsed
}
