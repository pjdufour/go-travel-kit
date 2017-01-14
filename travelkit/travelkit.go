package travelkit

import (
	"errors"
	"encoding/base64"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/user"
	"path"
	//"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	//"math"
	//"html/template"
	"strings"
	"time"
	"strconv"
	"reflect"
	//"path/filepath"
	"image/jpeg"
)

import (
	"github.com/ttacon/chalk"
	"github.com/mattn/go-zglob"
  "github.com/pjdufour/go-gypsy/yaml"
  //"github.com/pjdufour/go-extract/extract"
  "github.com/dimfeld/httptreemux"
	"github.com/nfnt/resize"
	//"github.com/rwcarlsen/goexif/exif"
	"github.com/patrickmn/go-cache"
	"github.com/GeertJohan/go.rice"
)

//import (
//	"github.com/pjdufour/go-travel-kit/unzip"
//)

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

func ConvertYAMLListToStringList(list yaml.List) []string {
	out := make([](string), list.Len())
	for index, _ := range list {
		y := list.Item(index)
		if reflect.TypeOf(y).String() == "yaml.Scalar" {
			out[index] = Trim(y.(yaml.Scalar).String())
		}
	}
	return out
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
		fmt.Println(chalk.Red, "Could Not Collect files at ", dir, chalk.Reset)
		fmt.Println(chalk.Red, err, chalk.Reset)
		return nil , nil , err
	}

	if len(files) == 0 {
		return nil, nil, errors.New("No files found at '"+ dir+"'.")
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

/*
 * No longer necessary, b/c can embed resources using go-rice https://github.com/GeertJohan/go.rice
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
}*/

func TravelKit(){

	fmt.Println(chalk.Cyan, "Booting Travel Kit!", chalk.Reset)
	args := os.Args

  // Load Settings //
	filename := ""
	if len(args) > 1 {
		filename = args[1]
	} else {
		filename = "~/travelkit.yml"
	}

  fmt.Println(chalk.Cyan, "Loading settings...", chalk.Reset)
  s := LoadSettings(filename)
	fmt.Println(chalk.Cyan, "Settings Loaded\n", s, chalk.Reset)

  /* No longer necessary b/c using go rice box
	if s.Templates == "" {
	  setup(s.Home)
		s.Templates = "~/.travelkit/repos/go-travel-kit-master/templates/*"
		fmt.Println(chalk.Green, "s.Templates set to", s.Templates, chalk.Reset)
	}*/

	err := check(s)
	if err != nil {
		fmt.Println(chalk.Red, err, chalk.Reset)
		return
	}

  // Load Media //
	fmt.Println(chalk.Cyan, "Loading Media...", chalk.Reset)
  fmt.Println(chalk.Cyan, "Media Locations: ", s.Media.Locations, chalk.Reset)
  media_list, media_map, err := CollectMedia(s, s.Media.Locations)
	//file_photos, err = zglob.Glob(settings.Media.Locations)
	if err != nil {
		fmt.Println(chalk.Red, err, chalk.Reset)
		return
	}

	// Load Contacts //
	fmt.Println(chalk.Cyan, "Loading Contacts...", chalk.Reset)
	contacts_list, contacts_map, err := CollectContacts(s, media_list, media_map)
	if err != nil {
		fmt.Println(chalk.Red, err, chalk.Reset)
		return
	}

	// Load Contacts //
	fmt.Println(chalk.Cyan, "Loading emails...", chalk.Reset)
	emails_list, emails_map, err := CollectEmails(s, media_list, media_map)
	if err != nil {
		fmt.Println(chalk.Red, err, chalk.Reset)
		return
	}
	fmt.Println(chalk.Cyan, "Emails List:", emails_list, chalk.Reset)

  fmt.Println(chalk.Cyan, "Loading templates", chalk.Reset)
  tmpl, err := LoadTemplatesFromBinary(s)
	if err != nil {
		fmt.Println(chalk.Red, err, chalk.Reset)
		return
	} else {
		fmt.Println(chalk.Cyan, "Templates Loaded\n", tmpl, chalk.Reset)
	}

	thumbnails := cache.New(5*time.Minute, 30*time.Second)

  router := httptreemux.New()

	router.GET("/", func(w http.ResponseWriter, r *http.Request, params map[string]string){
		ctx := struct{
			Site Site;
			Query map[string]string;
		}{
			s.Site,
			map[string]string{"Text": ""},
		}
		fmt.Println(chalk.Cyan, "About to execue index template", chalk.Reset)
		fmt.Println(tmpl)
		err = tmpl.ExecuteTemplate(w, "index.html", ctx)
	});


	staticBox, err := rice.FindBox("static")
	if err != nil {
		fmt.Println(chalk.Cyan, "Error: Could not find Rice Box static.", chalk.Reset)
		return
	}
	fmt.Println(chalk.Cyan, "Static Box", staticBox, chalk.Reset)

	router.GET("/static", func(w http.ResponseWriter, r *http.Request, params map[string]string){
		pathtofile := param(r, params, "id", "")
		if len(pathtofile) == 0 {
			fmt.Println(chalk.Cyan, "Error: Not path for static file found.", chalk.Reset)
			return
		}

		content, err := staticBox.String(pathtofile)

		if err != nil {
			fmt.Println(chalk.Cyan, "Error: Could read static file at path", pathtofile, ".", chalk.Reset)
			return
		}

    content = strings.Replace(content, "%", "%%", -1)

		contentType := fileExtensionToContentType(pathtofile)
		w.Header().Set("Content-Type", contentType)

		fmt.Fprintf(w, content)

	});

	router.GET("/about", func(w http.ResponseWriter, r *http.Request, params map[string]string){
		ctx := struct{
			Home string;
			Site Site;
			Media Media;
			Query map[string]string;
		}{
			s.Home,
			s.Site,
			s.Media,
			map[string]string{"Text": ""},
		}
		err = tmpl.ExecuteTemplate(w, "about.html", ctx)
	});

	router.GET("/contacts", func(w http.ResponseWriter, r *http.Request, params map[string]string){

		order := param(r, params, "order", "most_recent")
		text := param(r, params, "text", "")

		ctx := struct{
			Site Site;
			ContactsAll []Contact;
			Orders []map[string]string;
			Query map[string]string;
		}{
		  s.Site,
			FilterContacts(contacts_list, text, s.Media.Page_Size, 0, order),
			CreateOrdersForContacts(text, order),
			map[string]string{"Text": text, "Order": order},
    }
		err = tmpl.ExecuteTemplate(w, "contacts.html", ctx)
	});

	router.GET("/contacts/view", func(w http.ResponseWriter, r *http.Request, params map[string]string){
		id := param(r, params, "id", "")
		item := contacts_map[id]

		ctx := struct{
			Site Site;
			Query map[string]string;
			Contact Contact;
		}{
			s.Site,
			map[string]string{"Text": id},
			item,
		}
		err = tmpl.ExecuteTemplate(w, "contacts_view.html", ctx)
	})

	router.GET("/emails", func(w http.ResponseWriter, r *http.Request, params map[string]string){

		order := param(r, params, "order", "most_recent")
		text := param(r, params, "text", "")

		ctx := struct{
			Site Site;
			EmailsAll []Email;
			Emails7Days []Email;
			Emails30Days []Email;
			Emails90Days []Email;
			Emails180Days []Email;
			Orders []map[string]string;
			Query map[string]string;
		}{
		  s.Site,
			FilterEmails(emails_list, 0, text, s.Media.Page_Size, 0, order),
			FilterEmails(emails_list, 7, text, s.Media.Page_Size, 0, order),
			FilterEmails(emails_list, 30, text, s.Media.Page_Size, 0, order),
			FilterEmails(emails_list, 90, text, s.Media.Page_Size, 0, order),
			FilterEmails(emails_list, 180, text, s.Media.Page_Size, 0, order),
			CreateOrdersForEmails(text, order),
			map[string]string{"Text": text, "Order": order},
    }
		err = tmpl.ExecuteTemplate(w, "emails.html", ctx)
	});

	router.GET("/emails/view", func(w http.ResponseWriter, r *http.Request, params map[string]string){
		id := param(r, params, "id", "")
		item := emails_map[id]

		ctx := struct{
			Site Site;
			Query map[string]string;
			Email Email;
		}{
			s.Site,
			map[string]string{"Text": id},
			item,
		}
		err = tmpl.ExecuteTemplate(w, "emails_view.html", ctx)
	})

	router.GET("/media", func(w http.ResponseWriter, r *http.Request, params map[string]string){

    typeOfMedia := param(r, params, "type", "all")
		order := param(r, params, "order", "most_recent")
		text := param(r, params, "text", "")
    countsByType := buildCountsByType(s, media_list, order)

		ctx := struct{
			Site Site;
			TypeOfMedia string;
			All bool;
			Images bool;
			Videos bool;
			Years []map[string]string;
			MediaAll []MediaAttributes;
			Media7Days []MediaAttributes;
			Media30Days []MediaAttributes;
			Media90Days []MediaAttributes;
			Media180Days []MediaAttributes;
			Types []map[string]string;
			Orders []map[string]string;
			Query map[string]string;
			CountsByType map[string]string;
		}{
		  s.Site,
		  typeOfMedia,
		  typeOfMedia == "all",
			typeOfMedia == "image",
			typeOfMedia == "video",
			buildYears(s, media_list),
      FilterMedia(media_list, typeOfMedia, 0, text, s.Media.Page_Size, 0, order),
			FilterMedia(media_list, typeOfMedia, 7, text, s.Media.Page_Size, 0, order),
			FilterMedia(media_list, typeOfMedia, 30, text, s.Media.Page_Size, 0, order),
			FilterMedia(media_list, typeOfMedia, 90, text, s.Media.Page_Size, 0, order),
			FilterMedia(media_list, typeOfMedia, 180, text, s.Media.Page_Size, 0, order),
			CreateTypes(s, typeOfMedia, text, order, countsByType),
			CreateOrdersForMedia(typeOfMedia, text, order),
			map[string]string{"Text": text, "Order": order},
			Stringify(countsByType),
    }
		err = tmpl.ExecuteTemplate(w, "media.html", ctx)
	});

	router.GET("/media/view", func(w http.ResponseWriter, r *http.Request, params map[string]string){
		id := param(r, params, "id", "")
		item := media_map[id]
		textContent := ""

    if item.TypeOfMedia.Id == "text" {
			b, err := ioutil.ReadFile(item.Path)
			if err != nil {
				msg := "Could not open text file with id "+id+"."
				fmt.Println(chalk.Cyan, msg, chalk.Reset)
				fmt.Fprintf(w, msg)
			}
			textContent = string(b)
		}


		uri := s.Site.Url+"/api/media/download/"+id
		if item.TypeOfMedia.Id == "geojson" {
			uri = s.Site.Url+"/api/media/geojson/"+id
		}

		ctx := struct{
			Site Site;
			Query map[string]string;
			Id string;
			Title string;
			URI string;
			IsText bool;
			IsImage bool;
			IsVideo bool;
			IsGeoJSON bool;
			Width int;
			Height int;
			Rotation int;
			TextContent string;
		}{
			s.Site,
			map[string]string{"Text": id},
			id,
			id,
			uri,
			item.IsText,
			item.IsImage,
			item.IsVideo,
			item.IsGeoJSON,
			item.Width,
			item.Height,
			item.Rotation,
			textContent,
		}
		err = tmpl.ExecuteTemplate(w, "media_view.html", ctx)
	})

  group := router.NewGroup("/api")

	group.GET("/contacts/list/page/:page", func(w http.ResponseWriter, r *http.Request, params map[string]string){
		text := param(r, params, "text", "")
		order := param(r, params, "order", "most_recent")

		pageNumber := 0
		if len(params["page"]) > 0 {
			pageNumber, _ = strconv.Atoi(params["page"])
		}

		ext := param(r, params, "ext", "json")

		fmt.Println("params", params)

		data := FilterContacts(contacts_list, text, s.Media.Page_Size, pageNumber, order)
		if ext == "json" {
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(data); err != nil {
				panic(err)
			}
		} else if ext == "yml" {
			w.Header().Set("Content-Type", "plain/text")
		}
	})

	group.GET("/contacts/thumbnail", func(w http.ResponseWriter, r *http.Request, params map[string]string){
	    id := param(r, params, "id", "")
			//_, ext := ParseFilename(id, true)

			contact := contacts_map[id]

      if len(contact.Photo) > 0 {
				b, err := base64.StdEncoding.DecodeString(contact.Photo)
				if err != nil {
					msg := "Error: Could not decode contact photo from base64 encoding."
					fmt.Println(chalk.Red, msg, chalk.Reset)
					w.Header().Set("Content-Type", "plain/text")
					fmt.Fprintf(w, msg)
				} else {
					w.Header().Set("Content-Disposition", "attachment; filename="+id )
					w.Header().Set("Content-Type", "image/jpeg")
					w.Write(b)
				}
			} else {
				w.Header().Set("Content-Type", "plain/text")
			}

	})

	group.GET("/contacts/download/:id", func(w http.ResponseWriter, r *http.Request, params map[string]string){
			id := param(r, params, "id", "")

			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Content-Disposition", "attachment; filename="+id )
			if err := json.NewEncoder(w).Encode(contacts_map[id]); err != nil {
				panic(err)
			}
	})

	group.GET("/media/list/type/:type/days/:days/page/:page", func(w http.ResponseWriter, r *http.Request, params map[string]string){
		typeOfMedia := param(r, params, "type", "all")
		text := param(r, params, "text", "")
		order := param(r, params, "order", "most_recent")

		days := 0
		if len(params["days"]) > 0 {
			days, _ = strconv.Atoi(params["days"])
		}

		pageNumber := 0
		if len(params["page"]) > 0 {
			pageNumber, _ = strconv.Atoi(params["page"])
		}

		ext := param(r, params, "ext", "json")

		fmt.Println("params", params)

		data := FilterMedia(media_list, typeOfMedia, days, text, s.Media.Page_Size, pageNumber, order)
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
	    id := param(r, params, "id", "")
			_, ext := ParseFilename(id, true)

			if ext == "jpg" || ext == "jpeg" {
        thumbnailAsBytes := make([]byte, 0)
				thumbnailFromCache, found := thumbnails.Get(id)
        if found {
					//fmt.Println(chalk.Cyan, "Cache hit for thumbnail", id, chalk.Reset)
					thumbnailAsBytes = thumbnailFromCache.([]byte)
        } else {
					//fmt.Println(chalk.Cyan, "Cache miss for thumbnail", id, chalk.Reset)
					f, err := os.Open(media_map[id].Path)
					defer f.Close()

					if err != nil {
						log.Println(err) // perhaps handle this nicer
						fmt.Fprintf(w, "Could not open media file at id /%s", id)
						return
					}

					original, err := jpeg.Decode(f)
			    if err != nil {
			        log.Fatal(err)
			    }

	        thumbnail := resize.Resize(220, 0, original, resize.Lanczos3)
					buf := new(bytes.Buffer)
					err = jpeg.Encode(buf, thumbnail, nil)
					thumbnailAsBytes = buf.Bytes()
					thumbnails.Set(id, thumbnailAsBytes, cache.DefaultExpiration)
				}
				w.Header().Set("Content-type", "image/jpeg")
				//w.Header().Set("Content-Disposition", "attachment; filename="+id )
				w.Write(thumbnailAsBytes)

			} else {
				f, err := os.Open(media_map[id].Path)
				defer f.Close()

				if err != nil {
					log.Println(err) // perhaps handle this nicer
					fmt.Fprintf(w, "Could not open media file at id /%s", id)
					return
				}

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

	group.GET("/media/geojson/:id", func(w http.ResponseWriter, r *http.Request, params map[string]string){
	    id := param(r, params, "id", "")
			item := media_map[id]

			if item.TypeOfMedia.Id == "geojson" {
				b, err := ioutil.ReadFile(item.Path)
				if err != nil {
					msg := "Could not open text file with id "+id+"."
					fmt.Println(chalk.Cyan, msg, chalk.Reset)
					fmt.Fprintf(w, msg)
				}
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintf(w, string(b))
			} else {
				fmt.Fprintf(w, "The media file %s is not geojson.", id)
			}
	})

  group.GET("/media/download/:id", func(w http.ResponseWriter, r *http.Request, params map[string]string){
	    id := param(r, params, "id", "")

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

			w.Header().Set("Content-Disposition", "attachment; filename="+id )
			w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
			w.Write(data)
	})

  //SITE_URL := extract.Extract("http.SITE_URL", f.Root, "").(string)
	u, err := url.Parse(s.Site.Url)
	_, port, _ := net.SplitHostPort(u.Host)
	fmt.Println(chalk.Cyan, "Listening on port", port, chalk.Reset)
  log.Fatal(http.ListenAndServe(":"+port, router))
}
