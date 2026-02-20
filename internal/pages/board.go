package pages

import (
	"log"
	"lsd3/internal/content"
	"net/http"
)

type boardData struct {
	PageData
	Subtitle string
	Members  []content.BoardMember
}

func (h *PageHandler) BoardMembers(w http.ResponseWriter, r *http.Request) {
	data := boardData{
		PageData: pageDataFromRequest(r),
		Subtitle: content.BoardSubtitle,
		Members:  content.BoardMembers,
	}
	if err := h.renderer.Render(w, "board-members.html", data); err != nil {
		log.Printf("render board-members: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
