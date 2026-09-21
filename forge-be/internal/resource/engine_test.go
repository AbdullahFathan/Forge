package resource

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func d(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return t
}

func TestDailyPercentOverAllocation(t *testing.T) {
	u := uuid.New()
	p1, p2, p3 := uuid.New(), uuid.New(), uuid.New()
	from, to := d("2026-03-02"), d("2026-03-06") // Mon-Fri
	allocs := []Slice{
		{UserID: u, ProjectID: p1, Percent: 50, Start: from, End: to},
		{UserID: u, ProjectID: p2, Percent: 50, Start: from, End: to},
		{UserID: u, ProjectID: p3, Percent: 10, Start: from, End: to},
	}
	daily := DailyPercent(allocs, from, to)
	require.True(t, OverAllocated(daily))
	require.InDelta(t, 110, MaxDailyPercent(daily), 0.01)
}

func TestOverlappingSameProjectDates(t *testing.T) {
	require.True(t, Overlaps(d("2026-01-01"), d("2026-01-10"), d("2026-01-10"), d("2026-01-20")))
	require.False(t, Overlaps(d("2026-01-01"), d("2026-01-09"), d("2026-01-10"), d("2026-01-20")))
}

func TestBandBoundaries(t *testing.T) {
	require.Equal(t, BandGreen, Band(79.9))
	require.Equal(t, BandYellow, Band(80))
	require.Equal(t, BandYellow, Band(100))
	require.Equal(t, BandRed, Band(100.1))
}

func TestOverloadWindowEdges(t *testing.T) {
	u := uuid.New()
	p := uuid.New()
	now := d("2026-03-01")
	winTo := now.AddDate(0, 0, 13) // day 14 inclusive
	inside := Slice{UserID: u, ProjectID: p, Percent: 110, Start: now.AddDate(0, 0, 13), End: now.AddDate(0, 0, 13)}
	outside := Slice{UserID: u, ProjectID: p, Percent: 110, Start: now.AddDate(0, 0, 14), End: now.AddDate(0, 0, 14)}
	require.True(t, HasOverloadInWindow([]Slice{inside}, now, winTo))
	require.False(t, HasOverloadInWindow([]Slice{outside}, now, winTo))
}

func TestAllocationEndingSundayIgnoredForHours(t *testing.T) {
	u := uuid.New()
	p := uuid.New()
	// Fri 2026-03-06 to Sun 2026-03-08
	allocs := []Slice{{UserID: u, ProjectID: p, Percent: 100, Start: d("2026-03-06"), End: d("2026-03-08")}}
	daily := DailyPercent(allocs, d("2026-03-06"), d("2026-03-08"))
	hset := map[time.Time]struct{}{}
	hours := AllocatedHours(daily, 8, hset)
	require.InDelta(t, 8, hours, 0.01) // only Friday is a weekday
}
