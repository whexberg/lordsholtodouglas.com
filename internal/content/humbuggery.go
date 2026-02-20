package content

import (
	"gopkg.in/yaml.v3"
)

// HumbuggeryIntro is the intro text for the Hall of Humbuggery page.
var HumbuggeryIntro string

// HumbuggeryImages lists all humbuggery portrait filenames.
var HumbuggeryImages []string

func loadHumbuggery(dir string) error {
	fm, _, err := loadFile(dir, "humbuggery/_index.md")
	if err != nil {
		return err
	}

	var page struct {
		Images []string `yaml:"images"`
	}
	if err := yaml.Unmarshal(fm, &page); err != nil {
		return err
	}

	HumbuggeryImages = page.Images

	// Hardcode the intro text (not in the markdown files)
	HumbuggeryIntro = `Here we honor the legendary men who have ascended to the highest echelons of our board, carrying the torch of our sacred traditions through the roaring tides of history into the present day. With a resolve as unyielding as granite and a refusal to give any sucker an even break, these intrepid volunteers have dedicated their hearts and minds to forging unforgettable functions, raising funds to aid widows, orphans, and families in need. Even in the shadow of the darkest days, they have risen like titans, tearing away the blindfold of despair to illuminate a path of hope and possibility. These are not merely board members—they are stewards of dreams, architects of change, and torchbearers of the indomitable spirit that defines us all.`

	return nil
}
