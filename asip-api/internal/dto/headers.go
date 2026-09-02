package dto

type HttpHeaderEntry struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type HttpHeadersResponse struct {
	Ip      string            `json:"ip"`
	Headers []HttpHeaderEntry `json:"headers"`
}
