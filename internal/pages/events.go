package pages

import (
	"log"
	"lsd3/internal/content"
	"lsd3/internal/data_view"
	"lsd3/templates"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

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

	allViews := data_view.ExpandAndSort(eligible, now, to)

	// Separate featured from upcoming
	var featured []data_view.EventView
	var upcoming []data_view.EventView

	for _, v := range allViews {
		if v.Featured {
			featured = append(featured, v)
		} else {
			upcoming = append(upcoming, v)
		}
	}

	data := data_view.EventsPageData{
		PageData:    data_view.PageDataFromRequest(r),
		Featured:    featured,
		MonthGroups: data_view.GroupByMonth(upcoming),
	}
	if err := templates.EventsPage(data).Render(r.Context(), w); err != nil {
		log.Printf("render events: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
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

	data := data_view.EventDetailData{
		PageData: data_view.PageDataFromRequest(r),
		Instance: instance,
	}
	if err := templates.EventPage(data).Render(r.Context(), w); err != nil {
		log.Printf("render event: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
