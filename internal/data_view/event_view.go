package data_view

import (
	"lsd3/internal/content"
	"sort"
	"time"
)

// EventView is a template-friendly projection of an Event with a URL.
type EventView struct {
	Title       string
	Slug        string
	Description string
	Location    string
	Date        string
	Time        string
	EventType   string
	MembersOnly bool
	Featured    bool
	DateTBD     bool
	URL         string
	SortDate    time.Time // for sorting, not displayed
}

// MonthGroup groups event views under a month header.
type MonthGroup struct {
	Label  string // e.g. "February 2026"
	Events []EventView
}

func eventToView(e *content.Event) EventView {
	return EventView{
		Title:       e.Title,
		Slug:        e.Slug,
		Description: e.Description,
		Location:    e.Location,
		Date:        e.DisplayDate(),
		Time:        e.DisplayTime(),
		EventType:   e.EventType,
		MembersOnly: e.MembersOnly,
		Featured:    e.Featured,
		DateTBD:     e.DateTBD,
		URL:         e.URL(),
		SortDate:    e.Date,
	}
}

// ExpandAndSort returns sorted event views for events within [from, to).
func ExpandAndSort(events []content.Event, from, to time.Time) []EventView {
	var views []EventView
	for i := range events {
		e := &events[i]
		if e.DateTBD || e.Featured || (!e.Date.Before(from) && e.Date.Before(to)) {
			views = append(views, eventToView(e))
		}
	}
	sort.SliceStable(views, func(i, j int) bool {
		// TBD dates sort last
		if views[i].SortDate.IsZero() && !views[j].SortDate.IsZero() {
			return false
		}
		if !views[i].SortDate.IsZero() && views[j].SortDate.IsZero() {
			return true
		}
		return views[i].SortDate.Before(views[j].SortDate)
	})
	return views
}

// GroupByMonth groups event views into monthly groups.
func GroupByMonth(views []EventView) []MonthGroup {
	var groups []MonthGroup
	var current *MonthGroup

	for _, v := range views {
		if v.SortDate.IsZero() {
			continue // TBD events go in featured, not month groups
		}
		label := v.SortDate.Format("January 2006")
		if current == nil || current.Label != label {
			groups = append(groups, MonthGroup{Label: label})
			current = &groups[len(groups)-1]
		}
		current.Events = append(current.Events, v)
	}

	return groups
}
