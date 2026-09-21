package resource

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// DateUTC truncates t to a UTC calendar date.
func DateUTC(t time.Time) time.Time {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func Overlaps(aStart, aEnd, bStart, bEnd time.Time) bool {
	aStart, aEnd = DateUTC(aStart), DateUTC(aEnd)
	bStart, bEnd = DateUTC(bStart), DateUTC(bEnd)
	return !aEnd.Before(bStart) && !aStart.After(bEnd)
}

func EachDay(from, to time.Time) []time.Time {
	from, to = DateUTC(from), DateUTC(to)
	if to.Before(from) {
		return nil
	}
	var days []time.Time
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		days = append(days, d)
	}
	return days
}

func InclusiveDays(from, to time.Time) int {
	from, to = DateUTC(from), DateUTC(to)
	if to.Before(from) {
		return 0
	}
	return int(to.Sub(from).Hours()/24) + 1
}

// DailyPercent sums allocation % per calendar day for one user.
func DailyPercent(allocs []Slice, from, to time.Time) map[time.Time]float64 {
	out := map[time.Time]float64{}
	for _, day := range EachDay(from, to) {
		var sum float64
		for _, a := range allocs {
			if Overlaps(a.Start, a.End, day, day) {
				sum += a.Percent
			}
		}
		out[day] = sum
	}
	return out
}

func MaxDailyPercent(daily map[time.Time]float64) float64 {
	var max float64
	for _, v := range daily {
		if v > max {
			max = v
		}
	}
	return max
}

func OverAllocated(daily map[time.Time]float64) bool {
	return MaxDailyPercent(daily) > 100
}

// Band maps utilization % to FR-CP-02 colors.
// GREEN: <80, YELLOW: 80–100, RED: >100.
func Band(util float64) string {
	if util < 80 {
		return BandGreen
	}
	if util <= 100 {
		return BandYellow
	}
	return BandRed
}

type Period struct {
	Key   string
	Start time.Time
	End   time.Time
}

func WeekStart(t time.Time) time.Time {
	t = DateUTC(t)
	// Monday = 1 ... Sunday = 0 in time.Weekday with Sunday=0
	wd := int(t.Weekday())
	if wd == 0 {
		wd = 7
	}
	return t.AddDate(0, 0, -(wd - 1))
}

func Periods(from, to time.Time, granularity string) []Period {
	from, to = DateUTC(from), DateUTC(to)
	if granularity == "month" {
		var out []Period
		cur := time.Date(from.Year(), from.Month(), 1, 0, 0, 0, 0, time.UTC)
		for !cur.After(to) {
			next := cur.AddDate(0, 1, 0)
			end := next.AddDate(0, 0, -1)
			if end.After(to) {
				end = to
			}
			start := cur
			if start.Before(from) {
				start = from
			}
			out = append(out, Period{
				Key:   cur.Format("2006-01"),
				Start: start,
				End:   end,
			})
			cur = next
		}
		return out
	}
	var out []Period
	cur := WeekStart(from)
	for !cur.After(to) {
		end := cur.AddDate(0, 0, 6)
		start := cur
		if start.Before(from) {
			start = from
		}
		if end.After(to) {
			end = to
		}
		y, w := cur.ISOWeek()
		out = append(out, Period{
			Key:   fmt.Sprintf("%d-W%02d", y, w),
			Start: start,
			End:   end,
		})
		cur = cur.AddDate(0, 0, 7)
	}
	return out
}

func HasOverloadInWindow(allocs []Slice, from, to time.Time) bool {
	return OverAllocated(DailyPercent(allocs, from, to))
}

func FilterUser(allocs []Slice, userID uuid.UUID) []Slice {
	out := make([]Slice, 0, len(allocs))
	for _, a := range allocs {
		if a.UserID == userID {
			out = append(out, a)
		}
	}
	return out
}
