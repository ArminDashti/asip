package model

import "time"

type AsRecord struct {
	ID             int
	Name           string
	Category       *string
	Registered     *time.Time
	Origin         *string
	LastModified   *time.Time
	LastAnnounced  *time.Time
	CountryCode    *string
	CountryName    *string
	AsnNumber      int
	PrefixCount    int
	PrefixesAgg    int
	LargestPrefix  *int
	TotalAddresses int64
	ProviderAsns   []int
	Prefixes       []string
}
