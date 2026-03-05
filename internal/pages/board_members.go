package pages

import (
	"log"
	"lsd3/internal/content"
	"lsd3/internal/data_view"
	"lsd3/templates"
	"net/http"
)

type boardData struct {
	data_view.PageData
	Subtitle string
	Members  []content.BoardMember
}

func (h *PageHandler) BoardMembers(w http.ResponseWriter, r *http.Request) {
	data := data_view.BoardMemberPageData{
		PageData: data_view.PageDataFromRequest(r),
		Subtitle: content.BoardSubtitle,
		Members:  content.BoardMembers,
	}
	if err := templates.BoardMembers(data).Render(r.Context(), w); err != nil {
		log.Printf("render board-members: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
