package travelkit

import (
  "fmt"
  "math"
  "io/ioutil"
  "sort"
  "strings"
  "time"
  "strconv"
)

import (
  "github.com/ttacon/chalk"
  //"github.com/imdario/mergo"
)
//Tue, 17 Nov 2015 17:50:26 +0000
func ParseDateFromEmail(line string) (time.Time, error) {
  x := strings.Split(line, " ")

  year, err := strconv.Atoi(x[3])
  if err != nil { return time.Time{}, err }

  month := IndexOf(MONTHS_SHORT_3, x[2])

  date, err := strconv.Atoi(x[1])
  if err != nil { return time.Time{}, err }

  hour, err := strconv.Atoi(x[4][0:2])
  if err != nil { return time.Time{}, err }

  minute, err := strconv.Atoi(x[4][3:5])
  if err != nil { return time.Time{}, err }

  second, err := strconv.Atoi(x[4][6:8])
  if err != nil { return time.Time{}, err }

  d := time.Date(
    year,
    time.Month(month),
    date,
    hour,
    minute,
    second,
    0,
    time.UTC)

  return d, err
}

func CollectEmails(s Settings, media_list []MediaAttributes, media_map map[string]MediaAttributes) ([]Email, map[string]Email, error) {
	emails_list := make([]Email,0)
	emails_map := make(map[string]Email)

  media_mbox := make([]MediaAttributes, 0)
  media_mbox = filterByType(media_list, "mbox")

  for _, x := range media_mbox {
    content, err := ioutil.ReadFile(x.Path)
    if err != nil {
      msg := "Could not open .mbox file with id "+x.Id+"."
      fmt.Println(chalk.Cyan, msg, chalk.Reset)
    }
    lines := strings.Split(string(content), "\n")
    y := Email{}
    for _, line := range lines {
      line_trimmed := Trim(line)
      line_trimmed_lc := strings.ToLower(line_trimmed)
      if strings.HasPrefix(line_trimmed_lc, "from ") {
        if len(y.Id) > 0 {
          emails_list = append(emails_list, y)
          emails_map[y.Id] = y
          y = Email{}
        }
        y.Id = strings.Split(line_trimmed, "@")[0][len("From "):]
      } else if strings.HasPrefix(line_trimmed_lc, "from:") {
        y.From = line_trimmed[len("From:"):]
      } else if strings.HasPrefix(line_trimmed_lc, "to:") {
        y.TO = line_trimmed[len("To:"):]
      } else if strings.HasPrefix(line_trimmed_lc, "cc:") {
        y.CC = line_trimmed[len("Cc:"):]
      } else if strings.HasPrefix(line_trimmed_lc, "bcc:") {
        y.BCC = line_trimmed[len("Bcc:"):]
      } else if strings.HasPrefix(line_trimmed_lc, "subject:") {
        y.Subject = line_trimmed[len("Subject:"):]
      } else if strings.HasPrefix(line_trimmed_lc, "date:") {
        y.DateTime, _ = ParseDateFromEmail(Trim(line_trimmed[len("date:"):]))
      }
    }
    if len(y.Id) > 0 {
      emails_list = append(emails_list, y)
      emails_map[y.Id] = y
    }
  }
	fmt.Println(chalk.Cyan, "Done collecting emails", chalk.Reset)
	return emails_list, emails_map, nil
}

func filterEmailsByDays(emails_in []Email, days int) []Email {
	if days <= 0 {
		return emails_in
	} else {
		n := time.Now()
		today := time.Date(
			n.Year(),
			n.Month(),
			n.Day(),
			0,
			0,
			0,
			0,
			time.UTC)

		emails_out := make([]Email, 0)
		for _, x := range emails_in {
			if int(today.Sub(x.DateTime).Hours()) <= (24.0 * days) {
				emails_out = append(emails_out, x)
			}
		}
		return emails_out
	}
}

func filterEmailsByText(emails_in []Email, text string) []Email {
  fmt.Println(chalk.Cyan, "Filtering emails by text", text, ".", chalk.Reset)
	if len(Trim(text)) == 0 {
		return emails_in
	} else {
		text_lc := strings.ToLower(text)
		emails_out := make([]Email, 0)
		for _, x := range emails_in {
      y := x.From+" "+x.TO+" "+x.CC+" "+x.BCC+" "+x.Subject
			if strings.Contains(strings.ToLower(y), text_lc) {
				emails_out = append(emails_out, x)
			}
		}
		return emails_out
	}
}

func FilterEmails(emails_in []Email, days int, text string, pageSize int, pageNumber int, order string) []Email {
	emails_out := make([]Email, 0)
	emails_out = filterEmailsByText(filterEmailsByDays(emails_in, days), text)

  // Order Media
  if len(order) > 0 {
		if order == "z_a" {
			sort.Sort(EmailsByZA(emails_out))
		} else {
			sort.Sort(EmailsByAZ(emails_out))
    }
	}

  // Paginate
  if pageSize > 0 {
		if pageNumber > 0 {
			start := int(math.Min(float64(len(emails_out)), float64(pageSize*pageNumber)))
			end := int(math.Min(float64(len(emails_out)), float64(pageSize*(pageNumber+1))))
			return emails_out[start:end]
		} else {
			start := 0
			end := int(math.Min(float64(len(emails_out)), float64(pageSize*(pageNumber+1))))
			return emails_out[start:end]
	  }
	}

	return emails_out
}


func CreateOrdersForEmails(text string, currentOrder string) ([]map[string]string) {

	list := []map[string]string{}

	x := map[string]string{
	  "id": "a_z",
	  "title": "A - Z",
	  "url": "/emails?order=a_z&text="+text,
	  "class": "dropdown-item",
	}
	if x["id"] == currentOrder { x["class"] = x["class"] + " disabled"; x["url"] = "#";}
	list = append(list, x)
	//
	x = map[string]string{
	  "id": "z_a",
	  "title": "Z - A",
	  "url": "/emails?order=z_a&text="+text,
	  "class": "dropdown-item",
	}
	if x["id"] == currentOrder { x["class"] = x["class"] + " disabled"; x["url"] = "#";}
	list = append(list, x)

	return list
}
