package pages

import (
	"log"
	"lsd3/internal/content"
	"lsd3/internal/data_view"
	"lsd3/templates"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type boardData struct {
	data_view.PageData
	Subtitle string
	Members  []content.BoardMember
}

func (h *PageHandler) BoardMember(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	member := content.GetBoardMember(slug)
	if member == nil {
		h.NotFound(w, r)
		return
	}

	data := data_view.BoardMemberDetailData{
		PageData: data_view.PageDataFromRequest(r),
		Member:   member,
	}
	if err := templates.BoardMemberPage(data).Render(r.Context(), w); err != nil {
		log.Printf("render board-member: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
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
