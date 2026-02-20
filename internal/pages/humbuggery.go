package pages

import (
	"log"
	"lsd3/internal/content"
	"net/http"
)

type humbuggeryData struct {
	PageData
	Intro  string
	Images []string
}

func (h *PageHandler) Humbuggery(w http.ResponseWriter, r *http.Request) {
	data := humbuggeryData{
		PageData: pageDataFromRequest(r),
		Intro:    content.HumbuggeryIntro,
		Images:   content.HumbuggeryImages,
	}
	if err := h.renderer.Render(w, "humbuggery.html", data); err != nil {
		log.Printf("render humbuggery: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
