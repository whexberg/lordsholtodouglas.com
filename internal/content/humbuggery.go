package content

import (
	"html/template"
	"sort"

	"gopkg.in/yaml.v3"
)

// Humbug represents a past Noble Grand Humbug.
type Humbug struct {
	Slug   string
	Name   string
	Image  string
	Years  string
	Weight int
	Bio    template.HTML
}

// HumbuggeryIntro is the intro text for the Hall of Humbuggery page.
var HumbuggeryIntro string

// Humbugs lists all past humbugs, sorted by weight.
var Humbugs []Humbug

// GetHumbug returns the humbug with the given slug, or nil.
func GetHumbug(slug string) *Humbug {
	for i := range Humbugs {
		if Humbugs[i].Slug == slug {
			return &Humbugs[i]
		}
	}
	return nil
}

func loadHumbuggery(dir string) error {
	fm, body, err := loadFile(dir, "humbuggery/_index.md")
	if err != nil {
		return err
	}

	// Preserve any intro from _index.md body, otherwise use default
	if len(body) > 0 {
		HumbuggeryIntro = string(body)
	} else {
		HumbuggeryIntro = `Here we honor the legendary men who have ascended to the highest echelons of our board, carrying the torch of our sacred traditions through the roaring tides of history into the present day. With a resolve as unyielding as granite and a refusal to give any sucker an even break, these intrepid volunteers have dedicated their hearts and minds to forging unforgettable functions, raising funds to aid widows, orphans, and families in need. Even in the shadow of the darkest days, they have risen like titans, tearing away the blindfold of despair to illuminate a path of hope and possibility. These are not merely board members—they are stewards of dreams, architects of change, and torchbearers of the indomitable spirit that defines us all.`
	}
	_ = fm

	// Load individual humbug files
	files, err := loadDir(dir, "humbuggery")
	if err != nil {
		return err
	}

	Humbugs = make([]Humbug, 0, len(files))
	for _, f := range files {
		var m struct {
			Name   string `yaml:"name"`
			Image  string `yaml:"image"`
			Years  string `yaml:"years"`
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

		Humbugs = append(Humbugs, Humbug{
			Slug:   f.name,
			Name:   m.Name,
			Image:  m.Image,
			Years:  m.Years,
			Weight: m.Weight,
			Bio:    bio,
		})
	}

	sort.Slice(Humbugs, func(i, j int) bool {
		return Humbugs[i].Weight < Humbugs[j].Weight
	})

	return nil
}
