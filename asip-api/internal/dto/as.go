package dto

type AsSummary struct {
	ID           int     `json:"id"`
	Handle       string  `json:"handle"`
	Organization string  `json:"organization"`
	CountryCode  *string `json:"countryCode,omitempty"`
	Country      *string `json:"country,omitempty"`
	City         *string `json:"city,omitempty"`
	Isp          *string `json:"isp,omitempty"`
}

type AsMappingResponse struct {
	Asn int       `json:"asn"`
	As  AsSummary `json:"as"`
}

type IpPrefixCollection struct {
	Asn         int      `json:"asn"`
	Handle      string   `json:"handle"`
	CountryCode string   `json:"countryCode"`
	Country     string   `json:"country"`
	Prefixes    []string `json:"prefixes"`
}
