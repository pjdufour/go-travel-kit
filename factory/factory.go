package factory

import (
	"strconv"
)

func Orders(typeOfMedia string, text string, currentOrder string) ([]map[string]string) {

	list := []map[string]string{}

	x := map[string]string{
	  "id": "most_recent",
	  "title": "Most Recent",
	  "url": "/media?type="+typeOfMedia+"&order=most_recent&text="+text,
	  "class": "dropdown-item",
	}
	if x["id"] == currentOrder { x["class"] = x["class"] + " disabled"; x["url"] = "#";}
	list = append(list, x)
	//
	x = map[string]string{
	  "id": "least_recent",
	  "title": "Least Recent",
	  "url": "/media?type="+typeOfMedia+"&order=least_recent&text="+text,
	  "class": "dropdown-item",
	}
	if x["id"] == currentOrder { x["class"] = x["class"] + " disabled"; x["url"] = "#";}
	list = append(list, x)
	//
	x = map[string]string{
	  "id": "a_z",
	  "title": "A - Z",
	  "url": "/media?type="+typeOfMedia+"&order=a_z&text="+text,
	  "class": "dropdown-item",
	}
	if x["id"] == currentOrder { x["class"] = x["class"] + " disabled"; x["url"] = "#";}
	list = append(list, x)
	//
	x = map[string]string{
	  "id": "z_a",
	  "title": "Z - A",
	  "url": "/media?type="+typeOfMedia+"&order=z_a&text="+text,
	  "class": "dropdown-item",
	}
	if x["id"] == currentOrder { x["class"] = x["class"] + " disabled"; x["url"] = "#";}
	list = append(list, x)

	return list
}

func Types(typeOfMedia string, text string, currentOrder string, countsByType map[string]int) ([]map[string]string) {

	list := []map[string]string{}
	list = append(list, map[string]string{
		"id": "all",
		"title": "All",
		"class": "list-group-item list-group-item-action justify-content-between",
	})
	list = append(list, map[string]string{
		"id": "image",
		"title": "Image",
		"class": "list-group-item list-group-item-action justify-content-between",
	})
	list = append(list, map[string]string{
		"id": "video",
		"title": "video",
		"class": "list-group-item list-group-item-action justify-content-between",
	})

	for _, x := range list {
		x["url"] = "/media?type="+x["id"]+"&order="+currentOrder+"&text="+text
		x["count"] = strconv.Itoa(countsByType[x["id"]])
	  if x["id"] == typeOfMedia {
		  x["class"] = x["class"] + " active";
		}
	}

  return list
}
