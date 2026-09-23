package model

import "time"

// IpAttribution is the latest known AS/country attribution for a single IP.
//
// Only the newest value per IP is kept: rows are upserted in place and older
// values are overwritten, so there is no history/archive to maintain.
type IpAttribution struct {
	Ip          string
	Asn         int
	AsName      string
	CountryCode string
	Country     string
	Source      string
	ResolvedAt  time.Time
}

// IsComplete reports whether both the ASN and a country name are known.
func (a IpAttribution) IsComplete() bool {
	return a.Asn != 0 && a.Country != ""
}
