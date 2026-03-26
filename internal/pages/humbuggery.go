package pages

import (
	"log"
	"lsd3/internal/content"
	"lsd3/internal/data_view"
	"lsd3/templates"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *PageHandler) Humbuggery(w http.ResponseWriter, r *http.Request) {
	data := data_view.HumbuggeryData{
		PageData: data_view.PageDataFromRequest(r),
		Intro:    content.HumbuggeryIntro,
		Humbugs:  content.Humbugs,
	}
	if err := templates.HumbuggeryPage(data).Render(r.Context(), w); err != nil {
		log.Printf("render humbuggery: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *PageHandler) Humbug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	humbug := content.GetHumbug(slug)
	if humbug == nil {
		h.NotFound(w, r)
		return
	}

	data := data_view.HumbugDetailData{
		PageData: data_view.PageDataFromRequest(r),
		Humbug:   humbug,
	}
	if err := templates.HumbugPage(data).Render(r.Context(), w); err != nil {
		log.Printf("render humbug: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
