package utils

import (
	"crypto/md5"
	"fmt"
	"os"
	"strings"
)

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		return !os.IsNotExist(err), err
	}
	return true, nil
}

func IsDir(path string) (bool, error) {
	s, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return s.IsDir(), nil
}

func MD5(raw string) string {
	bs := md5.Sum([]byte(raw))
	return fmt.Sprintf("%x", bs)
}

func Pop(slice []string, index int) ([]string, string) {
	if index < 0 || index >= len(slice) {
		return slice, ""
	}
	popped := slice[index]
	return append(slice[:index], slice[index+1:]...), popped
}

func StringInSlice(str string, list []string) bool {
	for _, v := range list {
		if strings.EqualFold(v, str) { // 忽略大小写判断是否相等
			return true
		}
	}
	return false
}

func CreateMap(list []string) map[string]bool {
	result := make(map[string]bool)
	for _, v := range list {
		result[v] = true
	}
	return result
}

func StringInMap(str string, m map[string]bool) bool {
	_, ok := m[str]
	return ok
}
