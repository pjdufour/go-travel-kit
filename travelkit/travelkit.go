package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"io/ioutil"
	"net/http"
	"html/template"
	//"strings"
	//"time"
	//"strconv"
	//"path/filepath"
)

import (
	"github.com/mattn/go-zglob"
  "github.com/pjdufour/go-gypsy/yaml"
  "github.com/pjdufour/go-extract/extract"
  "github.com/dimfeld/httptreemux"
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

  path_photos := extract.Extract("photos", f.Root, "").(string)
	siteurl := extract.Extract("http.siteurl", f.Root, "").(string)
	path_templates := extract.Extract("http.templates", f.Root, "").(string)
  fmt.Println("Photos", path_photos)

	if path_photos == "" {
		return
	}

  photos_list, photos_map, photos_err := Collect(path_photos)
	//file_photos, err = zglob.Glob(path_photos)
	if photos_err != nil {
		fmt.Println(photos_err)
		return
	}

  fmt.Println(photos_list)

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

  group := router.NewGroup("/api")

  group.GET("/media/download/:id", func(w http.ResponseWriter, r *http.Request, params map[string]string){
	    id := params["id"]

			img, err := os.Open(photos_map[id])
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
			w.Write(data)

	})

	group.GET("/media/view/:id", func(w http.ResponseWriter, r *http.Request, params map[string]string){
			id := params["id"]
      image := siteurl+"/api/media/download/"+id
			//p := &Page{Title: id, Image: image}
			//renderTemplate(w, "templates/view", p)
			tmpl, err := template.ParseFiles(path_templates+"/view.html")
			if err != nil {
				log.Println(err)
				return
			}
			err = tmpl.Execute(w, struct{Title string; Image string}{id, image,})

	})

  log.Fatal(http.ListenAndServe(":8080", router))
}
