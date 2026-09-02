package dto

type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
