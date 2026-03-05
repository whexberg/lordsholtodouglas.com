package data_view

import (
	"net/http"

	"lsd3/internal/content"
	"lsd3/internal/middleware"
)

// PageData holds common data passed to every page template.
type PageData struct {
	Title       string
	Subtitle    string
	SiteName    string
	SiteURL     string
	CartCount   int
	HeadScripts []string
}

type HomeData struct {
	PageData
	Sponsors       []content.Sponsor
	FeaturedEvents []EventView
	UpcomingEvents []EventView
}

// PageDataFromRequest builds PageData from the request context.
func PageDataFromRequest(r *http.Request) PageData {
	count, _ := r.Context().Value(middleware.CartCountKey).(int)
	return PageData{CartCount: count}
}
