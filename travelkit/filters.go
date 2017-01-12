package travelkit

import (
  "time"
)

func firstItem(list []string) string {
  if len(list) > 0 {
    return list[0]
  } else {
    return ""
  }
}

func formatDate(t time.Time) string {
  return t.Format("January 02, 2006")
}

func formatTime(t time.Time) string {
  return t.Format("03:04:05 PM")
}
