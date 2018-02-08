package meta

import (
	"encoding/json"
	"strconv"

	"github.com/appscode/kutil"
)

type ParserFunc func(map[string]string, string) (interface{}, error)

var _ ParserFunc = GetBool
var _ ParserFunc = GetInt
var _ ParserFunc = GetString
var _ ParserFunc = GetList
var _ ParserFunc = GetMap

func GetBool(m map[string]string, key string) (interface{}, error) {
	if m == nil {
		return false, kutil.ErrNotFound
	}
	return strconv.ParseBool(m[key])
}

func GetBoolValue(m map[string]string, key string) bool {
	v, _ := GetBool(m, key)
	return v.(bool)
}

func GetInt(m map[string]string, key string) (interface{}, error) {
	if m == nil {
		return 0, kutil.ErrNotFound
	}
	v, ok := m[key]
	if !ok {
		return 0, kutil.ErrNotFound
	}
	return strconv.Atoi(v)
}

func GetIntValue(m map[string]string, key string) int {
	v, _ := GetInt(m, key)
	return v.(int)
}

func GetString(m map[string]string, key string) (interface{}, error) {
	if m == nil {
		return "", kutil.ErrNotFound
	}
	v, ok := m[key]
	if !ok {
		return "", kutil.ErrNotFound
	}
	return v, nil
}

func GetStringValue(m map[string]string, key string) string {
	v, _ := GetInt(m, key)
	return v.(string)
}

func HasKey(m map[string]string, key string) bool {
	if m == nil {
		return false
	}
	_, ok := m[key]
	return ok
}

func RemoveKey(m map[string]string, key string) map[string]string {
	if m == nil {
		return nil
	}
	delete(m, key)
	return m
}

func GetList(m map[string]string, key string) (interface{}, error) {
	if m == nil {
		return []string{}, kutil.ErrNotFound
	}
	s, ok := m[key]
	if !ok {
		return []string{}, kutil.ErrNotFound
	}
	v := make([]string, 0)
	err := json.Unmarshal([]byte(s), &v)
	return v, err
}

func GetListValue(m map[string]string, key string) []string {
	v, _ := GetList(m, key)
	return v.([]string)
}

func GetMap(m map[string]string, key string) (interface{}, error) {
	if m == nil {
		return map[string]string{}, kutil.ErrNotFound
	}
	s, ok := m[key]
	if !ok {
		return map[string]string{}, kutil.ErrNotFound
	}
	v := make(map[string]string)
	err := json.Unmarshal([]byte(s), &v)
	return v, err
}

func GetMapValue(m map[string]string, key string) map[string]string {
	v, _ := GetMap(m, key)
	return v.(map[string]string)
}
