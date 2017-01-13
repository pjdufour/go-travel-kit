package travelkit

import (
  "time"
  "reflect"
  "errors"
)

func firstItem(list []string) string {
  if len(list) > 0 {
    return list[0]
  } else {
    return ""
  }
}

func isLast(v interface{}, i int) (bool, error) {
  rv := reflect.ValueOf(v)
  if rv.Kind() != reflect.Slice {
    return false, errors.New("not a slice")
  }
  return rv.Len()-1 == i, nil
}
func isNotLast(v interface{}, i int) (bool, error) {
  rv := reflect.ValueOf(v)
  if rv.Kind() != reflect.Slice {
    return false, errors.New("not a slice")
  }
  return rv.Len()-1 != i, nil
}


func formatDate(t time.Time) string {
  return t.Format("January 02, 2006")
}

func formatTime(t time.Time) string {
  return t.Format("03:04:05 PM")
}
