package pages

import (
	"lsd3/internal/middleware"
	"lsd3/internal/templates"
	"net/http"
)

// PageHandler serves content pages.
type PageHandler struct {
	renderer templates.Renderer
}

// NewPageHandler creates a new page handler.
func NewPageHandler(r templates.Renderer) *PageHandler {
	return &PageHandler{renderer: r}
}

// PageData holds common data passed to every page template.
type PageData struct {
	CartCount int
}

// pageDataFromRequest builds PageData from the request context.
func pageDataFromRequest(r *http.Request) PageData {
	count, _ := r.Context().Value(middleware.CartCountKey).(int)
	return PageData{CartCount: count}
}
