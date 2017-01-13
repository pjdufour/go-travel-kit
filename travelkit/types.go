package travelkit

import (
	"time"
	"strings"
)

type Contact struct {
	Id string `json:"id"`
	FullName string `json:"full_name"`
	GivenName string `json:"given_name"`
	FamilyName string `json:"family_name"`
	Numbers []string `json:"numbers"`
	Faxes []string `json:"faxes"`
	Emails []string `json:"emails"`
	Photo string `json:"-"`
	HasPhoto bool `json:"has_photo"`
}

type MediaAttributes struct{
	Id string `json:"id"`
	Path string
	TypeOfMedia MediaType `json:"type"`
	IsText bool
	IsImage bool
	IsVideo bool
	IsGeoJSON bool
	Rotation int `json:"rotation"`
	Date time.Time `json:"date"`
	Width int `json:"height"`
	Height int `json:"width"`
}

type MediaAttributesByMostRecent []MediaAttributes
func (s MediaAttributesByMostRecent) Len() int { return len(s); }
func (s MediaAttributesByMostRecent) Swap(i, j int) { s[i], s[j] = s[j], s[i]; }
func (s MediaAttributesByMostRecent) Less(i, j int) bool { return s[i].Date.After(s[j].Date); }

type MediaAttributesByLeastRecent []MediaAttributes
func (s MediaAttributesByLeastRecent) Len() int { return len(s); }
func (s MediaAttributesByLeastRecent) Swap(i, j int) { s[i], s[j] = s[j], s[i]; }
func (s MediaAttributesByLeastRecent) Less(i, j int) bool { return s[i].Date.Before(s[j].Date); }

type MediaAttributesByAZ []MediaAttributes
func (s MediaAttributesByAZ) Len() int { return len(s); }
func (s MediaAttributesByAZ) Swap(i, j int) { s[i], s[j] = s[j], s[i]; }
func (s MediaAttributesByAZ) Less(i, j int) bool { return strings.Compare(s[i].Id, s[j].Id) < 0; }

type MediaAttributesByZA []MediaAttributes
func (s MediaAttributesByZA) Len() int { return len(s); }
func (s MediaAttributesByZA) Swap(i, j int) { s[i], s[j] = s[j], s[i]; }
func (s MediaAttributesByZA) Less(i, j int) bool { return strings.Compare(s[j].Id, s[i].Id) < 0; }

type ContactsByAZ []Contact
func (s ContactsByAZ) Len() int { return len(s); }
func (s ContactsByAZ) Swap(i, j int) { s[i], s[j] = s[j], s[i]; }
func (s ContactsByAZ) Less(i, j int) bool { return strings.Compare(s[i].FullName, s[j].FullName) < 0; }

type ContactsByZA []Contact
func (s ContactsByZA) Len() int { return len(s); }
func (s ContactsByZA) Swap(i, j int) { s[i], s[j] = s[j], s[i]; }
func (s ContactsByZA) Less(i, j int) bool { return strings.Compare(s[j].FullName, s[i].FullName) < 0; }


type Settings struct {
	Home string
	Site Site
	Media Media
	//Templates string  No longer necessary, using go.rice box
}

type Site struct {
	Name string
	Url string
}

type MediaType struct {
	Id string
	Title string
	Extensions []string
}

type Media struct {
	Types []MediaType
	Page_Size int
	Locations []string
}

type Page struct {
	Site Site
}
