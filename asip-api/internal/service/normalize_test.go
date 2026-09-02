package service

import (
	"errors"
	"testing"
)

func TestNormalizeCountryCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "uppercase iso code", input: "US", want: "US"},
		{name: "lowercase iso code", input: "de", want: "DE"},
		{name: "trimmed iso code", input: "  fr  ", want: "FR"},
		{name: "rejects country name", input: "United States", wantErr: true},
		{name: "rejects alpha-3 code", input: "USA", wantErr: true},
		{name: "rejects empty", input: "  ", wantErr: true},
		{name: "rejects numeric", input: "12", wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := normalizeCountryCode(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got code %q", got)
				}
				if !errors.Is(err, ErrBadRequest) {
					t.Fatalf("expected ErrBadRequest, got %v", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNormalizeDomain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "bare host", input: "example.com", want: "example.com"},
		{name: "http url", input: "http://example.com", want: "example.com"},
		{name: "https url", input: "https://example.com", want: "example.com"},
		{name: "https with path", input: "https://example.com/path?x=1", want: "example.com"},
		{name: "typo host", input: "exampl.com", want: "exampl.com"},
		{name: "trailing slash url", input: "https://example.com/", want: "example.com"},
		{name: "uppercase host", input: "Example.COM", want: "example.com"},
		{name: "fqdn trailing dot", input: "example.com.", want: "example.com"},
		{name: "leading slash from gin", input: "/https://example.com", want: "example.com"},
		{name: "empty", input: "  ", wantErr: true},
		{name: "whitespace only after slash", input: "/", wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := normalizeDomain(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %q", got)
				}
				if !errors.Is(err, ErrBadRequest) {
					t.Fatalf("expected ErrBadRequest, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
