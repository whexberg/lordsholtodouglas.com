package content

import (
	"sort"

	"gopkg.in/yaml.v3"
)

// BoardMember represents an officer or board member of the chapter.
type BoardMember struct {
	Name   string
	Label  string
	Image  string
	Weight int
}

// BoardSubtitle is the subtitle for the board members page (e.g. "Clamp Year 6031").
var BoardSubtitle string

// BoardMembers lists all board members, sorted by weight.
var BoardMembers []BoardMember

func loadBoardMembers(dir string) error {
	fm, _, err := loadFile(dir, "board-members.md")
	if err != nil {
		return err
	}

	var page struct {
		Subtitle string `yaml:"subtitle"`
		Members  []struct {
			Name   string `yaml:"name"`
			Label  string `yaml:"label"`
			Image  string `yaml:"image"`
			Weight int    `yaml:"weight"`
		} `yaml:"members"`
	}
	if err := yaml.Unmarshal(fm, &page); err != nil {
		return err
	}

	BoardSubtitle = page.Subtitle
	BoardMembers = make([]BoardMember, len(page.Members))
	for i, m := range page.Members {
		BoardMembers[i] = BoardMember{
			Name:   m.Name,
			Label:  m.Label,
			Image:  m.Image,
			Weight: m.Weight,
		}
	}
	sort.Slice(BoardMembers, func(i, j int) bool {
		return BoardMembers[i].Weight < BoardMembers[j].Weight
	})

	return nil
}
