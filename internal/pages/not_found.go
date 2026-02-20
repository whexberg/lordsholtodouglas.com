package pages

import (
	"log"
	"net/http"
)

func (h *PageHandler) NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	data := pageDataFromRequest(r)
	if err := h.renderer.Render(w, "404.html", data); err != nil {
		log.Printf("render 404: %v", err)
		http.Error(w, "Not Found", http.StatusNotFound)
	}
}
