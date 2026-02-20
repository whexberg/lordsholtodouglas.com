package pages

import (
	"log"
	"lsd3/internal/content"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type historyReportsData struct {
	PageData
	Reports []content.HistoryReport
}

func (h *PageHandler) HistoryReports(w http.ResponseWriter, r *http.Request) {
	data := historyReportsData{
		PageData: pageDataFromRequest(r),
		Reports:  content.HistoryReports,
	}
	if err := h.renderer.Render(w, "history-reports.html", data); err != nil {
		log.Printf("render history-reports: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

type historyReportData struct {
	PageData
	Report *content.HistoryReport
}

func (h *PageHandler) HistoryReport(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	report := content.GetHistoryReport(slug)
	if report == nil {
		h.NotFound(w, r)
		return
	}
	data := historyReportData{
		PageData: pageDataFromRequest(r),
		Report:   report,
	}
	if err := h.renderer.Render(w, "history-report.html", data); err != nil {
		log.Printf("render history-report: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
