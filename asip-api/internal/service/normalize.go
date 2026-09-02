package service

import (
	"fmt"
	"net/url"
	"strings"
)

func normalizeRequired(value, fieldName string) (string, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return "", fmt.Errorf("%s is required: %w", fieldName, ErrBadRequest)
	}
	return normalized, nil
}

func normalizeDomain(value string) (string, error) {
	normalized := strings.TrimSpace(value)
	normalized = strings.TrimPrefix(normalized, "/")
	normalized = strings.TrimSpace(normalized)
	if normalized == "" {
		return "", fmt.Errorf("domain is required: %w", ErrBadRequest)
	}

	candidate := normalized
	if strings.Contains(candidate, "://") ||
		strings.HasPrefix(strings.ToLower(candidate), "http:") ||
		strings.HasPrefix(strings.ToLower(candidate), "https:") {
		if !strings.Contains(candidate, "://") {
			candidate = strings.Replace(candidate, "http:", "http://", 1)
			candidate = strings.Replace(candidate, "https:", "https://", 1)
		}
		parsed, err := url.Parse(candidate)
		if err != nil || parsed.Hostname() == "" {
			return "", fmt.Errorf("invalid domain: %w", ErrBadRequest)
		}
		candidate = parsed.Hostname()
	} else if strings.Contains(candidate, "/") {
		parsed, err := url.Parse("http://" + candidate)
		if err != nil || parsed.Hostname() == "" {
			return "", fmt.Errorf("invalid domain: %w", ErrBadRequest)
		}
		candidate = parsed.Hostname()
	}

	candidate = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(candidate)), ".")
	if candidate == "" || strings.ContainsAny(candidate, " \t") {
		return "", fmt.Errorf("invalid domain: %w", ErrBadRequest)
	}
	return candidate, nil
}

func normalizeCountryCode(country string) (string, error) {
	normalized, err := normalizeRequired(country, "country")
	if err != nil {
		return "", err
	}

	code := strings.ToUpper(normalized)
	if len(code) != 2 || !isAlpha2(code) {
		return "", fmt.Errorf("country must be a 2-letter ISO 3166-1 alpha-2 code: %w", ErrBadRequest)
	}

	return code, nil
}

func isAlpha2(code string) bool {
	for _, r := range code {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}
