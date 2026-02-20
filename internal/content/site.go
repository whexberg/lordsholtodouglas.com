package content

import (
	"gopkg.in/yaml.v3"
)

// SiteConfig holds global site info (populated by loader).
var SiteConfig struct {
	Name        string
	Subtitle    string
	OrgName     string
	Description string
}

func loadSiteConfig(dir string) error {
	fm, _, err := loadFile(dir, "_index.md")
	if err != nil {
		return err
	}

	var cfg struct {
		Title    string `yaml:"title"`
		Subtitle string `yaml:"subtitle"`
	}
	if err := yaml.Unmarshal(fm, &cfg); err != nil {
		return err
	}

	SiteConfig.Name = cfg.Title
	SiteConfig.Subtitle = cfg.Subtitle
	SiteConfig.OrgName = cfg.Title + " Chapter #3 ECV"
	SiteConfig.Description = "A fraternal organization dedicated to the preservation of Western heritage, the appreciation of widders, and the pursuit of Credo Quia Absurdum since 1857."
	return nil
}
