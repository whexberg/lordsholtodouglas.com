package pages

import (
	"log"
	"lsd3/internal/content"
	"lsd3/internal/data_view"
	"lsd3/templates"
	"net/http"
	"time"
)

// type homeData struct {
// 	data_view.PageData
// 	SiteName       string
// 	Subtitle       string
// 	Sponsors       []content.Sponsor
// 	FeaturedEvents []data_view.EventView
// 	UpcomingEvents []data_view.EventView
// }

func (h *PageHandler) Home(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	to := now.AddDate(1, 0, 0)

	// Filter to non-meeting events with future occurrences or featured
	var eligible []content.Event
	for i := range content.Events {
		e := &content.Events[i]
		if e.EventType == "meeting" {
			continue
		}
		if e.Featured || e.IsFuture(now) {
			eligible = append(eligible, *e)
		}
	}

	allViews := data_view.ExpandAndSort(eligible, now, to)

	var featured []data_view.EventView
	var upcoming []data_view.EventView
	for _, v := range allViews {
		if v.Featured {
			featured = append(featured, v)
		} else if len(upcoming) < 3 {
			upcoming = append(upcoming, v)
		}
	}

	pageData := data_view.PageDataFromRequest(r)
	pageData.Title = "Lord Sholto Douglas #3 — E Clampus Vitus"
	pageData.SiteName = content.SiteConfig.Name
	pageData.Subtitle = content.SiteConfig.Subtitle
	data := data_view.HomeData{
		PageData:       pageData,
		Sponsors:       content.Sponsors,
		FeaturedEvents: featured,
		UpcomingEvents: upcoming,
	}

	if err := templates.HomePage(data).Render(r.Context(), w); err != nil {
		log.Printf("render home: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
