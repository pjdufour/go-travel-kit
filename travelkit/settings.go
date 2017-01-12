package travelkit

import (
  "errors"
  "github.com/pjdufour/go-gypsy/yaml"
)

func LoadSettings(filename string) (Settings, error) {

	f, err := yaml.ReadFile(filename)

	if err != nil {
		msg := "Error: Could not open settings file at "+filename+"."
		return Settings{} , errors.New(msg)
	}

  s := Settings{
		Home: ExtractString("TRAVELKIT_HOME", f.Root, ""),
		Site: Site{
			Name: ExtractString("SITE.NAME", f.Root, "Travel Kit"),
			Url: ExtractString("SITE.URL", f.Root, "http://localhost:8000"),
		},
		Media: Media{
			Types: ExtractMediaTypes("MEDIA.TYPES", f.Root),
			Page_Size: ExtractInt("MEDIA.PAGE_SIZE", f.Root, 100),
			Locations: ExtractStringList("MEDIA.LOCATIONS", f.Root, [](string){"~"}),
		},
		Templates: ExtractString("TEMPLATES", f.Root, ""),
	}

	if s.Home == "" {
		s.Home = "~/.travelkit"
	}

  return s, nil
}

func check(s Settings) error {
	if len(s.Media.Types) == 0 {
		return errors.New("Error: s.Media.Types is an empty.")
	} else {
	  return nil
  }
}
