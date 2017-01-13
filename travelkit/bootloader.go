package travelkit

import (
  "errors"
  "fmt"
  "html/template"
  "os"
  "sort"
  "reflect"
  "strconv"
  "strings"
)

import (
  "github.com/pjdufour/go-gypsy/yaml"
  "github.com/pjdufour/go-extract/extract"
  "github.com/GeertJohan/go.rice"
  "github.com/ttacon/chalk"
)

func ExtractString(envar string, keyChain string, node yaml.Node, fallback string) string {
  if len(envar) > 0 {
    if x := os.Getenv(envar); len(x) > 0 {
      return x
    }
  }
  if len(keyChain) > 0 {
    value := extract.Extract(keyChain, node, fallback)
  	if reflect.TypeOf(value).String() == "yaml.Scalar" {
  		return Trim(value.(yaml.Scalar).String())
  	} else if reflect.TypeOf(value).String() == "string" {
  		return Trim(value.(string))
  	} else {
  		return fallback
  	}
  } else {
    return fallback
  }
}

func ExtractStringList(envar string, keyChain string, node yaml.Node, fallback []string) []string {
  if len(envar) > 0 {
    if x := os.Getenv(envar); len(x) > 0 {
      return strings.Split(x, ":")
    }
  }
  if len(keyChain) > 0 {
    value := extract.Extract(keyChain, node, fallback)
    if reflect.TypeOf(value).String() == "yaml.List" {
      return ConvertYAMLListToStringList(value.(yaml.List))
    } else if reflect.TypeOf(value).String() == "yaml.Scalar" {
      out := Trim(value.(yaml.Scalar).String())
      return []string{out}
    } else if reflect.TypeOf(value).String() == "[]string" {
      return value.([]string)
    } else {
      return fallback
    }
  } else {
    return fallback
  }
}

func ExtractInt(envar string, keyChain string, node yaml.Node, fallback int) int {
  if len(envar) > 0 {
    if x := os.Getenv(envar); len(x) > 0 {
      if i, err := strconv.Atoi(x); err == nil {
        return i
      }
    }
  }
  if len(keyChain) > 0 {
    value := extract.Extract(keyChain, node, fallback)
  	if reflect.TypeOf(value).String() == "yaml.Scalar" {
  		i, err := strconv.Atoi(Trim(value.(yaml.Scalar).String()))
  		if err != nil {
  			return fallback
  		} else {
  			return i
  		}
  	} else if reflect.TypeOf(value).String() == "int" {
  		return value.(int)
  	} else {
  		return fallback
  	}
  } else {
    return fallback
  }
}

func ExtractMediaTypes(keyChain string, node yaml.Node, fallback []MediaType) []MediaType {
  if len(keyChain) > 0 && node != nil {
    value := extract.Extract(keyChain, node, "")
    if reflect.TypeOf(value).String() == "yaml.List" {
      mediaTypes := make([]MediaType, 0)
      for _, x := range value.(yaml.List) {
        //fmt.Println("type of ", x, "is", reflect.TypeOf(x).String())
        if reflect.TypeOf(x).String() == "yaml.Map" {
          y := x.(yaml.Map)
          //fmt.Println(reflect.TypeOf(y["extensions"]).String())
          mediaTypes = append(mediaTypes, MediaType{
            Id: Trim(y["id"].(yaml.Scalar).String()),
            Title: Trim(y["title"].(yaml.Scalar).String()),
            Extensions: ConvertYAMLListToStringList(y["extensions"].(yaml.List)),
          })
        }
      }
      return mediaTypes
    } else {
      return fallback
    }
  } else {
    return fallback
  }
}

func LoadSettings(filename string) (Settings) {

	f, err := yaml.ReadFile(filename)

	if err != nil {
		fmt.Println(chalk.Cyan, "Error: Could not open settings file at "+filename+".  Continuing with default.", chalk.Reset)
    return Settings{
  		Home: ExtractString("TRAVELKIT_HOME", "", nil, DEFAULT_HOME),
  		Site: Site{
  			Name: ExtractString("SITE_NAME", "", nil, DEFAULT_SITE_NAME),
  			Url: ExtractString("SITE_URL", "", nil, DEFAULT_SITE_URL),
  		},
  		Media: Media{
  			Types: DEFAULT_MEDIA_TYPES,
  			Page_Size: ExtractInt("MEDIA_PAGE_SIZE", "", nil, DEFAULT_PAGE_SIZE),
  			Locations: ExtractStringList("MEDIA_LOCATIONS", "", nil, DEFAULT_MEDIA_LOCATIONS),
  		},
  	}
	} else {
    return Settings{
  		Home: ExtractString("TRAVELKIT_HOME", "TRAVELKIT_HOME", f.Root, DEFAULT_HOME),
  		Site: Site{
  			Name: ExtractString("SITE_NAME", "SITE.NAME", f.Root, DEFAULT_SITE_NAME),
  			Url: ExtractString("SITE_URL", "SITE.URL", f.Root, DEFAULT_SITE_URL),
  		},
  		Media: Media{
  			Types: ExtractMediaTypes("MEDIA.TYPES", f.Root, DEFAULT_MEDIA_TYPES),
  			Page_Size: ExtractInt("MEDIA_PAGE_SIZE", "MEDIA.PAGE_SIZE", f.Root, DEFAULT_PAGE_SIZE),
  			Locations: ExtractStringList("MEDIA_LOCATIONS", "MEDIA.LOCATIONS", f.Root, DEFAULT_MEDIA_LOCATIONS),
  		},
  	}
  }
}

func check(s Settings) error {
	if len(s.Media.Types) == 0 {
		return errors.New("Error: s.Media.Types is an empty.")
	} else {
	  return nil
  }
}

/*func LoadTemplatesFromDisk(s Settings) (* template.Template, error) {

  // Load Templates //
  templates_list, _, err := CollectFiles(s.Templates)
  if err != nil {
    fmt.Println(chalk.Red, "Could Not Collect templates files from ", s.Templates, chalk.Reset)
    return nil, err
  }

  templateFilters := template.FuncMap{
    "first": firstItem,
    "join": Join,
    "date": formatDate,
    "time": formatTime,
  }

  tmpl, err := template.New("blank.tpl.html").Funcs(templateFilters).ParseFiles(templates_list...)
  return tmpl, err
}*/

func LoadTemplatesFromBinary(s Settings) (* template.Template, error) {

  templatesBox, err := rice.FindBox("templates")
	if err != nil {
		return nil, errors.New("Error: Could not find Rice Box for templates.")
	}

  templateFilters := template.FuncMap{
    "first": firstItem,
    "isLast": isLast,
    "isNotLast": isNotLast,
    "join": Join,
    "date": formatDate,
    "time": formatTime,
  }

  tmpl := template.New("blank.tpl.html").Funcs(templateFilters)

  root, err := templatesBox.Open("/")
  if err != nil {
    return nil, errors.New("Error: Could not open root of rice box for templates.")
  }

  stat, err := root.Stat()
  if err != nil {
  	return nil, errors.New("Error: Could not stat root of rice box for templates.")
  }

	if !stat.IsDir() {
		return nil, errors.New("Error: Root of rice box for templates is not a directory.")
	}

	infos, err := root.Readdir(0)
	root.Close()
	if err != nil {
		return nil, errors.New("Error: Could not read directory names from rice box for templates.")
	}

	var paths []string
	for _, info := range infos {
		paths = append(paths, info.Name())
	}
  sort.Strings(paths)

  for _, path := range paths {
    //fmt.Println(chalk.Cyan, "Loading template for path", path, chalk.Reset)
    content, err := templatesBox.String(path)
    if err != nil {
      return nil, errors.New("Error: Could not read content from template at '"+path+"'.")
    }
    //fmt.Println(chalk.Cyan, "Loading template content", content, chalk.Reset)
    tmpl, err = tmpl.New(path).Parse(content)
    if err != nil {
      return nil, errors.New("Error: Could not parse content from template at '"+path+"'.")
    }
  }

  return tmpl, err
}


//fmt.Println(chalk.Red, "Templates List:", templates_list, chalk.Reset)
//tmpl, err := template.ParseFiles(templates_list...)
//fmt.Println(chalk.Red, "Templates Compiled:", tmpl, chalk.Reset)
//tmpl, err := template.ParseFiles("index.html", "_include_header.tpl.html", "_include_head.tpl.html")
