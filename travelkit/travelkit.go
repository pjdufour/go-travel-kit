package main

import (
	"archive/zip"
	"fmt"
	"log"
	"os"
	"path"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
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
	Path string
	Rotation int
	Date time.Time
	Width int
	Height int
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


func ParseAttributes(pathtofile string) (time.Time, int, int, int) {
	filename := path.Base(pathtofile)
	ext := strings.Split(filename,".")[1]

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

	return d, r, w, h
}

func Collect(dir string) ([]string, map[string]string, error) {
	files, err := zglob.Glob(dir)

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

func CollectImages(dir string) ([]string, map[string]MediaAttributes, error) {
	files, err := zglob.Glob(dir)

	if err != nil {
		return nil , nil , err
	}

	m := make(map[string]MediaAttributes)
	for _ , f := range files {
		basepath := path.Base(f)
		//ext := strings.Split(basepath,".")[1]
		id := basepath
		date, r, w, h := ParseAttributes(f)
		m[id] = MediaAttributes{f, r, date, w, h}
	}
	return files, m, err
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
		os.Mkdir(path_home+"/repo", os.FileMode(0755))
		resp, err := http.Get("https://github.com/pjdufour/go-travel-kit/archive/master.zip")
		defer resp.Body.Close()
		if err != nil {
			return
		}
		out, err := os.Create(path_home+"/temp/go-travel-kit.zip")
		defer out.Close()
		_, _ = io.Copy(out, resp.Body)
		unzip(path_home+"/temp/go-travel-kit.zip", path_home+"/repo")
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
	path_templates := extract.Extract("http.templates", f.Root, "").(string)
	if path_templates == "" {
	  setup(path_home)
		path_templates = "~/.travelkit/repo/templates/*"
	}

  path_photos := extract.Extract("photos", f.Root, "").(string)
	siteurl := extract.Extract("http.siteurl", f.Root, "").(string)


  fmt.Println("Photos", path_photos)

	if path_photos == "" {
		return
	}

  photos_list, photos_map, photos_err := CollectImages(path_photos)
	//file_photos, err = zglob.Glob(path_photos)
	if photos_err != nil {
		fmt.Println(photos_err)
		return
	}
  fmt.Println(photos_list)

	templates_list, _, templates_err := Collect(path_templates)
	if templates_err != nil {
		fmt.Println(templates_err)
		return
	}
	fmt.Println(templates_list)
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
		err = tmpl.ExecuteTemplate(w, "media.html", struct{Media map[string]MediaAttributes}{photos_map})
	});

	router.GET("/media/view/:id", func(w http.ResponseWriter, r *http.Request, params map[string]string){
			id := params["id"]
			image := photos_map[id]
			ctx := struct{Title string; Image string; Width int; Height int; Rotation int}{
				id,
				siteurl+"/api/media/download/"+id,
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

  group.GET("/media/download/:id", func(w http.ResponseWriter, r *http.Request, params map[string]string){
	    id := params["id"]

			img, err := os.Open(photos_map[id].Path)
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
