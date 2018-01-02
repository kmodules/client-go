package meta

import (
	"encoding/json"
	"strconv"

	"github.com/appscode/kutil"
)

func GetBool(m map[string]string, key string) (bool, error) {
	if m == nil {
		return false, kutil.ErrNotFound
	}
	return strconv.ParseBool(m[key])
}

func GetInt(m map[string]string, key string) (int, error) {
	if m == nil {
		return 0, kutil.ErrNotFound
	}
	s, ok := m[key]
	if !ok {
		return 0, kutil.ErrNotFound
	}
	return strconv.Atoi(s)
}

func GetString(m map[string]string, key string) string {
	if m == nil {
		return ""
	}
	return m[key]
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

func GetList(m map[string]string, key string) ([]string, error) {
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

func GetMap(m map[string]string, key string) (map[string]string, error) {
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
