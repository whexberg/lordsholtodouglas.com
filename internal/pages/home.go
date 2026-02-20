package pages

import (
	"log"
	"lsd3/internal/content"
	"net/http"
	"time"
)

type homeData struct {
	PageData
	SiteName       string
	Subtitle       string
	Sponsors       []content.Sponsor
	FeaturedEvents []eventView
	UpcomingEvents []eventView
}

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
		if e.Featured || content.HasFutureOccurrence(e) {
			eligible = append(eligible, *e)
		}
	}

	allViews := expandAndSort(eligible, now, to)

	var featured []eventView
	var upcoming []eventView
	for _, v := range allViews {
		if v.Featured {
			featured = append(featured, v)
		} else if len(upcoming) < 3 {
			upcoming = append(upcoming, v)
		}
	}

	data := homeData{
		PageData:       pageDataFromRequest(r),
		SiteName:       content.SiteConfig.Name,
		Subtitle:       content.SiteConfig.Subtitle,
		Sponsors:       content.Sponsors,
		FeaturedEvents: featured,
		UpcomingEvents: upcoming,
	}
	if err := h.renderer.Render(w, "home.html", data); err != nil {
		log.Printf("render home: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
