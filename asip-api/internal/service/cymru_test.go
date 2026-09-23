package service

import (
	"strings"
	"testing"
	"time"

	"github.com/ArminDashti/as-ip/server/internal/model"
)

func TestCymruOriginName(t *testing.T) {
	t.Parallel()

	t.Run("ipv4 octets reversed", func(t *testing.T) {
		t.Parallel()
		got, err := cymruOriginName("172.68.205.217")
		if err != nil {
			t.Fatalf("cymruOriginName error: %v", err)
		}
		if got != "217.205.68.172.origin.asn.cymru.com" {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("ipv6 nibbles reversed", func(t *testing.T) {
		t.Parallel()
		// 2001:4860::1 ends with the byte 0x01, so the name must start with the
		// low nibble then the high nibble of that byte.
		const ip = "2001:4860::1"

		got, err := cymruOriginName(ip)
		if err != nil {
			t.Fatalf("cymruOriginName error: %v", err)
		}

		labels, ok := strings.CutSuffix(got, ".origin6.asn.cymru.com")
		if !ok {
			t.Fatalf("unexpected zone: %q", got)
		}

		parts := strings.Split(labels, ".")
		if len(parts) != 32 {
			t.Fatalf("expected 32 nibble labels, got %d (%q)", len(parts), got)
		}
		for i, part := range parts {
			if len(part) != 1 || !strings.Contains("0123456789abcdef", part) {
				t.Fatalf("label %d = %q is not a hex nibble", i, part)
			}
		}
		if parts[0] != "1" || parts[1] != "0" || parts[2] != "0" || parts[3] != "0" {
			t.Fatalf("leading nibbles = %v, want 1,0,0,0", parts[:4])
		}
	})

}

func TestCymruOriginNameRejectsGarbage(t *testing.T) {
	t.Parallel()

	if _, err := cymruOriginName("not-an-ip"); err == nil {
		t.Fatal("expected error for invalid ip")
	}
}

func TestParseCymruOriginRecords(t *testing.T) {
	t.Parallel()

	records := []string{"13335 | 172.68.205.0/24 | US | arin | 2015-02-25"}

	asn, countryCode, ok := parseCymruOriginRecords(records)
	if !ok {
		t.Fatal("expected the origin record to parse")
	}
	if asn != 13335 {
		t.Fatalf("asn = %d, want 13335", asn)
	}
	if countryCode != "US" {
		t.Fatalf("countryCode = %q, want US", countryCode)
	}
}

func TestParseCymruOriginRecordsMultipleOrigins(t *testing.T) {
	t.Parallel()

	asn, _, ok := parseCymruOriginRecords([]string{"13335 209242 | 1.1.1.0/24 | AU | apnic | 2011-08-11"})
	if !ok || asn != 13335 {
		t.Fatalf("asn = %d ok = %v, want 13335/true", asn, ok)
	}
}

func TestParseCymruOriginRecordsSkipsUnusableRows(t *testing.T) {
	t.Parallel()

	if _, _, ok := parseCymruOriginRecords([]string{"NA | NA | NA"}); ok {
		t.Fatal("expected NA rows to be skipped")
	}
	if _, _, ok := parseCymruOriginRecords(nil); ok {
		t.Fatal("expected empty records to be skipped")
	}
}

func TestParseCymruAsnRecords(t *testing.T) {
	t.Parallel()

	name, countryCode, ok := parseCymruAsnRecords([]string{
		"13335 | US | arin | 2010-07-14 | CLOUDFLARENET - Cloudflare, Inc., US",
	})
	if !ok {
		t.Fatal("expected the asn record to parse")
	}
	if name != "CLOUDFLARENET - Cloudflare, Inc." {
		t.Fatalf("name = %q", name)
	}
	if countryCode != "US" {
		t.Fatalf("countryCode = %q, want US", countryCode)
	}
}

func TestMergeIpAttributionKeepsKnownValues(t *testing.T) {
	t.Parallel()

	base := model.IpAttribution{
		Ip:      "8.8.8.8",
		Asn:     3356,
		AsName:  "Level 3 Parent LLC",
		Country: "United States",
		Source:  "dataset",
	}
	extra := model.IpAttribution{
		Asn:         13335,
		AsName:      "CLOUDFLARENET",
		CountryCode: "US",
		Country:     "United States",
		Source:      "cymru-dns",
		ResolvedAt:  time.Now().UTC(),
	}

	merged := mergeIpAttribution(base, extra)
	if merged.Asn != 3356 || merged.AsName != "Level 3 Parent LLC" {
		t.Fatalf("dataset values must win: %#v", merged)
	}
	if merged.CountryCode != "US" {
		t.Fatalf("missing country code should be filled: %#v", merged)
	}
	if merged.Source != "dataset" {
		t.Fatalf("source = %q, want dataset", merged.Source)
	}
}

func TestMergeIpAttributionFillsMissingValues(t *testing.T) {
	t.Parallel()

	resolvedAt := time.Now().UTC()
	merged := mergeIpAttribution(
		model.IpAttribution{Ip: "172.68.205.217"},
		model.IpAttribution{
			Asn:         13335,
			AsName:      "Cloudflare Inc.",
			CountryCode: "US",
			Country:     "United States",
			Source:      "cymru-dns",
			ResolvedAt:  resolvedAt,
		},
	)

	if !merged.IsComplete() {
		t.Fatalf("expected a complete attribution, got %#v", merged)
	}
	if merged.Source != "cymru-dns" {
		t.Fatalf("source = %q, want cymru-dns", merged.Source)
	}
	if !merged.ResolvedAt.Equal(resolvedAt) {
		t.Fatalf("resolvedAt = %v, want %v", merged.ResolvedAt, resolvedAt)
	}
}

func TestResolveIpAttributionWithoutRepository(t *testing.T) {
	t.Parallel()

	svc := &LookupService{}
	attribution := svc.resolveIpAttribution(t.Context(), "8.8.8.8", true)
	if attribution.Asn != 0 || attribution.AsName != "" || attribution.Country != "" {
		t.Fatalf("expected an empty attribution without a repository, got %#v", attribution)
	}
}
