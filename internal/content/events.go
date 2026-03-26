package content

import (
	"fmt"
	"html/template"
	"sort"
	"time"

	"gopkg.in/yaml.v3"
)

// Events lists all chapter events (non-draft, sorted by date).
var Events []Event

// Event represents a chapter event instance with a concrete date.
type Event struct {
	Title       string
	Slug        string        // URL slug (e.g., "board-meeting", "spring-doins")
	Description string
	Location    string
	Date        time.Time     // parsed date+time (zero for TBD)
	EndDate     time.Time     // parsed end date (zero if not multi-day)
	Duration    int           // minutes
	EventType   string
	MembersOnly bool
	Draft       bool
	Featured    bool
	Weight      int
	DateTBD     bool          // true when date is undetermined
	IsInstance  bool          // true for files in subdirectories (slug/date URL)
	ContentHTML template.HTML // markdown body rendered to HTML
}

// GetEvent finds an event by slug and optional date string.
func GetEvent(slug string, dateStr string) *Event {
	for i := range Events {
		if Events[i].Slug != slug {
			continue
		}
		if dateStr == "" || Events[i].Date.Format("2006-01-02") == dateStr {
			return &Events[i]
		}
	}
	// Fallback: slug-only match for non-instance events
	if dateStr != "" {
		for i := range Events {
			if Events[i].Slug == slug && !Events[i].IsInstance {
				return &Events[i]
			}
		}
	}
	return nil
}

// IsFuture returns true if the event is in the future or has a TBD date.
func (e *Event) IsFuture(now time.Time) bool {
	return e.DateTBD || e.Date.After(now)
}

// URL returns the URL path for this event.
func (e *Event) URL() string {
	if e.IsInstance && !e.Date.IsZero() {
		return "/events/" + e.Slug + "/" + e.Date.Format("2006-01-02")
	}
	return "/events/" + e.Slug
}

// DisplayDate returns the formatted display date.
func (e *Event) DisplayDate() string {
	if e.DateTBD {
		return "TBD"
	}
	if e.Date.IsZero() {
		return ""
	}
	return e.Date.Format("January 2, 2006")
}

// DisplayEndDate returns the formatted display end date.
func (e *Event) DisplayEndDate() string {
	if e.EndDate.IsZero() {
		return ""
	}
	return e.EndDate.Format("January 2, 2006")
}

// DisplayTime returns the formatted display time.
func (e *Event) DisplayTime() string {
	if e.Date.IsZero() {
		return ""
	}
	if e.Date.Hour() == 0 && e.Date.Minute() == 0 {
		return ""
	}
	return formatTime(e.Date)
}

// DisplayEndTime returns the formatted display end time.
func (e *Event) DisplayEndTime() string {
	if e.Duration > 0 && !e.Date.IsZero() {
		endT := e.Date.Add(time.Duration(e.Duration) * time.Minute)
		return formatTime(endT)
	}
	if !e.EndDate.IsZero() && (e.EndDate.Hour() != 0 || e.EndDate.Minute() != 0) {
		return formatTime(e.EndDate)
	}
	return ""
}

func formatTime(t time.Time) string {
	if t.Minute() == 0 {
		return t.Format("3 PM")
	}
	return t.Format("3:04 PM")
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
	Date        string `yaml:"date"`      // "YYYY-MM-DD"
	StartTime   string `yaml:"startTime"` // "HH:MM"
	EndDate     string `yaml:"endDate"`   // "YYYY-MM-DD"
	EndTime     string `yaml:"endTime"`   // "HH:MM"
	Duration    int    `yaml:"duration"`  // minutes
	DateTBD     bool   `yaml:"dateTBD"`   // true when date is undetermined
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

		// Derive slug: explicit slug > parent directory > filename
		slug := ef.Slug
		if slug == "" && f.parentDir != "" {
			slug = f.parentDir
		}
		if slug == "" {
			slug = f.name
		}

		// Parse date
		var date time.Time
		if ef.Date != "" {
			t, err := time.ParseInLocation("2006-01-02", ef.Date, loc)
			if err != nil {
				return fmt.Errorf("%s: invalid date %q: %w", f.name, ef.Date, err)
			}
			if ef.StartTime != "" {
				t = setTime(t, ef.StartTime)
			}
			date = t
		}

		// Parse end date
		var endDate time.Time
		if ef.EndDate != "" {
			t, err := time.ParseInLocation("2006-01-02", ef.EndDate, loc)
			if err != nil {
				return fmt.Errorf("%s: invalid endDate %q: %w", f.name, ef.EndDate, err)
			}
			if ef.EndTime != "" {
				t = setTime(t, ef.EndTime)
			}
			endDate = t
		}

		var contentHTML template.HTML
		if len(f.body) > 0 {
			html, err := renderMarkdown(f.body)
			if err != nil {
				return fmt.Errorf("%s: markdown: %w", f.name, err)
			}
			contentHTML = html
		}

		e := Event{
			Title:       ef.Title,
			Slug:        slug,
			Description: ef.Description,
			Location:    ef.Location,
			Date:        date,
			EndDate:     endDate,
			Duration:    ef.Duration,
			EventType:   ef.EventType,
			MembersOnly: ef.MembersOnly,
			Draft:       ef.Draft,
			Featured:    ef.Featured,
			Weight:      ef.Weight,
			DateTBD:     ef.DateTBD,
			IsInstance:  f.parentDir != "",
			ContentHTML: contentHTML,
		}

		Events = append(Events, e)
	}

	sort.SliceStable(Events, func(i, j int) bool {
		iDate := Events[i].Date
		jDate := Events[j].Date
		// Zero dates (TBD) sort last
		if iDate.IsZero() && jDate.IsZero() {
			return false
		}
		if iDate.IsZero() {
			return false
		}
		if jDate.IsZero() {
			return true
		}
		return iDate.Before(jDate)
	})

	return nil
}

func setTime(t time.Time, timeStr string) time.Time {
	var h, m int
	fmt.Sscanf(timeStr, "%d:%d", &h, &m)
	return time.Date(t.Year(), t.Month(), t.Day(), h, m, 0, 0, t.Location())
}
