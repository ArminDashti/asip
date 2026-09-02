package dto

type CountryItem struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type CountryListResponse struct {
	Items []CountryItem `json:"items"`
	Total int           `json:"total"`
}

type CityItem struct {
	Name        string `json:"name"`
	CountryCode string `json:"countryCode"`
}

type RegionItem struct {
	Name        string  `json:"name"`
	CountryCode string  `json:"countryCode"`
	Code        *string `json:"code,omitempty"`
}

type CityListResponse struct {
	Items []CityItem `json:"items"`
	Total int        `json:"total"`
}

type RegionListResponse struct {
	Items []RegionItem `json:"items"`
	Total int          `json:"total"`
}
