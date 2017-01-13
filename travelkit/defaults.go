package travelkit


var DEFAULT_HOME = "~/.travelkit"

var DEFAULT_SITE_NAME = "Travel Kit"

var DEFAULT_SITE_URL = "http://localhost:8000"

var DEFAULT_MEDIA_LOCATIONS = []string{
  "~/Desktop/*",
}

var DEFAULT_PAGE_SIZE = 16

var DEFAULT_MEDIA_TYPES = []MediaType{
  MediaType{
    Id: "image",
    Title: "Image",
    Extensions: []string{"png", "jpg", "jpeg", "tif"},
  },
  MediaType{
    Id: "video",
    Title: "Video",
    Extensions: []string{"mp4",},
  },
  MediaType{
    Id: "vcf",
    Title: "VCF",
    Extensions: []string{"vcf", "vcf~",},
  },
  MediaType{
    Id: "text",
    Title: "text",
    Extensions: []string{"txt", "txt~", "yaml", "log", "html", "yml", "yaml", "json", "json~", "pem", "xml", "md",},
  },
  MediaType{
    Id: "geojson",
    Title: "GeoJSON",
    Extensions: []string{"geojson", "geojson~",},
  },
  MediaType{
    Id: "shapefile",
    Title: "Shapefile",
    Extensions: []string{"shp",},
  },
  MediaType{
    Id: "word",
    Title: "Word",
    Extensions: []string{".doc", ".docx",},
  },
  MediaType{
    Id: "pdf",
    Title: "PDF",
    Extensions: []string{".pdf",},
  },
  MediaType{
    Id: "power_point",
    Title: "Power Point",
    Extensions: []string{".ppt", "pptx",},
  },
  MediaType{
    Id: "archive",
    Title: "Archive",
    Extensions: []string{".zip", "war", "gz",},
  },
}
