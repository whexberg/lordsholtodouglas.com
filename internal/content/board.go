package content

import (
	"html/template"
	"sort"

	"gopkg.in/yaml.v3"
)

// BoardMember represents an officer or board member of the chapter.
type BoardMember struct {
	Slug   string
	Name   string
	Label  string
	Image  string
	Weight int
	Bio    template.HTML
}

// GetBoardMember returns the board member with the given slug, or nil.
func GetBoardMember(slug string) *BoardMember {
	for i := range BoardMembers {
		if BoardMembers[i].Slug == slug {
			return &BoardMembers[i]
		}
	}
	return nil
}

// BoardSubtitle is the subtitle for the board members page (e.g. "Clamp Year 6031").
var BoardSubtitle string

// BoardMembers lists all board members, sorted by weight.
var BoardMembers []BoardMember

func loadBoardMembers(dir string) error {
	// Load page-level metadata from _index.md
	fm, _, err := loadFile(dir, "board-members/_index.md")
	if err != nil {
		return err
	}

	var index struct {
		Subtitle string `yaml:"subtitle"`
	}
	if err := yaml.Unmarshal(fm, &index); err != nil {
		return err
	}
	BoardSubtitle = index.Subtitle

	// Load individual member files
	files, err := loadDir(dir, "board-members")
	if err != nil {
		return err
	}

	BoardMembers = make([]BoardMember, 0, len(files))
	for _, f := range files {
		var m struct {
			Name   string `yaml:"name"`
			Label  string `yaml:"label"`
			Image  string `yaml:"image"`
			Weight int    `yaml:"weight"`
		}
		if err := yaml.Unmarshal(f.frontMatter, &m); err != nil {
			return err
		}

		var bio template.HTML
		if len(f.body) > 0 {
			bio, err = renderMarkdown(f.body)
			if err != nil {
				return err
			}
		}

		BoardMembers = append(BoardMembers, BoardMember{
			Slug:   f.name,
			Name:   m.Name,
			Label:  m.Label,
			Image:  m.Image,
			Weight: m.Weight,
			Bio:    bio,
		})
	}

	sort.Slice(BoardMembers, func(i, j int) bool {
		return BoardMembers[i].Weight < BoardMembers[j].Weight
	})

	return nil
}
