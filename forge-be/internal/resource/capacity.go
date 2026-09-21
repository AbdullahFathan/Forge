package resource

import "time"

// Working-day and utilization math (FR-CP-01).
//
// Weekday = Monday–Friday (holidays still count as weekdays for allocated hours).
// Effective working day = weekday that is not a holiday.
//
// allocatedHours (period) = sum over weekdays of (dailyAllocationPercent/100 * capacityHoursPerDay)
// effectiveHours (period) = capacityHoursPerDay * count(effective working days)
// utilizationPercent     = allocatedHours / effectiveHours * 100
// If capacityHoursPerDay is 0 or effectiveHours is 0, utilization is 0.

func IsWeekend(t time.Time) bool {
	w := DateUTC(t).Weekday()
	return w == time.Saturday || w == time.Sunday
}

func IsWeekday(t time.Time) bool {
	return !IsWeekend(t)
}

func HolidaySet(dates []time.Time) map[time.Time]struct{} {
	out := make(map[time.Time]struct{}, len(dates))
	for _, d := range dates {
		out[DateUTC(d)] = struct{}{}
	}
	return out
}

func IsHoliday(t time.Time, holidays map[time.Time]struct{}) bool {
	_, ok := holidays[DateUTC(t)]
	return ok
}

func IsEffectiveWorkingDay(t time.Time, holidays map[time.Time]struct{}) bool {
	return IsWeekday(t) && !IsHoliday(t, holidays)
}

func CountEffectiveDays(from, to time.Time, holidays map[time.Time]struct{}) int {
	n := 0
	for _, d := range EachDay(from, to) {
		if IsEffectiveWorkingDay(d, holidays) {
			n++
		}
	}
	return n
}

func EffectiveHours(from, to time.Time, hoursPerDay int, holidays map[time.Time]struct{}) float64 {
	if hoursPerDay <= 0 {
		return 0
	}
	return float64(hoursPerDay) * float64(CountEffectiveDays(from, to, holidays))
}

func AllocatedHours(dailyPercent map[time.Time]float64, hoursPerDay int, holidays map[time.Time]struct{}) float64 {
	if hoursPerDay <= 0 {
		return 0
	}
	var h float64
	for day, pct := range dailyPercent {
		if !IsWeekday(day) {
			continue
		}
		_ = holidays // holidays still count toward allocated hours (denominator shrinks → util can rise)
		h += pct / 100 * float64(hoursPerDay)
	}
	return h
}

func UtilizationPercent(allocatedHours, effectiveHours float64) float64 {
	if effectiveHours <= 0 {
		return 0
	}
	return allocatedHours / effectiveHours * 100
}
