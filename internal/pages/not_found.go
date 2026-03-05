package pages

import (
	"log"
	"lsd3/templates"
	"net/http"
)

func (h *PageHandler) NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	if err := templates.NotFound().Render(r.Context(), w); err != nil {
		log.Printf("render 404: %v", err)
		http.Error(w, "Not Found", http.StatusNotFound)
	}
}
