package content

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// Sponsors lists all chapter sponsors.
var Sponsors []Sponsor

// Sponsor represents a chapter sponsor.
type Sponsor struct {
	Name string
	Logo string
	Link string
}

func loadSponsors(dir string) error {
	files, err := loadDir(dir, "sponsors")
	if err != nil {
		return err
	}

	Sponsors = nil
	for _, f := range files {
		var s struct {
			Name string `yaml:"name"`
			Logo string `yaml:"logo"`
			Link string `yaml:"link"`
		}
		if err := yaml.Unmarshal(f.frontMatter, &s); err != nil {
			return fmt.Errorf("%s: %w", f.name, err)
		}
		Sponsors = append(Sponsors, Sponsor{
			Name: s.Name,
			Logo: s.Logo,
			Link: s.Link,
		})
	}
	return nil
}
