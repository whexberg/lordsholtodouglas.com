package pages

import (
	"log"
	"lsd3/internal/content"
	"lsd3/internal/data_view"
	"lsd3/templates"
	"net/http"
)

func (h *PageHandler) Humbuggery(w http.ResponseWriter, r *http.Request) {
	data := data_view.HumbuggeryData{
		PageData: data_view.PageDataFromRequest(r),
		Intro:    content.HumbuggeryIntro,
		Images:   content.HumbuggeryImages,
	}
	if err := templates.HumbuggeryPage(data).Render(r.Context(), w); err != nil {
		log.Printf("render humbuggery: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
