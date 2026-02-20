package pages

import (
	"log"
	"lsd3/internal/content"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type eventsData struct {
	PageData
	Featured    []eventView
	MonthGroups []MonthGroup
}

func (h *PageHandler) Events(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	to := now.AddDate(1, 0, 0) // look 1 year ahead

	// Filter to non-draft events with future occurrences or featured
	var eligible []content.Event
	for i := range content.Events {
		e := &content.Events[i]
		if e.Featured || content.HasFutureOccurrence(e) {
			eligible = append(eligible, *e)
		}
	}

	allViews := expandAndSort(eligible, now, to)

	// Separate featured from upcoming
	var featured []eventView
	var upcoming []eventView
	for _, v := range allViews {
		if v.Featured {
			featured = append(featured, v)
		} else {
			upcoming = append(upcoming, v)
		}
	}

	data := eventsData{
		PageData:    pageDataFromRequest(r),
		Featured:    featured,
		MonthGroups: groupByMonth(upcoming),
	}
	if err := h.renderer.Render(w, "events.html", data); err != nil {
		log.Printf("render events: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

type eventData struct {
	PageData
	Instance *content.EventInstance
}

func (h *PageHandler) Event(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	event := content.GetEvent(slug)
	if event == nil {
		h.NotFound(w, r)
		return
	}

	dateStr := chi.URLParam(r, "date")
	instance := content.ResolveInstance(event, dateStr)

	data := eventData{
		PageData: pageDataFromRequest(r),
		Instance: instance,
	}
	if err := h.renderer.Render(w, "event.html", data); err != nil {
		log.Printf("render event: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
