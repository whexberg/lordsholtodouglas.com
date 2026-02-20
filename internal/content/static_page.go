package content

import (
	"html/template"

	"gopkg.in/yaml.v3"
)

// StaticPage represents a simple content page with a title, description, and body.
type StaticPage struct {
	Title       string
	Slug        string
	Description string
	ContentHTML template.HTML
}

// StaticPages maps slug to page (e.g. "cg-webhook/success" -> page).
var StaticPages map[string]StaticPage

func loadStaticPages(dir string) error {
	StaticPages = make(map[string]StaticPage)

	// Load cg-webhook pages
	files, err := loadDir(dir, "cg-webhook")
	if err != nil {
		return err
	}

	for _, f := range files {
		var fm struct {
			Title       string `yaml:"title"`
			Description string `yaml:"description"`
		}
		if err := yaml.Unmarshal(f.frontMatter, &fm); err != nil {
			return err
		}

		var contentHTML template.HTML
		if len(f.body) > 0 {
			html, err := renderMarkdown(f.body)
			if err != nil {
				return err
			}
			contentHTML = html
		}

		slug := "cg-webhook/" + f.name
		StaticPages[slug] = StaticPage{
			Title:       fm.Title,
			Slug:        slug,
			Description: fm.Description,
			ContentHTML: contentHTML,
		}
	}

	return nil
}
