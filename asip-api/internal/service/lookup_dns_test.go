package service

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"
)

func TestIsDNSRecordMiss(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil", err: nil, want: false},
		{name: "generic", err: errors.New("boom"), want: false},
		{name: "deadline", err: context.DeadlineExceeded, want: true},
		{name: "canceled", err: context.Canceled, want: true},
		{name: "not found", err: &net.DNSError{IsNotFound: true}, want: true},
		{name: "timeout", err: &net.DNSError{IsTimeout: true}, want: true},
		{name: "temporary", err: &net.DNSError{IsTemporary: true}, want: true},
		{name: "other dns", err: &net.DNSError{Err: "server misbehaving"}, want: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := isDNSRecordMiss(tt.err); got != tt.want {
				t.Fatalf("isDNSRecordMiss(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

type stubDNSResolver struct {
	lookupIP    func(ctx context.Context, network, host string) ([]net.IP, error)
	lookupNS    func(ctx context.Context, name string) ([]*net.NS, error)
	lookupMX    func(ctx context.Context, name string) ([]*net.MX, error)
	lookupTXT   func(ctx context.Context, name string) ([]string, error)
	lookupCNAME func(ctx context.Context, name string) (string, error)
}

func (s stubDNSResolver) LookupIP(ctx context.Context, network, host string) ([]net.IP, error) {
	if s.lookupIP != nil {
		return s.lookupIP(ctx, network, host)
	}
	return nil, &net.DNSError{IsNotFound: true}
}

func (s stubDNSResolver) LookupNS(ctx context.Context, name string) ([]*net.NS, error) {
	if s.lookupNS != nil {
		return s.lookupNS(ctx, name)
	}
	return nil, &net.DNSError{IsNotFound: true}
}

func (s stubDNSResolver) LookupMX(ctx context.Context, name string) ([]*net.MX, error) {
	if s.lookupMX != nil {
		return s.lookupMX(ctx, name)
	}
	return nil, &net.DNSError{IsNotFound: true}
}

func (s stubDNSResolver) LookupTXT(ctx context.Context, name string) ([]string, error) {
	if s.lookupTXT != nil {
		return s.lookupTXT(ctx, name)
	}
	return nil, &net.DNSError{IsNotFound: true}
}

func (s stubDNSResolver) LookupCNAME(ctx context.Context, name string) (string, error) {
	if s.lookupCNAME != nil {
		return s.lookupCNAME(ctx, name)
	}
	return "", &net.DNSError{IsNotFound: true}
}

func TestLookupDnsSoftFailsTimeoutOnTXT(t *testing.T) {
	t.Parallel()

	svc := &LookupService{
		resolver: stubDNSResolver{
			lookupIP: func(ctx context.Context, network, host string) ([]net.IP, error) {
				return []net.IP{net.ParseIP("140.82.121.3")}, nil
			},
			lookupTXT: func(ctx context.Context, name string) ([]string, error) {
				return nil, context.DeadlineExceeded
			},
		},
	}

	resp, err := svc.LookupDns(context.Background(), "github.com")
	if err != nil {
		t.Fatalf("expected soft-fail success, got err=%v", err)
	}
	if len(resp.A) != 1 || resp.A[0] != "140.82.121.3" {
		t.Fatalf("unexpected A records: %#v", resp.A)
	}
	if len(resp.TXT) != 0 {
		t.Fatalf("expected empty TXT after timeout, got %#v", resp.TXT)
	}
}

func TestLookupDnsHardErrorStillFails(t *testing.T) {
	t.Parallel()

	svc := &LookupService{
		resolver: stubDNSResolver{
			lookupIP: func(ctx context.Context, network, host string) ([]net.IP, error) {
				return nil, errors.New("resolver exploded")
			},
		},
	}

	_, err := svc.LookupDns(context.Background(), "github.com")
	if err == nil {
		t.Fatal("expected hard DNS error to fail the request")
	}
}

func TestLookupDnsPerQueryTimeoutDoesNotStarveLaterLookups(t *testing.T) {
	t.Parallel()

	var sawCNAME bool
	svc := &LookupService{
		resolver: stubDNSResolver{
			lookupIP: func(ctx context.Context, network, host string) ([]net.IP, error) {
				return []net.IP{net.ParseIP("93.184.216.34")}, nil
			},
			lookupTXT: func(ctx context.Context, name string) ([]string, error) {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(3 * time.Second):
					return []string{"too-slow"}, nil
				}
			},
			lookupCNAME: func(ctx context.Context, name string) (string, error) {
				sawCNAME = true
				return "cdn.example.net.", nil
			},
		},
	}

	start := time.Now()
	resp, err := svc.LookupDns(context.Background(), "example.com")
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !sawCNAME {
		t.Fatal("expected CNAME lookup to run after TXT timeout")
	}
	if resp.CNAME != "cdn.example.net." {
		t.Fatalf("cname = %q", resp.CNAME)
	}
	if elapsed > 4*time.Second {
		t.Fatalf("lookup took too long: %v", elapsed)
	}
}
