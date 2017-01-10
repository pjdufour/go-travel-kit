package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/user"
	"path"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"math"
	"html/template"
	"strings"
	"time"
	"strconv"
	//"reflect"
	"path/filepath"
)

import (
	"github.com/ttacon/chalk"
	"github.com/mattn/go-zglob"
  "github.com/pjdufour/go-gypsy/yaml"
  "github.com/pjdufour/go-extract/extract"
  "github.com/dimfeld/httptreemux"
	"github.com/rwcarlsen/goexif/exif"
	//"github.com/patrickmn/go-cache"
)

//func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
//    fmt.Fprint(w, "Not protected!\n")
//}

//type Page struct {
//	Title string
//	Image  string
//}

//func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
//    t, _ := template.ParseFiles(tmpl + ".html")
//    t.Execute(w, p)
//}


type MediaAttributes struct{
	Id string
	Path string
	TypeOfMedia string
	IsImage bool
	IsVideo bool
	Rotation int
	Date time.Time
	Width int
	Height int
}

func FilterMedia(media_in []MediaAttributes, typeOfMedia string, days int, pageSize int, pageNumber int) []MediaAttributes {
	media_out := make([]MediaAttributes, 0)

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

  if typeOfMedia == "all" || typeOfMedia == "" {
		if days <= 0 {
			media_out = media_in
		} else {
			for _, x := range media_in {
				if int(today.Sub(x.Date).Hours()) <= (24.0 * days) {
					media_out = append(media_out, x)
				}
			}
		}
	} else {
		if days <= 0 {
			for _, x := range media_in {
				if x.TypeOfMedia == typeOfMedia {
					media_out = append(media_out, x)
				}
			}
		} else {
			for _, x := range media_in {
				if x.TypeOfMedia == typeOfMedia {
					if int(today.Sub(x.Date).Hours()) <= (24.0 * days) {
						media_out = append(media_out, x)
					}
				}
			}
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

func unzip(archive, target string) error {
	reader, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(target, 0755); err != nil {
		return err
	}

	for _, file := range reader.File {
		path := filepath.Join(target, file.Name)
		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.Mode())
			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			return err
		}
		defer fileReader.Close()

		targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}
		defer targetFile.Close()

		if _, err := io.Copy(targetFile, fileReader); err != nil {
			return err
		}
	}

	return nil
}


func ParseAttributes(pathtofile string) (string, time.Time, int, int, int) {
	filename := path.Base(pathtofile)
	ext := strings.Split(filename,".")[1]

	typeOfMedia := "image"
	if ext == "mp4" {
		typeOfMedia = "video"
	}

	year, _ := strconv.Atoi(filename[0:4])
	month, _ := strconv.Atoi(filename[4:6])
	date, _ := strconv.Atoi(filename[6:8])
	hour, _ := strconv.Atoi(filename[9:11])
	minute, _ := strconv.Atoi(filename[11:13])
	second, _ := strconv.Atoi(filename[13:15])

	d := time.Date(
		year,
		time.Month(month),
		date,
		hour,
		minute,
		second,
		0,
		time.UTC)

	r := 0

	w := 0
	h := 0

	if ext == "jpeg" || ext == "jpg" {
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
			fmt.Println(exif_datetime)
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
		fmt.Println("Orientation", orientation)
	}

  //if r == 90 {
	//	return d, r, h, w
	//}

	return typeOfMedia, d, r, w, h
}

func normalizePath(pathtofile string) string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal( err )
	}
	return strings.Replace(pathtofile, "~", usr.HomeDir, -1)
}

func CollectFiles(dir string) ([]string, map[string]string, error) {
	files, err := zglob.Glob(normalizePath(dir))

	if err != nil {
		return nil , nil , err
	}

	m := make(map[string]string)
	for _ , f := range files {
		basepath := path.Base(f)
		//basepath_split := strings.Split(basepath, "_")
		//date = basepath_split[0]
		//time = strings.Split(basepath_split[1], ".")

  	//id = string.Split(basepath,".")[0]
		id := basepath
		//data = file, err = os.Open("img/" + vars["item"])
		m[id]= f
	}
	return files, m, err
}

func CollectMedia(dir string) ([]MediaAttributes, map[string]MediaAttributes, error) {
	media_list := make([]MediaAttributes,0)

	files, err := zglob.Glob(normalizePath(dir))

	if err != nil {
		return nil , nil , err
	}

	media_map := make(map[string]MediaAttributes)
	for _ , f := range files {
		basepath := path.Base(f)
		//ext := strings.Split(basepath,".")[1]
		id := basepath
		typeOfMedia, date, r, w, h := ParseAttributes(f)
		media_map[id] = MediaAttributes{
			id,
			f,
			typeOfMedia,
			typeOfMedia == "image",
			typeOfMedia == "video",
			r,
			date,
			w,
			h,
		}
		media_list = append(media_list, media_map[id])
	}
	return media_list, media_map, err
}

func exists(path string) (bool, error) {
    _, err := os.Stat(path)
    if err == nil { return true, nil }
    if os.IsNotExist(err) { return false, nil }
    return true, err
}

func setup(path_home string){
	fmt.Println(chalk.Cyan, "Setting up Travel Kit!", chalk.Reset)
	if home_exists, _ := exists(path_home); ! home_exists {
		fmt.Println(chalk.Red, "Travel Kit Home does not exist.  Creating now!", chalk.Reset)
		os.Mkdir(path_home, os.FileMode(0755))
		os.Mkdir(path_home+"/temp", os.FileMode(0755))
		os.Mkdir(path_home+"/repos", os.FileMode(0755))
		resp, err := http.Get("https://github.com/pjdufour/go-travel-kit/archive/master.zip")
		defer resp.Body.Close()
		if err != nil {
			return
		}
		out, err := os.Create(path_home+"/temp/go-travel-kit.zip")
		defer out.Close()
		_, _ = io.Copy(out, resp.Body)
		unzip(path_home+"/temp/go-travel-kit.zip", path_home+"/repos")
	}
}

func main(){
	args := os.Args

	filename := ""

	if len(args) > 1 {
		filename = args[1]
	} else {
		filename = "/home/vagrant/src/github.com/pjdufour/go-travel-kit/travelkit.yml"
	}

	f, err := yaml.ReadFile(filename)

	if err != nil {
		fmt.Println("Could not open input file ", filename)
		return
	}

  path_home := extract.Extract("TRAVELKIT_HOME", f.Root, "").(string)

	if path_home == "" {
		path_home = "~/.travelkit"
	}

	path_templates := extract.Extract("TEMPLATES", f.Root, "").(string)
	if path_templates == "" {
	  setup(path_home)
		path_templates = "~/.travelkit/repos/go-travel-kit-master/templates/*"
		fmt.Println(chalk.Green, "path_templates set to", path_templates, chalk.Reset)
	}

  path_photos := extract.Extract("photos", f.Root, "").(string)
	siteurl := extract.Extract("http.siteurl", f.Root, "").(string)
	MEDIA_PAGE_SIZE := extract.Extract("MEDIA_PAGE_SIZE", f.Root, 100).(int)

  fmt.Println("Photos", path_photos)

	if path_photos == "" {
		return
	}

  media_list, media_map, photos_err := CollectMedia(path_photos)
	//file_photos, err = zglob.Glob(path_photos)
	if photos_err != nil {
		fmt.Println(photos_err)
		return
	}
  fmt.Println(media_list)

	templates_list, _, templates_err := CollectFiles(path_templates)
	if templates_err != nil {
		fmt.Println(templates_err)
		return
	}
	//fmt.Println(templates_list)
	//tmpl, err := template.ParseFiles(templates_list)
	tmpl, err := template.ParseFiles(templates_list...)

	//files, err := ioutil.ReadDir(photos)
	//files, err := ioutil.ReadDir(".")
	//err = filepath.Walk(photos, func(path string, info os.FileInfo, err error) error {
	//	file_photos = append(file_photos, path)
	//	return nil
	//})
  //g := photos + "/[a-Z0-9]"
	//fmt.Println("G", g)
  //file_photos, err = filepath.Glob(photos)

	//fmt.Println("Files", file_photos)
	//for _, f := range file {
	//	file_photos = append(file_photos, f.Name())
	//}

	//fmt.Println(file_photos)

	//c := cache.New(5*time.Minute, 30*time.Second)

  router := httptreemux.New()

  //router.GET("/", Index)

	router.GET("/", func(w http.ResponseWriter, r *http.Request, params map[string]string){
		err = tmpl.ExecuteTemplate(w, "index.html", struct{}{})
	});

	router.GET("/media", func(w http.ResponseWriter, r *http.Request, params map[string]string){

		typeOfMedia := r.URL.Query().Get("type")
		if len(typeOfMedia) == 0 {
			typeOfMedia = "all"
		}

		ctx := struct{
			TypeOfMedia string;
			All bool;
			Images bool;
			Videos bool;
			CountAll int;
			CountImages int;
			CountVideos int;
			MediaAll []MediaAttributes;
			Media7Days []MediaAttributes;
			Media30Days []MediaAttributes;
			Media90Days []MediaAttributes
		}{
		  typeOfMedia,
		  typeOfMedia == "all",
			typeOfMedia == "image",
			typeOfMedia == "video",
			len(FilterMedia(media_list, "all", 0, 0, 0)),
			len(FilterMedia(media_list, "image", 0, 0, 0)),
			len(FilterMedia(media_list, "video", 0, 0, 0)),
      FilterMedia(media_list, typeOfMedia, 0, MEDIA_PAGE_SIZE, 0),
			FilterMedia(media_list, typeOfMedia, 7, MEDIA_PAGE_SIZE, 0),
			FilterMedia(media_list, typeOfMedia, 30, MEDIA_PAGE_SIZE, 0),
			FilterMedia(media_list, typeOfMedia, 90, MEDIA_PAGE_SIZE, 0),
    }
		err = tmpl.ExecuteTemplate(w, "media.html", ctx)
	});

	router.GET("/media/view/:id", func(w http.ResponseWriter, r *http.Request, params map[string]string){
			id := params["id"]
			image := media_map[id]
			ctx := struct{Title string; URI string; IsImage bool; IsVideo bool; Width int; Height int; Rotation int}{
				id,
				siteurl+"/api/media/download/"+id,
				image.IsImage,
				image.IsVideo,
				image.Width,
				image.Height,
				image.Rotation,
			}
			err = tmpl.ExecuteTemplate(w, "view.html", ctx)
	})

	router.GET("/about", func(w http.ResponseWriter, r *http.Request, params map[string]string){
		err = tmpl.ExecuteTemplate(w, "about.html", struct{}{})
	});

  group := router.NewGroup("/api")

	group.GET("/media/list/type/:type/days/:days/page/:page", func(w http.ResponseWriter, r *http.Request, params map[string]string){
			typeOfMedia := params["type"]
			if len(typeOfMedia) == 0 {
				typeOfMedia = "all"
			}
			days := 0
			if len(params["days"]) > 0 {
				days, _ = strconv.Atoi(params["days"])
			}

			pageNumber := 0
			if len(params["page"]) > 0 {
				pageNumber, _ = strconv.Atoi(params["page"])
			}

			ext := params["ext"]
			if ext == "" {
				ext = "json"
			}

			fmt.Println("params", params)

			data := FilterMedia(media_list, typeOfMedia, days, MEDIA_PAGE_SIZE, pageNumber)
			//w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
			if ext == "json" {
				w.Header().Set("Content-Type", "application/json")
				if err := json.NewEncoder(w).Encode(data); err != nil {
					panic(err)
				}
			} else if ext == "yml" {
				w.Header().Set("Content-Type", "plain/text")
				//yaml.
			}
	})


  group.GET("/media/download/:id", func(w http.ResponseWriter, r *http.Request, params map[string]string){
	    id := params["id"]

			img, err := os.Open(media_map[id].Path)
			defer img.Close()

			if err != nil {
				log.Println(err) // perhaps handle this nicer
				fmt.Fprintf(w, "Could not open media file at id /%s", id)
				return
			}

			data, err := ioutil.ReadAll(img)
			if err != nil {
				log.Println(err) // perhaps handle this nicer
				fmt.Fprintf(w, "Could not read media file at id /%s", id)
				return
			}

			w.Header().Set("Content-type", "image/jpeg")
			w.Header().Set("Content-Disposition", "attachment; filename="+id )
			w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
			w.Write(data)
	})

  //siteurl := extract.Extract("http.siteurl", f.Root, "").(string)
	u, err := url.Parse(siteurl)
	_, port, _ := net.SplitHostPort(u.Host)
	fmt.Println("port", port)
  log.Fatal(http.ListenAndServe(":"+port, router))
}
