package travelkit

import (
  "os"
  "strconv"
  "strings"
  "net/http"
  "path"
)

func Coalesce(a string, b string) string {
  if len(a) > 0 {
    return a
  }
  return b
}

func Unique(list []string) []string {
  out := make([]string, 0)
  set := make(map[string]int)
  for _, x := range list {
    set[x] = 1
  }
  for x, _ := range set {
    out = append(out, x)
  }
  return out
}

func Join(list []string, delimiter string) string {
  return strings.Join(list, delimiter)
}

func Stringify(x map[string]int) (map[string]string) {
	y := map[string]string{}
	for i, v := range x {
		y[i] = strconv.Itoa(v)
	}
	return y
}

func Trim(x string) string {
	x = strings.Trim(x, " \t\n\r")
	if strings.HasPrefix(x, "\"") && strings.HasSuffix(x, "\"") {
		return x[1:len(x)-1]
	} else {
		return x
	}
}

func exists(path string) (bool, error) {
    _, err := os.Stat(path)
    if err == nil { return true, nil }
    if os.IsNotExist(err) { return false, nil }
    return true, err
}

func param(r *http.Request, params map[string]string, name string, fallback string) string {
	value := r.URL.Query().Get(name)
	if len(value) == 0 {
		value, ok := params[name]
		if ! ok {
			return fallback
		} else {
			return value
		}
	} else {
		return value
	}
}

func formatAsInteger(x int) (string, error) {
	return strconv.FormatInt(int64(x), 10), nil
}

func buildCountsByType(s Settings, media_list []MediaAttributes, order string) map[string]int {
  countsByType := map[string]int{}
  countsByType["all"] = len(FilterMedia(media_list, "all", 0, "", 0, 0, order))
  for _, x := range s.Media.Types {
    countsByType[x.Id] = len(FilterMedia(media_list, x.Id, 0, "", 0, 0, order))
  }
  return countsByType
}

func buildYears(s Settings, media_list []MediaAttributes) []map[string]string {
  CountYears := map[int]int{}
  for _, x := range FilterMedia(media_list, "all", 0, "", 0, 0, "most_recent") {
    year := x.Date.Year()
    CountYears[year] = CountYears[year] + 1 // No need to set to zero
  }
  years := make([]map[string]string, 0)
  for year, count := range CountYears {
    years = append(years, map[string]string{
      "active": "false",
      "year": strconv.Itoa(year),
      "count": strconv.Itoa(count),
    })
  }
  return years
}

func fileExtensionToContentType(pathtofile string) string {
  switch _, ext := ParseFilename(path.Base(pathtofile), true); ext {
    case "css": return "text/css"
    case "json" : return "application/json"
    case "js" : return "application/javascript; charset=utf-8"
    default : return "text/plain"
  }
}
