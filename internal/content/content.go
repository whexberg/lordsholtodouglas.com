package content

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"
)

type FileContent struct {
	name        string // filename without extension
	parentDir   string // subdirectory name, empty for top-level files
	frontMatter []byte
	body        []byte
}

// Load parses all markdown content from contentDir and populates package-level vars.
func Load(contentDir string) error {
	now := time.Now()
	loc := now.Location()

	if err := loadSiteConfig(contentDir); err != nil {
		return fmt.Errorf("site config: %w", err)
	}

	if err := loadBoardMembers(contentDir); err != nil {
		return fmt.Errorf("board members: %w", err)
	}

	if err := loadSponsors(contentDir); err != nil {
		return fmt.Errorf("sponsors: %w", err)
	}

	if err := loadEvents(contentDir, now, loc); err != nil {
		return fmt.Errorf("events: %w", err)
	}

	if err := loadHistoryReports(contentDir); err != nil {
		return fmt.Errorf("history reports: %w", err)
	}

	if err := loadHumbuggery(contentDir); err != nil {
		return fmt.Errorf("humbuggery: %w", err)
	}

	if err := loadStaticPages(contentDir); err != nil {
		return fmt.Errorf("static pages: %w", err)
	}

	return nil
}

func loadFile(dir, filename string) ([]byte, []byte, error) {
	data, err := os.ReadFile(filepath.Join(dir, filename))
	if err != nil {
		return nil, nil, err
	}

	const delim = "---"
	s := string(data)

	// Must start with ---
	s = strings.TrimLeft(s, "\n\r ")
	if !strings.HasPrefix(s, delim) {
		return nil, data, nil
	}

	s = s[len(delim):]
	if fm, body, ok := strings.Cut(s, "\n"+delim); !ok {
		// Frontmatter only, no body
		return []byte(strings.TrimSpace(s)), nil, nil
	} else {
		// Strip leading newline from body
		return []byte(fm), []byte(strings.TrimLeft(body, "\n\r")), nil
	}
}

func loadDir(dir ...string) ([]FileContent, error) {
	dirpath := ""
	for _, d := range dir {
		dirpath = filepath.Join(dirpath, d)
	}

	entries, err := os.ReadDir(dirpath)
	if err != nil {
		return nil, err
	}

	files := make([]FileContent, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			// Recurse one level into subdirectories
			subEntries, err := os.ReadDir(filepath.Join(dirpath, entry.Name()))
			if err != nil {
				return nil, err
			}
			for _, sub := range subEntries {
				if sub.IsDir() || !strings.HasSuffix(sub.Name(), ".md") || sub.Name() == "_index.md" {
					continue
				}
				fm, body, err := loadFile(filepath.Join(dirpath, entry.Name()), sub.Name())
				if err != nil {
					return nil, err
				}
				files = append(files, FileContent{
					name:        strings.TrimSuffix(sub.Name(), ".md"),
					parentDir:   entry.Name(),
					frontMatter: fm,
					body:        body,
				})
			}
			continue
		}

		if !strings.HasSuffix(entry.Name(), ".md") || entry.Name() == "_index.md" {
			continue
		}

		fm, body, err := loadFile(dirpath, entry.Name())
		if err != nil {
			return nil, err
		}

		files = append(files, FileContent{
			name:        strings.TrimSuffix(entry.Name(), ".md"),
			frontMatter: fm,
			body:        body,
		})
	}

	return files, nil
}

// md is a goldmark instance with unsafe HTML rendering enabled,
// so inline HTML in markdown (links, divs, scripts) renders correctly.
var md = goldmark.New(
	goldmark.WithRendererOptions(html.WithUnsafe()),
)

// Hugo shortcode patterns
var (
	// {{< gallery class="..." >}} → <div class="...">
	reGalleryOpen  = regexp.MustCompile(`\{\{<\s*gallery\s+class="([^"]*)"[^>]*>\}\}`)
	reGalleryClose = regexp.MustCompile(`\{\{<\s*/gallery\s*>\}\}`)
	// {{< figure src="..." alt="..." imgclass="..." >}} → <img src="..." alt="..." class="...">
	reFigure = regexp.MustCompile(`\{\{<\s*figure\s+([^>]*?)>\}\}`)
	// Generic shortcode fallback — strip any remaining
	reShortcode = regexp.MustCompile(`\{\{<\s*/?[^>]*>\}\}`)

	reFigureAttr = regexp.MustCompile(`(\w+)="([^"]*)"`)
)

// expandShortcodes converts Hugo shortcodes to plain HTML.
func expandShortcodes(src []byte) []byte {
	s := string(src)

	s = reGalleryOpen.ReplaceAllString(s, `<div class="$1">`)
	s = reGalleryClose.ReplaceAllString(s, `</div>`)

	s = reFigure.ReplaceAllStringFunc(s, func(match string) string {
		attrs := reFigureAttr.FindAllStringSubmatch(match, -1)
		var srcVal, altVal, classVal string
		for _, a := range attrs {
			switch a[1] {
			case "src":
				srcVal = a[2]
			case "alt":
				altVal = a[2]
			case "imgclass":
				classVal = a[2]
			}
		}
		if classVal != "" {
			return fmt.Sprintf(`<img src="%s" alt="%s" class="%s">`, srcVal, altVal, classVal)
		}
		return fmt.Sprintf(`<img src="%s" alt="%s">`, srcVal, altVal)
	})

	// Strip any remaining shortcodes
	s = reShortcode.ReplaceAllString(s, "")

	return []byte(s)
}

// renderMarkdown converts markdown to HTML, expanding Hugo shortcodes first.
func renderMarkdown(src []byte) (template.HTML, error) {
	src = expandShortcodes(src)

	var buf bytes.Buffer
	if err := md.Convert(src, &buf); err != nil {
		return "", err
	}

	return template.HTML(buf.String()), nil
}
