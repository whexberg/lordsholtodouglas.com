package templates

import (
	"fmt"
	"html/template"
	"io"
	"path/filepath"
)

// Renderer renders named templates with data.
type Renderer interface {
	Render(w io.Writer, name string, data any) error
}

// HTMLRenderer loads and renders HTML templates with a shared layout.
type HTMLRenderer struct {
	templates map[string]*template.Template
}

var funcMap = template.FuncMap{
	"add":        func(a, b int) int { return a + b },
	"subtract":   func(a, b int) int { return a - b },
	"divCents":   func(cents int) float64 { return float64(cents) / 100.0 },
	"divCents64": func(cents int64) float64 { return float64(cents) / 100.0 },
}

// NewRenderer parses layout + partials + each page template in templateDir.
// Page templates are all .html files directly in templateDir (not in subdirs).
func NewRenderer(templateDir string) (*HTMLRenderer, error) {
	// Shared files: layout + all partials
	shared, err := filepath.Glob(filepath.Join(templateDir, "partials", "*.html"))
	if err != nil {
		return nil, fmt.Errorf("glob partials: %w", err)
	}
	layoutFile := filepath.Join(templateDir, "layout.html")
	shared = append([]string{layoutFile}, shared...)

	// Each page template
	pages, err := filepath.Glob(filepath.Join(templateDir, "*.html"))
	if err != nil {
		return nil, fmt.Errorf("glob pages: %w", err)
	}

	r := &HTMLRenderer{templates: make(map[string]*template.Template)}

	for _, page := range pages {
		name := filepath.Base(page)
		if name == "layout.html" {
			continue
		}

		files := make([]string, len(shared)+1)
		copy(files, shared)
		files[len(shared)] = page

		tmpl, err := template.New("").Funcs(funcMap).ParseFiles(files...)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", name, err)
		}
		r.templates[name] = tmpl
	}

	return r, nil
}

// Render executes the named page template with the given data.
// The template must define a "content" block used by layout.html.
func (r *HTMLRenderer) Render(w io.Writer, name string, data any) error {
	tmpl, ok := r.templates[name]
	if !ok {
		return fmt.Errorf("template %q not found", name)
	}
	return tmpl.ExecuteTemplate(w, "layout.html", data)
}
