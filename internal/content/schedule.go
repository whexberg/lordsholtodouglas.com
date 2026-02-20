package content

import (
	"fmt"
	"strings"
	"time"
)

var weekdayMap = map[string]time.Weekday{
	"sunday":    time.Sunday,
	"monday":    time.Monday,
	"tuesday":   time.Tuesday,
	"wednesday": time.Wednesday,
	"thursday":  time.Thursday,
	"friday":    time.Friday,
	"saturday":  time.Saturday,
}

// nextOccurrence returns the next occurrence of a schedule after the given time.
// Returns zero time and false if no future occurrence exists.
func nextOccurrence(s Schedule, after time.Time) (time.Time, bool) {
	loc := after.Location()

	switch s.Frequency {
	case "once":
		t, err := time.ParseInLocation("2006-01-02", s.StartDate, loc)
		if err != nil {
			return time.Time{}, false
		}

		if s.StartTime != "" {
			t = setTime(t, s.StartTime)
		}

		if t.After(after) {
			return t, true
		}
		return time.Time{}, false

	case "monthly":
		wd, ok := weekdayMap[strings.ToLower(s.Weekday)]
		if !ok {
			return time.Time{}, false
		}
		until := parseUntil(s.Until, loc)

		// Start searching from the month of `after`
		y, m, _ := after.Date()
		for range 24 { // search up to 24 months ahead
			candidate := nthWeekdayOfMonth(y, m, s.Week, wd, loc)
			if s.StartTime != "" {
				candidate = setTime(candidate, s.StartTime)
			}

			if candidate.After(after) {
				if !until.IsZero() && candidate.After(until) {
					return time.Time{}, false
				}
				return candidate, true
			}

			m++
			if m > 12 {
				m = 1
				y++
			}
		}
		return time.Time{}, false

	case "yearly":
		wd, ok := weekdayMap[strings.ToLower(s.Weekday)]
		if !ok {
			return time.Time{}, false
		}

		until := parseUntil(s.Until, loc)
		targetMonth := time.Month(s.Month)

		y := after.Year()
		for range 10 { // search up to 10 years ahead
			candidate := nthWeekdayOfMonth(y, targetMonth, s.Week, wd, loc)
			if s.StartTime != "" {
				candidate = setTime(candidate, s.StartTime)
			}
			if candidate.After(after) {
				if !until.IsZero() && candidate.After(until) {
					return time.Time{}, false
				}
				return candidate, true
			}
			y++
		}
		return time.Time{}, false
	}

	return time.Time{}, false
}

// nthWeekdayOfMonth returns the nth occurrence of a weekday in the given month/year.
func nthWeekdayOfMonth(year int, month time.Month, week int, wd time.Weekday, loc *time.Location) time.Time {
	first := time.Date(year, month, 1, 0, 0, 0, 0, loc)

	// Find first occurrence of the weekday
	offset := int(wd) - int(first.Weekday())
	if offset < 0 { // candidate is after "until"
		offset += 7
	}

	day := 1 + offset + (week-1)*7
	return time.Date(year, month, day, 0, 0, 0, 0, loc)
}

func setTime(t time.Time, timeStr string) time.Time {
	var h, m int
	fmt.Sscanf(timeStr, "%d:%d", &h, &m)
	return time.Date(t.Year(), t.Month(), t.Day(), h, m, 0, 0, t.Location())
}

func parseUntil(s string, loc *time.Location) time.Time {
	if s == "" {
		return time.Time{}
	}

	t, err := time.ParseInLocation("2006-01-02", s, loc)
	if err != nil {
		return time.Time{}
	}

	// Set to end of day so events on the until date are included
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, loc)
}

// bestSchedule picks the currently-active schedule from a list.
// For events with multiple schedules (e.g. board meeting changed day/time),
// it returns the schedule whose date range contains `now`.
func bestSchedule(schedules []Schedule, now time.Time) *Schedule {
	loc := now.Location()

	// For single schedule, just return it
	if len(schedules) == 1 {
		return &schedules[0]
	}

	// For multiple schedules, find the one that applies now.
	// Schedules are ordered chronologically. Skip expired ones (until < now),
	// then pick the last one whose startDate <= now, or the first future one.
	var best *Schedule
	for i := range schedules {
		// Skip schedules that have expired
		if schedules[i].Until != "" {
			until := parseUntil(schedules[i].Until, loc)
			if !until.IsZero() && now.After(until) {
				continue
			}
		}
		sd, err := time.ParseInLocation("2006-01-02", schedules[i].StartDate, loc)
		if err != nil {
			continue
		}
		if !sd.After(now) {
			best = &schedules[i]
		} else if best == nil {
			// All remaining schedules are in the future; pick the first one
			best = &schedules[i]
			break
		}
	}
	if best == nil && len(schedules) > 0 {
		best = &schedules[len(schedules)-1]
	}
	return best
}

// formatEventDates populates the Date, EndDate, and Time display fields on an Event.
func formatEventDates(e *Event, now time.Time) {
	if len(e.Schedules) == 0 {
		return
	}

	s := bestSchedule(e.Schedules, now)
	if s == nil {
		return
	}

	loc := now.Location()

	switch s.Frequency {
	case "once":
		sd, err := time.ParseInLocation("2006-01-02", s.StartDate, loc)
		if err != nil {
			return
		}
		e.Date = sd.Format("January 2, 2006")
		if s.EndDate != "" {
			ed, err := time.ParseInLocation("2006-01-02", s.EndDate, loc)
			if err == nil {
				e.EndDate = ed.Format("January 2, 2006")
			}
		}
		if s.StartTime != "" {
			t := setTime(sd, s.StartTime)
			e.Time = formatTime(t)
		}

	case "monthly", "yearly":
		// Check all schedules to find the actual next occurrence
		next := eventNextOccurrence(e, now, loc)
		if !next.IsZero() {
			e.Date = next.Format("January 2, 2006")
			if s.StartTime != "" {
				// Use the time from the schedule that produces this occurrence
				for _, sched := range e.Schedules {
					if t, ok := nextOccurrence(sched, now); ok && t.Equal(next) {
						e.Time = formatTime(t)
						break
					}
				}
			}
		} else {
			// No future occurrence; show the pattern as fallback
			if s.Frequency == "monthly" {
				e.Date = fmt.Sprintf("%s %s monthly", ordinalWord(s.Week), capitalizeFirst(s.Weekday))
			} else {
				e.Date = fmt.Sprintf("%s %s of %s", ordinalWord(s.Week), capitalizeFirst(s.Weekday), time.Month(s.Month).String())
			}
			if s.StartTime != "" {
				t := setTime(now, s.StartTime)
				e.Time = formatTime(t)
			}
		}
	}
}

func formatTime(t time.Time) string {
	if t.Minute() == 0 {
		return t.Format("3 PM")
	}
	return t.Format("3:04 PM")
}

func ordinalWord(n int) string {
	switch n {
	case 1:
		return "1st"
	case 2:
		return "2nd"
	case 3:
		return "3rd"
	case 4:
		return "4th"
	case 5:
		return "5th"
	default:
		return fmt.Sprintf("%d", n)
	}
}

func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}

// EventOccurrence represents a single instance of an event on a specific date.
// Recurring events produce multiple occurrences; one-time events produce one.
type EventOccurrence struct {
	Event *Event
	Date  time.Time // the specific occurrence date/time
}

// ExpandOccurrences generates one EventOccurrence per instance of each event
// that falls within [from, to). Featured events with TBD dates get a single
// occurrence with a zero time.
func ExpandOccurrences(events []Event, from, to time.Time) []EventOccurrence {
	loc := from.Location()
	var occs []EventOccurrence

	for i := range events {
		e := &events[i]

		// Featured events with TBD dates: include once with zero time
		if e.Featured && e.DateTBDReason != "" {
			occs = append(occs, EventOccurrence{Event: e})
			continue
		}

		for _, s := range e.Schedules {
			switch s.Frequency {
			case "once":
				t, err := time.ParseInLocation("2006-01-02", s.StartDate, loc)
				if err != nil {
					continue
				}
				if s.StartTime != "" {
					t = setTime(t, s.StartTime)
				}
				if !t.Before(from) && t.Before(to) {
					occs = append(occs, EventOccurrence{Event: e, Date: t})
				}

			case "monthly":
				wd, ok := weekdayMap[strings.ToLower(s.Weekday)]
				if !ok {
					continue
				}
				until := parseUntil(s.Until, loc)
				cursor := from
				for cursor.Before(to) {
					y, m, _ := cursor.Date()
					candidate := nthWeekdayOfMonth(y, m, s.Week, wd, loc)
					if s.StartTime != "" {
						candidate = setTime(candidate, s.StartTime)
					}
					if !until.IsZero() && candidate.After(until) {
						break
					}
					if !candidate.Before(from) && candidate.Before(to) {
						occs = append(occs, EventOccurrence{Event: e, Date: candidate})
					}
					// Move to next month
					cursor = time.Date(y, m+1, 1, 0, 0, 0, 0, loc)
				}

			case "yearly":
				wd, ok := weekdayMap[strings.ToLower(s.Weekday)]
				if !ok {
					continue
				}
				until := parseUntil(s.Until, loc)
				targetMonth := time.Month(s.Month)
				for y := from.Year(); y <= to.Year(); y++ {
					candidate := nthWeekdayOfMonth(y, targetMonth, s.Week, wd, loc)
					if s.StartTime != "" {
						candidate = setTime(candidate, s.StartTime)
					}
					if !until.IsZero() && candidate.After(until) {
						break
					}
					if !candidate.Before(from) && candidate.Before(to) {
						occs = append(occs, EventOccurrence{Event: e, Date: candidate})
					}
				}
			}
		}
	}

	return occs
}

// OccurrenceURL returns the URL path for an event occurrence.
func OccurrenceURL(occ EventOccurrence) string {
	if occ.Date.IsZero() {
		return "/events/" + occ.Event.Slug
	}
	if !isRecurring(occ.Event) {
		return "/events/" + occ.Event.Slug
	}
	return "/events/" + occ.Event.Slug + "/" + occ.Date.Format("2006-01-02")
}

// OccurrenceDisplayDate returns the formatted display date for an occurrence.
func OccurrenceDisplayDate(occ EventOccurrence) string {
	if occ.Event.DateTBDReason != "" {
		return occ.Event.DateTBDReason
	}
	if occ.Date.IsZero() {
		return ""
	}
	return occ.Date.Format("January 2, 2006")
}

// OccurrenceDisplayTime returns the formatted display time for an occurrence.
func OccurrenceDisplayTime(occ EventOccurrence) string {
	if occ.Date.IsZero() {
		return ""
	}
	s := bestSchedule(occ.Event.Schedules, occ.Date)
	if s == nil || s.StartTime == "" {
		return ""
	}
	t := setTime(occ.Date, s.StartTime)
	return formatTime(t)
}

// EventInstance represents a specific occurrence of an event with resolved dates.
type EventInstance struct {
	Event       *Event
	Date        string // formatted display date (e.g. "March 20, 2026")
	EndDate     string // formatted display end date for multi-day events
	Time        string // formatted display time
	EndTime     string // formatted display end time
	InstanceKey string // YYYY-MM-DD key for URL routing
}

// EventURL returns the URL path for an event. Recurring events include
// the next instance date as a path segment.
func EventURL(e *Event, now time.Time) string {
	if !isRecurring(e) {
		return "/events/" + e.Slug
	}
	// For recurring events, find the next occurrence and include it in the URL
	loc := now.Location()
	next := eventNextOccurrence(e, now, loc)
	if next.IsZero() {
		return "/events/" + e.Slug
	}
	return "/events/" + e.Slug + "/" + next.Format("2006-01-02")
}

// ResolveInstance computes the display dates for an event at a specific instance date.
// For one-time events, dateStr is ignored. For recurring events, dateStr is "YYYY-MM-DD".
func ResolveInstance(e *Event, dateStr string) *EventInstance {
	loc := time.Now().Location()
	now := time.Now()

	inst := &EventInstance{Event: e}

	if len(e.Schedules) == 0 {
		inst.Date = e.Date
		inst.EndDate = e.EndDate
		inst.Time = e.Time
		return inst
	}

	s := bestSchedule(e.Schedules, now)
	if s == nil {
		s = &e.Schedules[0]
	}

	if s.Frequency == "once" || dateStr == "" {
		// Use the schedule's own dates
		inst.Date = e.Date
		inst.EndDate = e.EndDate
		inst.Time = e.Time
		if s.StartDate != "" {
			inst.InstanceKey = s.StartDate
		}
		return inst
	}

	// Recurring event with a specific instance date
	instanceDate, err := time.ParseInLocation("2006-01-02", dateStr, loc)
	if err != nil {
		inst.Date = e.Date
		inst.Time = e.Time
		return inst
	}

	inst.InstanceKey = dateStr
	inst.Date = instanceDate.Format("January 2, 2006")

	// Compute end date for multi-day events
	if s.EndDate != "" && s.StartDate != "" {
		startBase, err1 := time.ParseInLocation("2006-01-02", s.StartDate, loc)
		endBase, err2 := time.ParseInLocation("2006-01-02", s.EndDate, loc)
		if err1 == nil && err2 == nil {
			dayOffset := int(endBase.Sub(startBase).Hours() / 24)
			instanceEnd := instanceDate.AddDate(0, 0, dayOffset)
			inst.EndDate = instanceEnd.Format("January 2, 2006")
		}
	}

	// Time from the schedule
	if s.StartTime != "" {
		t := setTime(instanceDate, s.StartTime)
		inst.Time = formatTime(t)

		if s.Duration > 0 {
			endT := t.Add(time.Duration(s.Duration) * time.Minute)
			inst.EndTime = formatTime(endT)
		}
		if s.EndTime != "" {
			endT := setTime(instanceDate, s.EndTime)
			inst.EndTime = formatTime(endT)
		}
	}

	return inst
}
