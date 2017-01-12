package travelkit;

import (
  "fmt"
  "math"
  "io/ioutil"
  "sort"
  "strings"
)

import (
  "github.com/ttacon/chalk"
  "github.com/imdario/mergo"
)

func CollectContacts(s Settings, media_list []MediaAttributes, media_map map[string]MediaAttributes) ([]Contact, map[string]Contact, error) {
	contacts_list := make([]Contact,0)
	contacts_map := make(map[string]Contact)

  media_vcf := make([]MediaAttributes, 0)
  media_vcf = filterByType(media_list, "vcf")

  for _, x := range media_vcf {
    content, err := ioutil.ReadFile(x.Path)
    if err != nil {
      msg := "Could not open .VCF file with id "+x.Id+"."
      fmt.Println(chalk.Cyan, msg, chalk.Reset)
    }
    lines := strings.Split(string(content), "\n")
    y := Contact{}
    numbers := make([]string, 0)
    emails := make([]string, 0)
    photo := ""
    readingPhoto := false
    for _, line := range lines {
      line_trimmed := Trim(line)
      if line_trimmed == "BEGIN:VCARD" {
        y = Contact{}
        numbers = make([]string, 0)
        emails = make([]string, 0)
        photo = ""
        readingPhoto = false
      } else if line_trimmed == "END:VCARD" {
        readingPhoto = false
        mergo.Merge(&y, Contact{
          Id: strings.ToLower(y.GivenName+"-"+y.FamilyName),
          Numbers: numbers,
          Emails: emails,
          Photo: photo,
        })
        contacts_list = append(contacts_list, y)
        contacts_map[y.Id] = y
      } else if strings.HasPrefix(line_trimmed, "N:") {
        readingPhoto = false
        n := strings.Split(line[2:], ";")
        y.GivenName = n[1]
        y.FamilyName = n[0]
      } else if strings.HasPrefix(line_trimmed, "EMAIL;") {
        readingPhoto = false
        emails = append(emails, strings.Split(line, ":")[1])
      } else if strings.HasPrefix(line_trimmed, "TEL;") {
        readingPhoto = false
        numbers = append(emails, strings.Split(line, ":")[1])
      } else if strings.HasPrefix(line_trimmed, "PHOTO;ENCODING=BASE64;JPEG:") {
        photo = strings.Split(line, ":")[1]
        readingPhoto = true
      } else if readingPhoto {
        if len(line_trimmed) == 0 {
          readingPhoto = false
        } else {
          photo += line_trimmed
        }
      }
    }
  }
	fmt.Println(chalk.Cyan, "Done collecting contacts", chalk.Reset)
	return contacts_list, contacts_map, nil
}

func filterContactsByText(contacts_in []Contact, text string) []Contact {
	if len(Trim(text)) == 0 {
		return contacts_in
	} else {
		text_lc := strings.ToLower(text)
		contacts_out := make([]Contact, 0)
		for _, x := range contacts_in {
			if strings.Contains(strings.ToLower(x.GivenName+" "+x.FamilyName), text_lc) {
				contacts_out = append(contacts_out, x)
			}
		}
		return contacts_out
	}
}

func FilterContacts(contacts_in []Contact, text string, pageSize int, pageNumber int, order string) []Contact {
	contacts_out := make([]Contact, 0)
	contacts_out = filterContactsByText(contacts_in, text)

  // Order Media
  if len(order) > 0 {
		if order == "z_a" {
			sort.Sort(ContactsByZA(contacts_out))
		} else {
			sort.Sort(ContactsByAZ(contacts_out))
    }
	}

  if pageSize > 0 {
		if pageNumber > 0 {
			start := int(math.Min(float64(len(contacts_out)), float64(pageSize*pageNumber)))
			end := int(math.Min(float64(len(contacts_out)), float64(pageSize*(pageNumber+1))))
			return contacts_out[start:end]
		} else {
			start := 0
			end := int(math.Min(float64(len(contacts_out)), float64(pageSize*(pageNumber+1))))
			return contacts_out[start:end]
	  }
	}

	return contacts_out
}


func CreateOrdersForConatcts(text string, currentOrder string) ([]map[string]string) {

	list := []map[string]string{}

	x := map[string]string{
	  "id": "a_z",
	  "title": "A - Z",
	  "url": "/media?order=a_z&text="+text,
	  "class": "dropdown-item",
	}
	if x["id"] == currentOrder { x["class"] = x["class"] + " disabled"; x["url"] = "#";}
	list = append(list, x)
	//
	x = map[string]string{
	  "id": "z_a",
	  "title": "Z - A",
	  "url": "/media?order=z_a&text="+text,
	  "class": "dropdown-item",
	}
	if x["id"] == currentOrder { x["class"] = x["class"] + " disabled"; x["url"] = "#";}
	list = append(list, x)

	return list
}
