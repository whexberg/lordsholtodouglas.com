package pages

import (
	"log"
	"net/http"

	"lsd3/internal/content"
	"lsd3/internal/data_view"
	"lsd3/internal/templates"

	"github.com/go-chi/chi/v5"
)

func (h *PageHandler) HistoryReports(w http.ResponseWriter, r *http.Request) {
	data := data_view.HistoryReportsData{
		PageData: data_view.PageDataFromRequest(r),
		Reports:  content.HistoryReports,
	}
	if err := templates.HistoryReportsPage(data).Render(r.Context(), w); err != nil {
		log.Printf("render history-reports: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *PageHandler) HistoryReport(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	report := content.GetHistoryReport(slug)
	if report == nil {
		h.NotFound(w, r)
		return
	}
	data := data_view.HistoryReportData{
		PageData: data_view.PageDataFromRequest(r),
		Report:   report,
	}
	if err := templates.HistoryReportPage(data).Render(r.Context(), w); err != nil {
		log.Printf("render history-report: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
