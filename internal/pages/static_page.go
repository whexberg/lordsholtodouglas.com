package pages

import (
	"log"
	"lsd3/internal/content"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type staticPageData struct {
	PageData
	Title       string
	Description string
	ContentHTML interface{}
}

func (h *PageHandler) StaticPage(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "*")
	page, ok := content.StaticPages[slug]
	if !ok {
		h.NotFound(w, r)
		return
	}

	data := staticPageData{
		PageData:    pageDataFromRequest(r),
		Title:       page.Title,
		Description: page.Description,
		ContentHTML: page.ContentHTML,
	}
	if err := h.renderer.Render(w, "static-page.html", data); err != nil {
		log.Printf("render static page %s: %v", slug, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
