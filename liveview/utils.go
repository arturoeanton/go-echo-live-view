package liveview

import (
	"fmt"
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
	// SEC-004: Validar path traversal
	if err := ValidatePath(name); err != nil {
		return "", fmt.Errorf("invalid file path: %w", err)
	}
	
	content, err := os.ReadFile(name)
	return string(content), err
}

func StringToFile(filename string, content string) error {
	// SEC-004: Validar path traversal
	if err := ValidatePath(filename); err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}
	
	d1 := []byte(content)
	err := os.WriteFile(filename, d1, 0644)
	return err
}
