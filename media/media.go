package media

import (
	"errors"
	//"fmt"
	"log"
	"os"
	"path"
	"math"
	"strings"
	"time"
	"strconv"
	"sort"
	//"reflect"
	//"path/filepath"
)

import (
	//"github.com/ttacon/chalk"
	//"github.com/mattn/go-zglob"
	"github.com/rwcarlsen/goexif/exif"
)

import (
  "github.com/pjdufour/go-travel-kit/types"
)

func ParseFilename(filename string) (string, string) {
	if strings.Contains(filename, ".") {
		s := strings.Split(filename,".")
		return s[0], s[1]
	} else {
		return filename, ""
	}
}

func filterByType(media_in []types.MediaAttributes, typeOfMedia string) []types.MediaAttributes {
	if typeOfMedia == "all" || typeOfMedia == "" {
		return media_in
	} else {
		media_out := make([]types.MediaAttributes, 0)
		for _, x := range media_in {
			if x.TypeOfMedia == typeOfMedia {
				media_out = append(media_out, x)
			}
		}
		return media_out
	}
}

func filterByDays(media_in []types.MediaAttributes, days int) []types.MediaAttributes {
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

		media_out := make([]types.MediaAttributes, 0)
		for _, x := range media_in {
			if int(today.Sub(x.Date).Hours()) <= (24.0 * days) {
				media_out = append(media_out, x)
			}
		}
		return media_out
	}
}
func filterByText(media_in []types.MediaAttributes, text string) []types.MediaAttributes {
	if len(strings.Trim(text, " \t\n\r")) == 0 {
		return media_in
	} else {
		media_out := make([]types.MediaAttributes, 0)
		for _, x := range media_in {
			if strings.Contains(x.Id, text) {
				media_out = append(media_out, x)
			}
		}
		return media_out
	}
}

func FilterMedia(media_in []types.MediaAttributes, typeOfMedia string, days int, text string, pageSize int, pageNumber int, order string) []types.MediaAttributes {
	media_out := make([]types.MediaAttributes, 0)

  // Filter Media
  media_out = filterByType(media_in, typeOfMedia)
	media_out = filterByDays(media_out, days)
	media_out = filterByText(media_out, text)

  // Order Media
  if len(order) > 0 {
		if order == "least_recent" {
			sort.Sort(types.MediaAttributesByLeastRecent(media_out))
		} else if order == "a_z" {
			sort.Sort(types.MediaAttributesByAZ(media_out))
		} else if order == "z_a" {
			sort.Sort(types.MediaAttributesByZA(media_out))
		} else {
		  sort.Sort(types.MediaAttributesByMostRecent(media_out))
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
	name, _ := ParseFilename(filename)
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


//func ParseAttributes(pathtofile string) (string, time.Time, int, int, int) {
func ParseAttributes(pathtofile string) (types.MediaAttributes, error) {
	filename := path.Base(pathtofile)
	_, ext := ParseFilename(filename)

	if ext == "mp4" {
		typeOfMedia := "video"
		d, err := ParseDate(filename)
		if err != nil {
				return types.MediaAttributes{}, errors.New("Error: Could not parse date from mp4 filename.")
		}
		r := 0
		w := 0
		h := 0
		return types.MediaAttributes{TypeOfMedia: typeOfMedia, Date: d, Rotation: r, Width: w, Height: h}, nil
		//return struct {TypeOfMedia string; Date time.Time; Rotation int; Width int; Height int}{typeOfMedia, d, r, w, h,}, nil
		//return typeOfMedia, d, r, w, h

	} else if ext == "jpg" || ext == "jpeg" {
		typeOfMedia := "image"
		d, err := ParseDate(filename)
		r := 0
		w := 0
		h := 0
		jpeg, err := os.Open(pathtofile)
		if err != nil {
				log.Fatal(err)
		}
		exifdata, err := exif.Decode(jpeg)
		if err != nil {
			log.Fatal(err)
		}

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
		//fmt.Println("Orientation", orientation)
		//return typeOfMedia, d, r, w, h
		//return struct {TypeOfMedia string; Date time.Time; Rotation int; Width int; Height int}{typeOfMedia, d, r, w, h,}, nil
		return types.MediaAttributes{TypeOfMedia: typeOfMedia, Date: d, Rotation: r, Width: w, Height: h}, nil
	} else {
		return types.MediaAttributes{}, errors.New("Error: Could not identify media type.")
	}

	//return typeOfMedia, d, r, w, h
}

func ParseAttributesError(pathtofile string) (types.MediaAttributes, error) {
	return types.MediaAttributes{}, errors.New("Error: via ParseAttributesError!")
}
