package dto

type AsnMetadata struct {
	Handle       string  `json:"handle"`
	Description  string  `json:"description"`
	CountryCode  string  `json:"countryCode"`
	Country      string  `json:"country"`
	Origin       *string `json:"origin"`
	Category     *string `json:"category"`
	Registered   *string `json:"registered"`
	LastModified *string `json:"lastModified"`
}

type Ipv4Stats struct {
	Prefixes           int   `json:"prefixes"`
	PrefixesAggregated int   `json:"prefixesAggregated"`
	LargestPrefix      *int  `json:"largestPrefix"`
	TotalAddresses     int64 `json:"totalAddresses"`
}

type ConnectivityStats struct {
	Providers    int   `json:"providers"`
	ProviderAsns []int `json:"providerAsns"`
}

type AsnStats struct {
	Ipv4         Ipv4Stats         `json:"ipv4"`
	Connectivity ConnectivityStats `json:"connectivity"`
}

// AsnResponse matches the legacy NestJS AsnResponseDto shape for compatibility.
type AsnResponse struct {
	Asn           int         `json:"asn"`
	Metadata      AsnMetadata `json:"metadata"`
	Stats         AsnStats    `json:"stats"`
	LastAnnounced *string     `json:"lastAnnounced"`
}

type AsnListResponse struct {
	Items []AsnResponse `json:"items"`
	Total int           `json:"total"`
}

type AsnItem struct {
	Asn    int    `json:"asn"`
	Handle string `json:"handle"`
}

type AsnCatalogResponse struct {
	Items []AsnItem `json:"items"`
	Total int       `json:"total"`
}
