package types

import (
	"time"
	"strings"
)

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

type Settings struct {
	Home string
	Site Site
	Media Media
	Templates string
}

type Site struct {
	Name string
	Url string
}

type Media struct {
	Page_Size int
	Locations []string
}

type Page struct {
	Site Site
}
