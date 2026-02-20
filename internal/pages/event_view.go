package pages

import (
	"lsd3/internal/content"
	"sort"
	"time"
)

// eventView is a template-friendly projection of an Event occurrence with a URL.
type eventView struct {
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
	Events []eventView
}

func occurrenceToView(occ content.EventOccurrence) eventView {
	return eventView{
		Title:       occ.Event.Title,
		Slug:        occ.Event.Slug,
		Description: occ.Event.Description,
		Location:    occ.Event.Location,
		Date:        content.OccurrenceDisplayDate(occ),
		Time:        content.OccurrenceDisplayTime(occ),
		EventType:   occ.Event.EventType,
		MembersOnly: occ.Event.MembersOnly,
		Featured:    occ.Event.Featured,
		DateTBD:     occ.Event.DateTBDReason != "",
		URL:         content.OccurrenceURL(occ),
		SortDate:    occ.Date,
	}
}

// expandAndSort returns sorted event views from occurrences within [from, to).
func expandAndSort(events []content.Event, from, to time.Time) []eventView {
	occs := content.ExpandOccurrences(events, from, to)
	views := make([]eventView, len(occs))
	for i, occ := range occs {
		views[i] = occurrenceToView(occ)
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

// groupByMonth groups event views into monthly groups.
func groupByMonth(views []eventView) []MonthGroup {
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
