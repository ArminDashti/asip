package sync

import (
	"testing"
	"time"
)

func TestDurationUntilNextSync_NoPriorSync(t *testing.T) {
	now := time.Date(2026, 9, 4, 1, 0, 0, 0, time.UTC)
	wait := durationUntilNextSync(nil, 2, 5, now)
	want := time.Hour
	if wait != want {
		t.Fatalf("wait = %s, want %s", wait, want)
	}
}

func TestDurationUntilNextSync_RespectsFiveDayInterval(t *testing.T) {
	last := time.Date(2026, 9, 1, 2, 0, 0, 0, time.UTC)
	now := time.Date(2026, 9, 4, 1, 0, 0, 0, time.UTC)
	wait := durationUntilNextSync(&last, 2, 5, now)

	// Due at 2026-09-06 02:00 UTC → 2 days + 1 hour from now
	want := 2*24*time.Hour + time.Hour
	if wait != want {
		t.Fatalf("wait = %s, want %s", wait, want)
	}
}

func TestDurationUntilNextSync_OverdueUsesNextHour(t *testing.T) {
	last := time.Date(2026, 8, 20, 2, 0, 0, 0, time.UTC)
	now := time.Date(2026, 9, 4, 1, 0, 0, 0, time.UTC)
	wait := durationUntilNextSync(&last, 2, 5, now)
	want := time.Hour
	if wait != want {
		t.Fatalf("wait = %s, want %s", wait, want)
	}
}
