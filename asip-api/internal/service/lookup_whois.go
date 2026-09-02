package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/ArminDashti/as-ip/server/internal/dto"
)

const defaultRdapBaseURL = "https://rdap.org/ip/"

type rdapHTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type rdapNetworkResponse struct {
	Handle       string `json:"handle"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	Country      string `json:"country"`
	StartAddress string `json:"startAddress"`
	EndAddress   string `json:"endAddress"`
	Status       []string
	Entities     []rdapEntity
	Events       []rdapEvent
	Remarks      []rdapRemark
	Links        []rdapLink
}

type rdapEntity struct {
	Handle       string        `json:"handle"`
	Roles        []string      `json:"roles"`
	VCardArray   []any         `json:"vcardArray"`
	Entities     []rdapEntity  `json:"entities"`
}

type rdapEvent struct {
	EventAction string `json:"eventAction"`
	EventDate   string `json:"eventDate"`
}

type rdapRemark struct {
	Description []string `json:"description"`
}

type rdapLink struct {
	Rel  string `json:"rel"`
	Href string `json:"href"`
}

func (s *LookupService) WithHTTPClient(client rdapHTTPClient) *LookupService {
	s.httpClient = client
	return s
}

func (s *LookupService) GetIPWhois(ctx context.Context, ip string) (dto.WhoisResponse, error) {
	normalizedIP := strings.TrimSpace(ip)
	if normalizedIP == "" {
		return dto.WhoisResponse{}, fmt.Errorf("ip is required: %w", ErrBadRequest)
	}
	parsed := net.ParseIP(normalizedIP)
	if parsed == nil {
		return dto.WhoisResponse{}, fmt.Errorf("invalid ip address: %w", ErrBadRequest)
	}
	normalizedIP = parsed.String()

	client := s.httpClient
	if client == nil {
		client = &http.Client{Timeout: 8 * time.Second}
	}

	requestURL := defaultRdapBaseURL + normalizedIP
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return dto.WhoisResponse{}, err
	}
	req.Header.Set("Accept", "application/rdap+json, application/json")

	resp, err := client.Do(req)
	if err != nil {
		return dto.WhoisResponse{}, fmt.Errorf("rdap request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return dto.WhoisResponse{}, fmt.Errorf("rdap response read failed: %w", err)
	}

	switch {
	case resp.StatusCode == http.StatusNotFound:
		return dto.WhoisResponse{}, fmt.Errorf("no whois/rdap record found for %s: %w", normalizedIP, ErrNotFound)
	case resp.StatusCode == http.StatusBadRequest:
		return dto.WhoisResponse{}, fmt.Errorf("invalid ip address for rdap: %w", ErrBadRequest)
	case resp.StatusCode < 200 || resp.StatusCode >= 300:
		return dto.WhoisResponse{}, fmt.Errorf("rdap upstream returned HTTP %d", resp.StatusCode)
	}

	var raw rdapNetworkResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return dto.WhoisResponse{}, fmt.Errorf("rdap response parse failed: %w", err)
	}

	return mapRdapToWhois(normalizedIP, requestURL, raw), nil
}

func mapRdapToWhois(ip, requestURL string, raw rdapNetworkResponse) dto.WhoisResponse {
	status := raw.Status
	if status == nil {
		status = []string{}
	}

	entities := collectEntityNames(raw.Entities)
	if entities == nil {
		entities = []string{}
	}

	events := make([]dto.WhoisEvent, 0, len(raw.Events))
	for _, event := range raw.Events {
		events = append(events, dto.WhoisEvent{
			Action: event.EventAction,
			Date:   event.EventDate,
		})
	}

	remarks := make([]string, 0)
	for _, remark := range raw.Remarks {
		for _, line := range remark.Description {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" {
				remarks = append(remarks, trimmed)
			}
		}
	}

	rdapURL := requestURL
	for _, link := range raw.Links {
		if strings.EqualFold(link.Rel, "self") && strings.TrimSpace(link.Href) != "" {
			rdapURL = link.Href
			break
		}
	}

	return dto.WhoisResponse{
		Ip:           ip,
		Handle:       raw.Handle,
		Name:         raw.Name,
		Type:         raw.Type,
		Country:      raw.Country,
		StartAddress: raw.StartAddress,
		EndAddress:   raw.EndAddress,
		Status:       status,
		Entities:     entities,
		Events:       events,
		Remarks:      remarks,
		RdapURL:      rdapURL,
	}
}

func collectEntityNames(entities []rdapEntity) []string {
	names := make([]string, 0)
	seen := map[string]struct{}{}

	var walk func([]rdapEntity)
	walk = func(list []rdapEntity) {
		for _, entity := range list {
			for _, name := range entityDisplayNames(entity) {
				if _, ok := seen[name]; ok {
					continue
				}
				seen[name] = struct{}{}
				names = append(names, name)
			}
			if len(entity.Entities) > 0 {
				walk(entity.Entities)
			}
		}
	}
	walk(entities)
	return names
}

func entityDisplayNames(entity rdapEntity) []string {
	names := make([]string, 0, 2)
	if fn := vcardFormattedName(entity.VCardArray); fn != "" {
		names = append(names, fn)
	}
	handle := strings.TrimSpace(entity.Handle)
	if handle != "" {
		role := ""
		if len(entity.Roles) > 0 {
			role = strings.TrimSpace(entity.Roles[0])
		}
		if role != "" {
			names = append(names, handle+" ("+role+")")
		} else if len(names) == 0 {
			names = append(names, handle)
		}
	}
	return names
}

func vcardFormattedName(vcardArray []any) string {
	if len(vcardArray) < 2 {
		return ""
	}
	properties, ok := vcardArray[1].([]any)
	if !ok {
		return ""
	}
	for _, property := range properties {
		parts, ok := property.([]any)
		if !ok || len(parts) < 4 {
			continue
		}
		name, _ := parts[0].(string)
		if !strings.EqualFold(name, "fn") {
			continue
		}
		value, _ := parts[3].(string)
		return strings.TrimSpace(value)
	}
	return ""
}
