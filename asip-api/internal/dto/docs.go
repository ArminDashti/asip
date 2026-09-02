package dto

type EndpointParameter struct {
	Name        string `json:"name"`
	In          string `json:"in"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

type EndpointDocument struct {
	Method      string              `json:"method"`
	Path        string              `json:"path"`
	Summary     string              `json:"summary"`
	Description string              `json:"description"`
	Auth        bool                `json:"auth"`
	Parameters  []EndpointParameter `json:"parameters,omitempty"`
}

type DocsResponse struct {
	Service   string             `json:"service"`
	Version   string             `json:"version"`
	BasePath  string             `json:"base_path"`
	Endpoints []EndpointDocument `json:"endpoints"`
}
