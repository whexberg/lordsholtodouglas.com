package content

import (
	"fmt"
	"html/template"
	"sort"
	"time"

	"gopkg.in/yaml.v3"
)

// Events lists all chapter events (non-draft, sorted by next occurrence).
var Events []Event

// Schedule describes when an event occurs.
type Schedule struct {
	Frequency string // "once", "monthly", "yearly"
	StartDate string // "YYYY-MM-DD"
	StartTime string // "HH:MM"
	EndDate   string // "YYYY-MM-DD" (for multi-day)
	EndTime   string // "HH:MM"
	Duration  int    // minutes
	Weekday   string // for monthly/yearly (e.g. "friday")
	Week      int    // which week (1-5) for monthly/yearly
	Month     int    // for yearly (1-12)
	Until     string // "YYYY-MM-DD" recurrence end
}

// Event represents a chapter event with schedule info.
type Event struct {
	Title          string
	Slug           string
	Description    string
	Location       string
	Date           string // formatted for display (computed from schedule)
	EndDate        string // formatted for display
	Time           string // formatted for display
	EventType      string
	MembersOnly    bool
	Draft          bool
	Featured       bool
	Weight         int
	DateTBDReason  string // "TBD" when date is undetermined
	Schedules      []Schedule
	ContentHTML    template.HTML // markdown body rendered to HTML
}

// GetEvent finds an event by slug.
func GetEvent(slug string) *Event {
	for i := range Events {
		if Events[i].Slug == slug {
			return &Events[i]
		}
	}
	return nil
}

// HasFutureOccurrence returns true if the event has a future occurrence.
func HasFutureOccurrence(e *Event) bool {
	now := time.Now()
	return !eventNextOccurrence(e, now, now.Location()).IsZero()
}

func isRecurring(e *Event) bool {
	for _, s := range e.Schedules {
		if s.Frequency == "monthly" || s.Frequency == "yearly" {
			return true
		}
	}
	return false
}

func eventNextOccurrence(e *Event, now time.Time, loc *time.Location) time.Time {
	var earliest time.Time
	for _, s := range e.Schedules {
		if t, ok := nextOccurrence(s, now); ok {
			if earliest.IsZero() || t.Before(earliest) {
				earliest = t
			}
		}
	}
	return earliest
}

type eventFrontmatter struct {
	Title       string `yaml:"title"`
	Slug        string `yaml:"slug"`
	Description string `yaml:"description"`
	Location    string `yaml:"location"`
	EventType   string `yaml:"eventType"`
	MembersOnly bool   `yaml:"isMembersOnly"`
	Draft       bool   `yaml:"draft"`
	Featured    bool   `yaml:"featured"`
	Weight      int    `yaml:"weight"`
	Meta        struct {
		DrawingDateTBD bool `yaml:"drawing_date_tbd"`
	} `yaml:"meta"`
	Schedules []struct {
		Frequency string `yaml:"frequency"`
		StartDate string `yaml:"startDate"`
		StartTime string `yaml:"startTime"`
		EndDate   string `yaml:"endDate"`
		EndTime   string `yaml:"endTime"`
		Duration  int    `yaml:"duration"`
		Weekday   string `yaml:"weekday"`
		Week      int    `yaml:"week"`
		Month     int    `yaml:"month"`
		Until     string `yaml:"until"`
	} `yaml:"schedules"`
}

func loadEvents(dir string, now time.Time, loc *time.Location) error {
	files, err := loadDir(dir, "events")
	if err != nil {
		return err
	}

	Events = nil
	for _, f := range files {
		var ef eventFrontmatter
		if err := yaml.Unmarshal(f.frontMatter, &ef); err != nil {
			return fmt.Errorf("%s: %w", f.name, err)
		}

		if ef.Draft {
			continue
		}

		slug := ef.Slug
		if slug == "" {
			slug = f.name
		}

		var schedules []Schedule
		for _, s := range ef.Schedules {
			schedules = append(schedules, Schedule{
				Frequency: s.Frequency,
				StartDate: s.StartDate,
				StartTime: s.StartTime,
				EndDate:   s.EndDate,
				EndTime:   s.EndTime,
				Duration:  s.Duration,
				Weekday:   s.Weekday,
				Week:      s.Week,
				Month:     s.Month,
				Until:     s.Until,
			})
		}

		var contentHTML template.HTML
		if len(f.body) > 0 {
			html, err := renderMarkdown(f.body)
			if err != nil {
				return fmt.Errorf("%s: markdown: %w", f.name, err)
			}
			contentHTML = html
		}

		var dateTBDReason string
		if ef.Meta.DrawingDateTBD {
			dateTBDReason = "TBD"
		}

		e := Event{
			Title:         ef.Title,
			Slug:          slug,
			Description:   ef.Description,
			Location:      ef.Location,
			EventType:     ef.EventType,
			MembersOnly:   ef.MembersOnly,
			Draft:         ef.Draft,
			Featured:      ef.Featured,
			Weight:        ef.Weight,
			DateTBDReason: dateTBDReason,
			Schedules:     schedules,
			ContentHTML:   contentHTML,
		}

		formatEventDates(&e, now)
		Events = append(Events, e)
	}

	sort.SliceStable(Events, func(i, j int) bool {
		iNext := eventNextOccurrence(&Events[i], now, loc)
		jNext := eventNextOccurrence(&Events[j], now, loc)
		if iNext.IsZero() && jNext.IsZero() {
			return false
		}
		if iNext.IsZero() {
			return false
		}
		if jNext.IsZero() {
			return true
		}
		return iNext.Before(jNext)
	})

	return nil
}
