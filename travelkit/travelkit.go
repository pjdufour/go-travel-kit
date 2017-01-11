package main

import (
	"errors"
	"bytes"
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
	//"math"
	"html/template"
	"strings"
	//"time"
	"strconv"
	"reflect"
	//"path/filepath"
	"image/jpeg"
)

import (
	"github.com/ttacon/chalk"
	"github.com/imdario/mergo"
	"github.com/mattn/go-zglob"
  "github.com/pjdufour/go-gypsy/yaml"
  "github.com/pjdufour/go-extract/extract"
  "github.com/dimfeld/httptreemux"
	"github.com/nfnt/resize"
	//"github.com/rwcarlsen/goexif/exif"
	//"github.com/patrickmn/go-cache"
)

import (
	"github.com/pjdufour/go-travel-kit/types"
	"github.com/pjdufour/go-travel-kit/unzip"
	"github.com/pjdufour/go-travel-kit/media"
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

func trim(path string) string {
	path = strings.Trim(path, " \n\r")
	if strings.HasPrefix(path, "\"") && strings.HasSuffix(path, "\"") {
		return path[1:len(path)-1]
	} else {
		return path
	}
}
func ExtractInt(keyChain string, node yaml.Node, fallback int) int {
	value := extract.Extract(keyChain, node, fallback)
	if reflect.TypeOf(value).String() == "yaml.Scalar" {
		i, err := strconv.Atoi(trim(value.(yaml.Scalar).String()))
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
}

func ExtractString(keyChain string, node yaml.Node, fallback string) string {
	value := extract.Extract(keyChain, node, fallback)
	if reflect.TypeOf(value).String() == "yaml.Scalar" {
		return trim(value.(yaml.Scalar).String())
	} else if reflect.TypeOf(value).String() == "string" {
		return trim(value.(string))
	} else {
		return fallback
	}
}

func ExtractStringList(keyChain string, node yaml.Node, fallback []string) []string {
  value := extract.Extract(keyChain, node, fallback)
	if reflect.TypeOf(value).String() == "yaml.List" {
    list := value.(yaml.List)
		out := make([](string), list.Len())
		for index, _ := range list {
			y := list.Item(index)
			if reflect.TypeOf(y).String() == "yaml.Scalar" {
				out[index] = trim(y.(yaml.Scalar).String())
			}
		}
		return out
	} else if reflect.TypeOf(value).String() == "yaml.Scalar" {
		out := trim(value.(yaml.Scalar).String())
		return []string{out}
	} else if reflect.TypeOf(value).String() == "[]string" {
		return value.([]string)
	} else {
	  return fallback
	}
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

func CollectMedia(directories []string) ([]types.MediaAttributes, map[string]types.MediaAttributes, error) {
	media_list := make([]types.MediaAttributes,0)
	media_map := make(map[string]types.MediaAttributes)

  for _, dir := range directories {
		fmt.Println(chalk.Cyan, "Collecting media at location", dir, chalk.Reset)
		files, err := zglob.Glob(normalizePath(dir))

		if err != nil {
			return nil , nil , err
		}

		for _ , f := range files {
			stats, err := os.Stat(f)
			if err != nil {
				fmt.Println(chalk.Red, "Could not stat path", "--", f, chalk.Reset)
			} else {
				if stats.Mode().IsRegular() {
					//fmt.Println(chalk.Cyan, "Parsing Attributes for ", f, chalk.Reset)
					basepath := path.Base(f)
					id := basepath
					attrs, err := media.ParseAttributes(f)
					if err != nil {
						fmt.Println(chalk.Red, err, "--", f, chalk.Reset)
					} else {
						mergo.Merge(&attrs, types.MediaAttributes{
							Id: id,
							Path: f,
							IsImage: attrs.TypeOfMedia == "image",
							IsVideo: attrs.TypeOfMedia == "video",
						})
						media_map[id] = attrs
						media_list = append(media_list, media_map[id])
					}
				}
			}
		}
	}

	return media_list, media_map, nil
}

func exists(path string) (bool, error) {
    _, err := os.Stat(path)
    if err == nil { return true, nil }
    if os.IsNotExist(err) { return false, nil }
    return true, err
}

func LoadSettings(filename string) (types.Settings, error) {

	f, err := yaml.ReadFile(filename)

	if err != nil {
		msg := "Error: Could not open settings file at "+filename+"."
		return types.Settings{} , errors.New(msg)
	}

  settings := types.Settings{
		Home: ExtractString("TRAVELKIT_HOME", f.Root, ""),
		Site: types.Site{
			Name: ExtractString("HTTP.SITE_NAME", f.Root, "Travel Kit"),
			Url: ExtractString("HTTP.SITE_URL", f.Root, "http://localhost:8000"),
		},
		Media: types.Media{
			Page_Size: ExtractInt("MEDIA_PAGE_SIZE", f.Root, 100),
			Locations: ExtractStringList("MEDIA.LOCATIONS", f.Root, make([](string), 0)),
		},
		Templates: ExtractString("TEMPLATES", f.Root, ""),
	}

	if settings.Home == "" {
		settings.Home = "~/.travelkit"
	}

  return settings, nil
}

func check(settings types.Settings) error {
	if len(settings.Media.Locations) == 0 {
		return errors.New("Error: settings.Media.Locations is an empty.")
	} else {
	  return nil
  }
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
		unzip.Unzip(path_home+"/temp/go-travel-kit.zip", path_home+"/repos")
	}
}

func param(r *http.Request, name string, fallback string) string {
	value := r.URL.Query().Get(name)
	if len(value) == 0 {
		value = fallback
	}
	return value
}

func main(){
	fmt.Println(chalk.Cyan, "Booting Travel Kit!", chalk.Reset)
	args := os.Args

  // Load Settings //
	filename := ""
	if len(args) > 1 {
		filename = args[1]
	} else {
		filename = "/home/vagrant/src/github.com/pjdufour/go-travel-kit/travelkit.yml"
	}

  fmt.Println(chalk.Cyan, "Loading settings...", chalk.Reset)
  settings, err := LoadSettings(filename)
	if err != nil {
		fmt.Println(chalk.Red, err, chalk.Reset)
		return
	} else {
		fmt.Println(chalk.Cyan, "Settings Loaded\n", settings, chalk.Reset)
	}

	if settings.Templates == "" {
	  setup(settings.Home)
		settings.Templates = "~/.travelkit/repos/go-travel-kit-master/templates/*"
		fmt.Println(chalk.Green, "settings.Templates set to", settings.Templates, chalk.Reset)
	}

  // Load Media //
	fmt.Println(chalk.Green, "Loading Media...", chalk.Reset)
  fmt.Println("Media Locations", settings.Media.Locations)

	err = check(settings)
	if err != nil {
		fmt.Println(chalk.Red, err, chalk.Reset)
		return
	}

  media_list, media_map, err := CollectMedia(settings.Media.Locations)
	//file_photos, err = zglob.Glob(settings.Media.Locations)
	if err != nil {
		fmt.Println(chalk.Red, err, chalk.Reset)
		return
	}
  //fmt.Println(media_list)


  // Load Templates //
	templates_list, _, err := CollectFiles(settings.Templates)
	if err != nil {
		fmt.Println(chalk.Red, err, chalk.Reset)
		return
	}
	tmpl, err := template.ParseFiles(templates_list...)

  router := httptreemux.New()

	router.GET("/", func(w http.ResponseWriter, r *http.Request, params map[string]string){
		err = tmpl.ExecuteTemplate(w, "index.html", struct{Site types.Site}{settings.Site,})
	});

	router.GET("/media", func(w http.ResponseWriter, r *http.Request, params map[string]string){

    typeOfMedia := param(r, "type", "all")
		order := param(r, "order", "most_recent")
		text := param(r, "text", "")

    CountYears := map[int]int{}
		for _, x := range media.FilterMedia(media_list, "all", 0, "", 0, 0, "most_recent") {
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

		orders := []map[string]string{}

		x := map[string]string{
			"id": "most_recent",
			"title": "Most Recent",
			"url": "/media?type="+typeOfMedia+"&order=most_recent&text="+text,
			"class": "dropdown-item",
		}
		if x["id"] == order { x["class"] = x["class"] + " disabled"; x["url"] = "#";}
		orders = append(orders, x)
		//
		x = map[string]string{
			"id": "least_recent",
			"title": "Least Recent",
			"url": "/media?type="+typeOfMedia+"&order=least_recent&text="+text,
			"class": "dropdown-item",
		}
		if x["id"] == order { x["class"] = x["class"] + " disabled"; x["url"] = "#";}
		orders = append(orders, x)
		//
		x = map[string]string{
			"id": "a_z",
			"title": "A - Z",
			"url": "/media?type="+typeOfMedia+"&order=a_z&text="+text,
			"class": "dropdown-item",
		}
		if x["id"] == order { x["class"] = x["class"] + " disabled"; x["url"] = "#";}
		orders = append(orders, x)
		//
		x = map[string]string{
			"id": "z_a",
			"title": "Z - A",
			"url": "/media?type="+typeOfMedia+"&order=z_a&text="+text,
			"class": "dropdown-item",
		}
		if x["id"] == order { x["class"] = x["class"] + " disabled"; x["url"] = "#";}
		orders = append(orders, x)
		//
		ctx := struct{
			Site types.Site;
			TypeOfMedia string;
			All bool;
			Images bool;
			Videos bool;
			Years []map[string]string;
			CountAll int;
			CountImages int;
			CountVideos int;
			MediaAll []types.MediaAttributes;
			Media7Days []types.MediaAttributes;
			Media30Days []types.MediaAttributes;
			Media90Days []types.MediaAttributes;
			Orders []map[string]string;
			Query map[string]string;
		}{
		  settings.Site,
		  typeOfMedia,
		  typeOfMedia == "all",
			typeOfMedia == "image",
			typeOfMedia == "video",
			years,
			len(media.FilterMedia(media_list, "all", 0, "", 0, 0, order)),
			len(media.FilterMedia(media_list, "image", 0, "", 0, 0, order)),
			len(media.FilterMedia(media_list, "video", 0, "", 0, 0, order)),
      media.FilterMedia(media_list, typeOfMedia, 0, text, settings.Media.Page_Size, 0, order),
			media.FilterMedia(media_list, typeOfMedia, 7, text, settings.Media.Page_Size, 0, order),
			media.FilterMedia(media_list, typeOfMedia, 30, text, settings.Media.Page_Size, 0, order),
			media.FilterMedia(media_list, typeOfMedia, 90, text, settings.Media.Page_Size, 0, order),
			orders,
			map[string]string{"Text": text},
    }
		err = tmpl.ExecuteTemplate(w, "media.html", ctx)
	});

	router.GET("/media/view/:id", func(w http.ResponseWriter, r *http.Request, params map[string]string){
			id := params["id"]
			image := media_map[id]
			ctx := struct{Title string; URI string; IsImage bool; IsVideo bool; Width int; Height int; Rotation int}{
				id,
				settings.Site.Url+"/api/media/download/"+id,
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

			data := media.FilterMedia(media_list, typeOfMedia, days, "", settings.Media.Page_Size, pageNumber, "most_recent")
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

	group.GET("/media/thumbnail/:id", func(w http.ResponseWriter, r *http.Request, params map[string]string){
	    id := params["id"]
			_, ext := media.ParseFilename(id)

			f, err := os.Open(media_map[id].Path)
			defer f.Close()

			if err != nil {
				log.Println(err) // perhaps handle this nicer
				fmt.Fprintf(w, "Could not open media file at id /%s", id)
				return
			}

			if ext == "jpg" || ext == "jpeg" {
				original, err := jpeg.Decode(f)
		    if err != nil {
		        log.Fatal(err)
		    }

        thumbnail := resize.Resize(220, 0, original, resize.Lanczos3)
				buf := new(bytes.Buffer)
				err = jpeg.Encode(buf, thumbnail, nil)
				w.Header().Set("Content-type", "image/jpeg")
				//w.Header().Set("Content-Disposition", "attachment; filename="+id )
				w.Write(buf.Bytes())
			} else {
				data, err := ioutil.ReadAll(f)
				if err != nil {
					log.Println(err) // perhaps handle this nicer
					fmt.Fprintf(w, "Could not read media file at id /%s", id)
					return
				}
				w.Header().Set("Content-Disposition", "attachment; filename="+id )
				w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
				w.Write(data)
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

  //SITE_URL := extract.Extract("http.SITE_URL", f.Root, "").(string)
	u, err := url.Parse(settings.Site.Url)
	_, port, _ := net.SplitHostPort(u.Host)
	fmt.Println(chalk.Cyan, "Listening on port", port, chalk.Reset)
  log.Fatal(http.ListenAndServe(":"+port, router))
}
