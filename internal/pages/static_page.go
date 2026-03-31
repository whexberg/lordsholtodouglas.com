package pages

import (
	"log"
	"lsd3/internal/content"
	"lsd3/internal/data_view"
	"lsd3/internal/templates"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *PageHandler) StaticPage(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "*")
	page, ok := content.StaticPages[slug]
	if !ok {
		h.NotFound(w, r)
		return
	}

	pd := data_view.PageDataFromRequest(r)
	pd.Title = page.Title
	data := data_view.StaticPageData{
		PageData:    pd,
		Description: page.Description,
		ContentHTML: page.ContentHTML,
	}
	if err := templates.StaticPage(data).Render(r.Context(), w); err != nil {
		log.Printf("render static page %s: %v", slug, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
