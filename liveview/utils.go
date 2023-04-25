package liveview

import (
	"io/ioutil"
	"os"
)

func ContainsString(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		return false
	}
	return true
}

func FileToString(name string) (string, error) {
	content, err := ioutil.ReadFile(name)
	return string(content), err
}

func StringToFile(filenanme string, content string) error {
	d1 := []byte(content)
	err := ioutil.WriteFile(filenanme, d1, 0644)
	return err
}
