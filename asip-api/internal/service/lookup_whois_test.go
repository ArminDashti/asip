package service

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) Do(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func TestGetIPWhoisInvalidIP(t *testing.T) {
	t.Parallel()

	svc := NewLookupService(nil)
	_, err := svc.GetIPWhois(context.Background(), "not-an-ip")
	if !errors.Is(err, ErrBadRequest) {
		t.Fatalf("expected ErrBadRequest, got %v", err)
	}
}

func TestGetIPWhoisMapsRdapResponse(t *testing.T) {
	t.Parallel()

	payload := `{
		"handle": "NET-8-8-8-0-1",
		"name": "GOGL",
		"type": "ALLOCATION",
		"country": "US",
		"startAddress": "8.8.8.0",
		"endAddress": "8.8.8.255",
		"status": ["active"],
		"entities": [
			{
				"handle": "GOGL",
				"roles": ["registrant"],
				"vcardArray": ["vcard", [["fn", {}, "text", "Google LLC"]]]
			}
		],
		"events": [{"eventAction": "last changed", "eventDate": "2024-01-01T00:00:00Z"}],
		"remarks": [{"description": ["Public DNS"]}],
		"links": [{"rel": "self", "href": "https://rdap.example/ip/8.8.8.8"}]
	}`

	svc := NewLookupService(nil).WithHTTPClient(roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if !strings.HasSuffix(req.URL.Path, "/8.8.8.8") {
			t.Fatalf("unexpected path %q", req.URL.Path)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(payload)),
			Header:     make(http.Header),
		}, nil
	}))

	result, err := svc.GetIPWhois(context.Background(), "8.8.8.8")
	if err != nil {
		t.Fatalf("GetIPWhois: %v", err)
	}
	if result.Name != "GOGL" {
		t.Fatalf("name = %q, want GOGL", result.Name)
	}
	if result.Country != "US" {
		t.Fatalf("country = %q, want US", result.Country)
	}
	if len(result.Entities) == 0 || result.Entities[0] != "Google LLC" {
		t.Fatalf("entities = %#v", result.Entities)
	}
	if result.RdapURL != "https://rdap.example/ip/8.8.8.8" {
		t.Fatalf("rdapUrl = %q", result.RdapURL)
	}
	if len(result.Remarks) != 1 || result.Remarks[0] != "Public DNS" {
		t.Fatalf("remarks = %#v", result.Remarks)
	}
}

func TestGetIPWhoisNotFound(t *testing.T) {
	t.Parallel()

	svc := NewLookupService(nil).WithHTTPClient(roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(strings.NewReader(`{"errorCode":404}`)),
			Header:     make(http.Header),
		}, nil
	}))

	_, err := svc.GetIPWhois(context.Background(), "203.0.113.1")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
