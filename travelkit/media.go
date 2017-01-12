package travelkit

import (
	"errors"
	"fmt"
	//"log"
	"os"
	"path"
	"math"
	"strings"
	"time"
	"strconv"
	"sort"
	//"reflect"
	"path/filepath"
)

import (
	"github.com/ttacon/chalk"
	"github.com/mattn/go-zglob"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/imdario/mergo"
)

func ParseFilename(filename string, lower bool) (string, string) {
	if strings.Contains(filename, ".") {
		ext := filepath.Ext(filename)
		if len(ext) > 0 {
			ext = ext[1:]
		}
		if lower {
		  return strings.TrimSuffix(filename, ext), strings.ToLower(ext)
		} else {
			return strings.TrimSuffix(filename, ext), ext
		}
	} else {
		return filename, ""
	}
}

func filterByType(media_in []MediaAttributes, typeOfMedia string) []MediaAttributes {
	if typeOfMedia == "all" || typeOfMedia == "" {
		return media_in
	} else {
		media_out := make([]MediaAttributes, 0)
		for _, x := range media_in {
			if x.TypeOfMedia.Id == typeOfMedia {
				media_out = append(media_out, x)
			}
		}
		return media_out
	}
}

func filterByDays(media_in []MediaAttributes, days int) []MediaAttributes {
	if days <= 0 {
		return media_in
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

		media_out := make([]MediaAttributes, 0)
		for _, x := range media_in {
			if int(today.Sub(x.Date).Hours()) <= (24.0 * days) {
				media_out = append(media_out, x)
			}
		}
		return media_out
	}
}

func filterByText(media_in []MediaAttributes, text string) []MediaAttributes {
	if len(Trim(text)) == 0 {
		return media_in
	} else {
		text_lc := strings.ToLower(text)
		media_out := make([]MediaAttributes, 0)
		for _, x := range media_in {
			if strings.Contains(strings.ToLower(x.Id), text_lc) {
				media_out = append(media_out, x)
			}
		}
		return media_out
	}
}

func FilterMedia(media_in []MediaAttributes, typeOfMedia string, days int, text string, pageSize int, pageNumber int, order string) []MediaAttributes {
	media_out := make([]MediaAttributes, 0)

  // Filter Media
  media_out = filterByType(media_in, typeOfMedia)
	media_out = filterByDays(media_out, days)
	media_out = filterByText(media_out, text)

  // Order Media
  if len(order) > 0 {
		if order == "least_recent" {
			sort.Sort(MediaAttributesByLeastRecent(media_out))
		} else if order == "a_z" {
			sort.Sort(MediaAttributesByAZ(media_out))
		} else if order == "z_a" {
			sort.Sort(MediaAttributesByZA(media_out))
		} else {
		  sort.Sort(MediaAttributesByMostRecent(media_out))
		}
	}

  if pageSize > 0 {
		if pageNumber > 0 {
			start := int(math.Min(float64(len(media_out)), float64(pageSize*pageNumber)))
			end := int(math.Min(float64(len(media_out)), float64(pageSize*(pageNumber+1))))
			return media_out[start:end]
		} else {
			start := 0
			end := int(math.Min(float64(len(media_out)), float64(pageSize*(pageNumber+1))))
			return media_out[start:end]
	  }
	}

	return media_out
}

func ParseDate(filename string) (time.Time, error) {
	name, _ := ParseFilename(filename, true)
	if len(name) == 15 {

		year, err := strconv.Atoi(name[0:4])
		if err != nil { return time.Time{}, err }

		month, err := strconv.Atoi(name[4:6])
		if err != nil { return time.Time{}, err }

		date, err := strconv.Atoi(name[6:8])
		if err != nil { return time.Time{}, err }

		hour, err := strconv.Atoi(name[9:11])
		if err != nil { return time.Time{}, err }

		minute, err := strconv.Atoi(name[11:13])
		if err != nil { return time.Time{}, err }

		second, err := strconv.Atoi(name[13:15])
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

	} else {
		return time.Time{} , errors.New("Could not parse date from filename")
	}
}

func ParseType(settings Settings, pathtofile string) (MediaType, error) {
	filename := path.Base(pathtofile)
	_, ext := ParseFilename(filename, true)
	for _, x := range settings.Media.Types {
		for _ , y := range x.Extensions {
			if ext == y {
				return x, nil
			}
		}
	}
	return MediaType{}, errors.New("Could not find media type for file at "+ pathtofile)
}


//func ParseAttributes(pathtofile string) (string, time.Time, int, int, int) {
func ParseAttributes(settings Settings, pathtofile string) (MediaAttributes, error) {
	//fmt.Println(chalk.Cyan, "Parsing Attributes for ", pathtofile, chalk.Reset)
	filename := path.Base(pathtofile)
	_, ext := ParseFilename(filename, true)

  typeOfMedia, err := ParseType(settings, pathtofile)
	if err != nil {
		return MediaAttributes{}, err
	}

	d:= time.Time{}
	info, err := os.Stat(pathtofile)
	if err == nil {
		d = info.ModTime()
	}

	r := 0
	w := 0
	h := 0

  if ext == "mp4" {
		d, err = ParseDate(filename)
		if err != nil {
			fmt.Println(chalk.Red, "Error: Could not parse date from mp4 file '"+pathtofile+"'.", chalk.Reset)
		}
	} else if ext == "jpg" || ext == "jpeg" {
		d, err = ParseDate(filename)
		if err != nil {
			fmt.Println(chalk.Red, "Error: Could not parse date from JPEG file '"+pathtofile+"'.", chalk.Reset)
		}

		jpeg, err := os.Open(pathtofile)
		if err != nil {
				fmt.Println(chalk.Red, "Error: Could not JPEG file at '"+pathtofile+"'.", chalk.Reset)
		}
		exifdata, err := exif.Decode(jpeg)
		if err != nil {
			fmt.Println(chalk.Red, "Error: Could not decode JPEG file at '"+pathtofile+"'.", chalk.Reset)
		} else {
			exif_datetime, err := exifdata.DateTime()
			//fmt.Println(reflect.TypeOf(exif_datetime))
			if err == nil {
				//fmt.Println(exif_datetime)
				d = exif_datetime
			}

			exif_width, err := exifdata.Get(exif.ImageWidth)
			if err == nil {
				exif_width_value, err := exif_width.Int(0)
				if err == nil {
					w = exif_width_value
				}
			}

			exif_height, err := exifdata.Get(exif.ImageLength)
			if err == nil {
				exif_height_value, err := exif_height.Int(0)
				if err == nil {
					h = exif_height_value
				}
			}

			orientation, err := exifdata.Get(exif.Orientation)
			if err == nil {
				orientation_value, err := orientation.Int(0)
				if err == nil {
					if orientation_value  == 6 {
						r = 90
					}
				}
			}
		}
	}
	return MediaAttributes{TypeOfMedia: typeOfMedia, Date: d, Rotation: r, Width: w, Height: h}, nil
}

func ParseAttributesError(pathtofile string) (MediaAttributes, error) {
	return MediaAttributes{}, errors.New("Error: via ParseAttributesError!")
}

func CollectMedia(s Settings, directories []string) ([]MediaAttributes, map[string]MediaAttributes, error) {
	media_list := make([]MediaAttributes,0)
	media_map := make(map[string]MediaAttributes)

  for _, x := range directories {
		dir := Trim(x)
		if len(dir) == 0 {
			fmt.Println(chalk.Cyan, "Skipping blank media location", dir, chalk.Reset)
		} else {
			fmt.Println(chalk.Cyan, "Collecting media at location", dir, chalk.Reset)
			files, err := zglob.Glob(normalizePath(dir))

			if err != nil {
				return nil , nil , err
			}

			for _ , f := range files {
				//fmt.Println(chalk.Cyan, "Collecting media file", f, chalk.Reset)
				stats, err := os.Stat(f)
				if err != nil {
					fmt.Println(chalk.Red, "Could not stat path", "--", f, chalk.Reset)
				} else {
					if stats.Mode().IsRegular() {
						basepath := path.Base(f)
						id := basepath
						attrs, err := ParseAttributes(s, f)
						if err != nil {
							fmt.Println(chalk.Red, err, "--", f, chalk.Reset)
						} else {
							mergo.Merge(&attrs, MediaAttributes{
								Id: id,
								Path: f,
								IsText: attrs.TypeOfMedia.Id == "text",
								IsImage: attrs.TypeOfMedia.Id == "image",
								IsVideo: attrs.TypeOfMedia.Id == "video",
								IsGeoJSON: attrs.TypeOfMedia.Id == "geojson",
							})
							media_map[id] = attrs
							media_list = append(media_list, media_map[id])
						}
					}
				}
			}
		}
	}
	fmt.Println(chalk.Cyan, "Done collecting media", chalk.Reset)
	return media_list, media_map, nil
}
