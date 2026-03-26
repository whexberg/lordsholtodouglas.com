package content

import (
	"fmt"
	"html/template"
	"sort"
	"time"

	"gopkg.in/yaml.v3"
)

// HistoryReports lists all history report articles, newest first.
var HistoryReports []HistoryReport

// HistoryReport represents a history report article.
type HistoryReport struct {
	Title       string
	Slug        string
	Author      string
	Date        string // human-readable display date (e.g. "Jan 2, 2006")
	RawDate     string // ISO date for sorting (e.g. "2006-01-02")
	ContentHTML template.HTML
}

// GetHistoryReport finds a report by slug.
func GetHistoryReport(slug string) *HistoryReport {
	for i := range HistoryReports {
		if HistoryReports[i].Slug == slug {
			return &HistoryReports[i]
		}
	}
	return nil
}

func loadHistoryReports(dir string) error {
	files, err := loadDir(dir, "history-reports")
	if err != nil {
		return err
	}

	HistoryReports = nil
	for _, f := range files {
		var rf struct {
			Title  string `yaml:"title"`
			Author string `yaml:"author"`
			Date   string `yaml:"date"`
		}
		if err := yaml.Unmarshal(f.frontMatter, &rf); err != nil {
			return fmt.Errorf("%s: %w", f.name, err)
		}

		var contentHTML template.HTML
		if len(f.body) > 0 {
			html, err := renderMarkdown(f.body)
			if err != nil {
				return fmt.Errorf("%s: markdown: %w", f.name, err)
			}
			contentHTML = html
		}

		displayDate := rf.Date
		if t, err := time.Parse("2006-01-02", rf.Date); err == nil {
			day := t.Day()
			suffix := "th"
			switch day {
			case 1, 21, 31:
				suffix = "st"
			case 2, 22:
				suffix = "nd"
			case 3, 23:
				suffix = "rd"
			}
			clampYear := t.Year() + 4005
			displayDate = fmt.Sprintf("%s %d%s, %d, Clamp Year %d",
				t.Format("January"), day, suffix, t.Year(), clampYear)
		}

		HistoryReports = append(HistoryReports, HistoryReport{
			Title:       rf.Title,
			Slug:        f.name,
			Author:      rf.Author,
			Date:        displayDate,
			RawDate:     rf.Date,
			ContentHTML: contentHTML,
		})
	}

	sort.Slice(HistoryReports, func(i, j int) bool {
		return HistoryReports[i].RawDate > HistoryReports[j].RawDate
	})

	return nil
}
