package liveview_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/arturoeanton/go-echo-live-view/liveview"
)

func TestFileToString(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := "This is test content\nWith multiple lines"

	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test reading existing file
	content, err := liveview.FileToString(testFile)
	if err != nil {
		t.Fatalf("FileToString returned error: %v", err)
	}

	if content != testContent {
		t.Errorf("Expected content '%s', got '%s'", testContent, content)
	}

	// Test reading non-existent file
	_, err = liveview.FileToString("/non/existent/file.txt")
	if err == nil {
		t.Error("FileToString should return error for non-existent file")
	}
}

func TestStringToFile(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "output.txt")
	testContent := "This is test content for writing"

	// Test writing to file
	err := liveview.StringToFile(testFile, testContent)
	if err != nil {
		t.Fatalf("StringToFile returned error: %v", err)
	}

	// Verify the file was created and has correct content
	if !liveview.Exists(testFile) {
		t.Error("File should exist after StringToFile")
	}

	// Read back the content
	content, err := liveview.FileToString(testFile)
	if err != nil {
		t.Fatalf("Failed to read back written file: %v", err)
	}

	if content != testContent {
		t.Errorf("Written content doesn't match. Expected '%s', got '%s'", testContent, content)
	}
}

func TestExists(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	// File should not exist initially
	if liveview.Exists(testFile) {
		t.Error("File should not exist initially")
	}

	// Create the file
	err := os.WriteFile(testFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// File should exist now
	if !liveview.Exists(testFile) {
		t.Error("File should exist after creation")
	}

	// Test with directory
	if !liveview.Exists(tmpDir) {
		t.Error("Directory should exist")
	}

	// Test with non-existent path
	if liveview.Exists("/non/existent/path") {
		t.Error("Non-existent path should return false")
	}
}

func TestContainsString(t *testing.T) {
	testSlice := []string{"apple", "banana", "cherry", "date"}

	// Test existing elements
	if !liveview.ContainsString(testSlice, "apple") {
		t.Error("Should find 'apple' in slice")
	}

	if !liveview.ContainsString(testSlice, "cherry") {
		t.Error("Should find 'cherry' in slice")
	}

	// Test non-existing element
	if liveview.ContainsString(testSlice, "grape") {
		t.Error("Should not find 'grape' in slice")
	}

	// Test empty slice
	emptySlice := []string{}
	if liveview.ContainsString(emptySlice, "anything") {
		t.Error("Should not find anything in empty slice")
	}

	// Test with empty string
	sliceWithEmpty := []string{"", "test"}
	if !liveview.ContainsString(sliceWithEmpty, "") {
		t.Error("Should find empty string in slice")
	}
}

func TestFileOperationsWithEmptyPaths(t *testing.T) {
	// Test FileToString with empty path
	_, err := liveview.FileToString("")
	if err == nil {
		t.Error("FileToString with empty path should return error")
	}

	// Test StringToFile with empty path
	err = liveview.StringToFile("", "content")
	if err == nil {
		t.Error("StringToFile with empty path should return error")
	}

	// Test Exists with empty path
	if liveview.Exists("") {
		t.Error("Exists with empty path should return false")
	}
}

func TestFilePermissions(t *testing.T) {
	// Create a file with restricted permissions
	tmpDir := t.TempDir()
	restrictedFile := filepath.Join(tmpDir, "restricted.txt")
	
	err := os.WriteFile(restrictedFile, []byte("restricted"), 0000)
	if err != nil {
		t.Fatalf("Failed to create restricted file: %v", err)
	}

	// Try to read the restricted file
	_, err = liveview.FileToString(restrictedFile)
	// On some systems this might still work, on others it might fail
	// We just check it doesn't panic
	t.Logf("Reading restricted file error (expected): %v", err)

	// Clean up
	os.Chmod(restrictedFile, 0644) // Change permissions so cleanup can delete it
}

func TestLargeFileHandling(t *testing.T) {
	// Create a moderately large test file (1KB to keep test fast)
	tmpDir := t.TempDir()
	largeFile := filepath.Join(tmpDir, "large.txt")
	
	// Create 1KB of content
	largeContent := make([]byte, 1024)
	for i := range largeContent {
		largeContent[i] = byte('A' + (i % 26))
	}

	err := liveview.StringToFile(largeFile, string(largeContent))
	if err != nil {
		t.Fatalf("Failed to create large file: %v", err)
	}

	// Test reading large file
	content, err := liveview.FileToString(largeFile)
	if err != nil {
		t.Fatalf("Failed to read large file: %v", err)
	}

	if len(content) != len(largeContent) {
		t.Errorf("Expected content length %d, got %d", len(largeContent), len(content))
	}
}

func TestFileWithSpecialCharacters(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Test content with special characters
	specialContent := `Special chars: < > & " ' 
	Unicode: 你好世界 🚀 
	Escapes: \n \t \r`

	testFile := filepath.Join(tmpDir, "special.txt")
	err := liveview.StringToFile(testFile, specialContent)
	if err != nil {
		t.Fatalf("Failed to create file with special content: %v", err)
	}

	content, err := liveview.FileToString(testFile)
	if err != nil {
		t.Fatalf("Failed to read file with special content: %v", err)
	}

	if content != specialContent {
		t.Error("Special characters not preserved correctly")
	}
}

func TestRoundTripWriteRead(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "roundtrip.txt")
	
	testCases := []string{
		"Simple text",
		"Text with\nnewlines\nand\ttabs",
		"",
		"Unicode: 🚀🌟💫",
		"JSON-like: {\"key\": \"value\", \"number\": 42}",
	}

	for i, testContent := range testCases {
		t.Run(fmt.Sprintf("RoundTrip_%d", i), func(t *testing.T) {
			// Write
			err := liveview.StringToFile(testFile, testContent)
			if err != nil {
				t.Fatalf("Failed to write: %v", err)
			}

			// Read
			readContent, err := liveview.FileToString(testFile)
			if err != nil {
				t.Fatalf("Failed to read: %v", err)
			}

			// Compare
			if readContent != testContent {
				t.Errorf("Round trip failed. Original: %q, Read: %q", testContent, readContent)
			}
		})
	}
}