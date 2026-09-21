package resource

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestHolidayRaisesUtilization(t *testing.T) {
	from, to := d("2026-03-02"), d("2026-03-06") // Mon-Fri
	u := Slice{Percent: 100, Start: from, End: to}
	daily := DailyPercent([]Slice{u}, from, to)

	none := map[time.Time]struct{}{}
	utilNone := UtilizationPercent(AllocatedHours(daily, 8, none), EffectiveHours(from, to, 8, none))
	require.InDelta(t, 100, utilNone, 0.01)

	hol := HolidaySet([]time.Time{d("2026-03-04")}) // Wednesday
	allocH := AllocatedHours(daily, 8, hol)
	eff := EffectiveHours(from, to, 8, hol)
	utilHol := UtilizationPercent(allocH, eff)
	require.InDelta(t, 40, allocH, 0.01) // 5 weekdays still allocated
	require.InDelta(t, 32, eff, 0.01)    // 4 effective days
	require.Greater(t, utilHol, utilNone)
	require.InDelta(t, 125, utilHol, 0.01)
}

func TestZeroCapacityUtilization(t *testing.T) {
	from, to := d("2026-03-02"), d("2026-03-06")
	daily := DailyPercent([]Slice{{Percent: 100, Start: from, End: to}}, from, to)
	hset := map[time.Time]struct{}{}
	require.Equal(t, 0.0, AllocatedHours(daily, 0, hset))
	require.Equal(t, 0.0, EffectiveHours(from, to, 0, hset))
	require.Equal(t, 0.0, UtilizationPercent(0, 0))
}

func TestInclusiveDaysRange(t *testing.T) {
	require.Equal(t, 28, InclusiveDays(d("2026-03-02"), d("2026-03-29")))
	require.Equal(t, 84, InclusiveDays(d("2026-01-01"), d("2026-03-25")))
}
