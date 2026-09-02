package dto

type DnsMxRecord struct {
	Host string `json:"host"`
	Pref uint16 `json:"pref"`
}

type DnsAddressInfo struct {
	Ip      string `json:"ip"`
	Asn     int    `json:"asn"`
	As      string `json:"as"`
	Country string `json:"country"`
}

type DnsLookupResponse struct {
	Domain    string           `json:"domain"`
	A         []string         `json:"a"`
	AAAA      []string         `json:"aaaa"`
	NS        []string         `json:"ns"`
	MX        []DnsMxRecord    `json:"mx"`
	TXT       []string         `json:"txt"`
	CNAME     string           `json:"cname"`
	Asn       int              `json:"asn"`
	As        string           `json:"as"`
	Country   string           `json:"country"`
	Addresses []DnsAddressInfo `json:"addresses"`
}
