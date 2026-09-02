package dto

type IpInfoResponse struct {
	Ip      string `json:"ip"`
	Asn     int    `json:"asn"`
	As      string `json:"as"`
	Country string `json:"country"`
}
