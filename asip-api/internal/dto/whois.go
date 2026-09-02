package dto

type WhoisEvent struct {
	Action string `json:"action"`
	Date   string `json:"date"`
}

type WhoisResponse struct {
	Ip           string       `json:"ip"`
	Handle       string       `json:"handle"`
	Name         string       `json:"name"`
	Type         string       `json:"type"`
	Country      string       `json:"country"`
	StartAddress string       `json:"startAddress"`
	EndAddress   string       `json:"endAddress"`
	Status       []string     `json:"status"`
	Entities     []string     `json:"entities"`
	Events       []WhoisEvent `json:"events"`
	Remarks      []string     `json:"remarks"`
	RdapURL      string       `json:"rdapUrl"`
}
