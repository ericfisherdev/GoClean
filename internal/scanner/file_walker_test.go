package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewFileWalker(t *testing.T) {
	includePaths := []string{"./src", "./internal"}
	excludePatterns := []string{"*.tmp", "vendor/"}
	fileTypes := []string{".go", ".js"}
	verbose := true

	fw := NewFileWalker(includePaths, excludePatterns, fileTypes, verbose)

	if fw == nil {
		t.Fatal("Expected file walker to be created, got nil")
	}

	if len(fw.includePaths) != 2 {
		t.Errorf("Expected 2 include paths, got %d", len(fw.includePaths))
	}

	if len(fw.excludePatterns) != 2 {
		t.Errorf("Expected 2 exclude patterns, got %d", len(fw.excludePatterns))
	}

	if len(fw.fileTypes) != 2 {
		t.Errorf("Expected 2 file types, got %d", len(fw.fileTypes))
	}

	if !fw.verbose {
		t.Error("Expected verbose to be true")
	}
}

func TestWalkEmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	fw := NewFileWalker([]string{tmpDir}, []string{}, []string{".go"}, false)

	files, err := fw.Walk()
	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	if len(files) != 0 {
		t.Errorf("Expected 0 files in empty directory, got %d", len(files))
	}
}

func TestWalkWithFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	goFile := filepath.Join(tmpDir, "test.go")
	jsFile := filepath.Join(tmpDir, "test.js")
	txtFile := filepath.Join(tmpDir, "readme.txt")
	tmpFile := filepath.Join(tmpDir, "temp.tmp")

	for _, file := range []string{goFile, jsFile, txtFile, tmpFile} {
		err := os.WriteFile(file, []byte("content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	// Test with file type filtering
	fw := NewFileWalker([]string{tmpDir}, []string{"*.tmp"}, []string{".go", ".js"}, false)

	files, err := fw.Walk()
	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	// Should find 2 files (.go and .js), excluding .txt and .tmp
	if len(files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(files))
	}

	// Verify file details
	foundGo := false
	foundJs := false
	for _, file := range files {
		if file.Extension == ".go" {
			foundGo = true
			if file.Language != "Go" {
				t.Errorf("Expected Go language for .go file, got %s", file.Language)
			}
		} else if file.Extension == ".js" {
			foundJs = true
			if file.Language != "JavaScript" {
				t.Errorf("Expected JavaScript language for .js file, got %s", file.Language)
			}
		}
	}

	if !foundGo {
		t.Error("Expected to find .go file")
	}
	if !foundJs {
		t.Error("Expected to find .js file")
	}
}

func TestWalkWithSubdirectories(t *testing.T) {
	tmpDir := t.TempDir()

	// Create subdirectories
	srcDir := filepath.Join(tmpDir, "src")
	testDir := filepath.Join(tmpDir, "tests")
	vendorDir := filepath.Join(tmpDir, "vendor")

	for _, dir := range []string{srcDir, testDir, vendorDir} {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create test files in subdirectories
	files := map[string]string{
		filepath.Join(srcDir, "main.go"):     "Go file",
		filepath.Join(testDir, "test.go"):    "Go test file",
		filepath.Join(vendorDir, "lib.go"):   "Vendor file",
		filepath.Join(tmpDir, "root.go"):     "Root file",
	}

	for path, content := range files {
		err := os.WriteFile(path, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create file %s: %v", path, err)
		}
	}

	// Test with vendor exclusion
	fw := NewFileWalker([]string{tmpDir}, []string{"vendor/"}, []string{".go"}, false)

	foundFiles, err := fw.Walk()
	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	// Should find 3 files (excluding vendor/)
	if len(foundFiles) != 3 {
		t.Errorf("Expected 3 files, got %d", len(foundFiles))
	}

	// Verify no vendor files
	for _, file := range foundFiles {
		if filepath.Dir(file.Path) == vendorDir {
			t.Errorf("Found file in excluded vendor directory: %s", file.Path)
		}
	}
}

func TestWalkNonExistentPath(t *testing.T) {
	fw := NewFileWalker([]string{"/non/existent/path"}, []string{}, []string{".go"}, false)

	files, err := fw.Walk()
	if err != nil {
		t.Fatalf("Walk should handle non-existent paths gracefully: %v", err)
	}

	if len(files) != 0 {
		t.Errorf("Expected 0 files for non-existent path, got %d", len(files))
	}
}

func TestShouldExclude(t *testing.T) {
	fw := NewFileWalker([]string{"."}, []string{"*.tmp", "vendor/", "_test.go", ".git/"}, []string{".go"}, false)

	testCases := []struct {
		path     string
		expected bool
	}{
		{"test.tmp", true},
		{"file.go", false},
		{"vendor/lib.go", true},
		{"src/vendor/lib.go", true},
		{"main_test.go", true}, // _test.go pattern should match main_test.go
		{"main.go", false},
		{".git/config", true},
		{".gitignore", false}, // .git/ pattern should not match .gitignore (more precise)
		{"temp.tmp", true},
		{"normal.js", false},
	}

	for _, tc := range testCases {
		result := fw.shouldExclude(tc.path)
		if result != tc.expected {
			t.Errorf("shouldExclude(%q) = %v, expected %v", tc.path, result, tc.expected)
		}
	}
}

func TestIsAllowedFileType(t *testing.T) {
	fw := NewFileWalker([]string{"."}, []string{}, []string{".go", ".js", ".ts"}, false)

	testCases := []struct {
		path     string
		expected bool
	}{
		{"main.go", true},
		{"script.js", true},
		{"types.ts", true},
		{"readme.txt", false},
		{"config.yaml", false},
		{"test.py", false},
		{"file.GO", false}, // Case sensitive
	}

	for _, tc := range testCases {
		result := fw.isAllowedFileType(tc.path)
		if result != tc.expected {
			t.Errorf("isAllowedFileType(%q) = %v, expected %v", tc.path, result, tc.expected)
		}
	}
}

func TestDetectLanguage(t *testing.T) {
	fw := NewFileWalker([]string{"."}, []string{}, []string{}, false)

	testCases := []struct {
		extension string
		expected  string
	}{
		{".go", "Go"},
		{".js", "JavaScript"},
		{".ts", "TypeScript"},
		{".py", "Python"},
		{".java", "Java"},
		{".cs", "C#"},
		{".cpp", "C++"},
		{".c", "C"},
		{".h", "C"},
		{".hpp", "C++"},
		{".php", "PHP"},
		{".rb", "Ruby"},
		{".rs", "Rust"},
		{".swift", "Swift"},
		{".kt", "Kotlin"},
		{".scala", "Scala"},
		{".unknown", "Unknown"},
	}

	for _, tc := range testCases {
		result := fw.detectLanguage(tc.extension)
		if result != tc.expected {
			t.Errorf("detectLanguage(%q) = %q, expected %q", tc.extension, result, tc.expected)
		}
	}
}

func TestIsSupportedLanguage(t *testing.T) {
	fw := NewFileWalker([]string{"."}, []string{}, []string{}, false)

	testCases := []struct {
		path     string
		expected bool
	}{
		{"test.go", true},
		{"script.js", true},
		{"types.ts", true},
		{"app.py", true},
		{"Main.java", true},
		{"Program.cs", true},
		{"code.cpp", true},
		{"header.h", true},
		{"readme.txt", false},
		{"config.yaml", false},
		{"data.json", false},
		{"image.png", false},
	}

	for _, tc := range testCases {
		result := fw.isSupportedLanguage(tc.path)
		if result != tc.expected {
			t.Errorf("isSupportedLanguage(%q) = %v, expected %v", tc.path, result, tc.expected)
		}
	}
}

func TestWalkWithMultiplePaths(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple directories
	dir1 := filepath.Join(tmpDir, "dir1")
	dir2 := filepath.Join(tmpDir, "dir2")

	err := os.MkdirAll(dir1, 0755)
	if err != nil {
		t.Fatalf("Failed to create dir1: %v", err)
	}

	err = os.MkdirAll(dir2, 0755)
	if err != nil {
		t.Fatalf("Failed to create dir2: %v", err)
	}

	// Create files in each directory
	file1 := filepath.Join(dir1, "file1.go")
	file2 := filepath.Join(dir2, "file2.go")

	for _, file := range []string{file1, file2} {
		err := os.WriteFile(file, []byte("package main"), 0644)
		if err != nil {
			t.Fatalf("Failed to create file %s: %v", file, err)
		}
	}

	fw := NewFileWalker([]string{dir1, dir2}, []string{}, []string{".go"}, false)

	files, err := fw.Walk()
	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	if len(files) != 2 {
		t.Errorf("Expected 2 files from multiple paths, got %d", len(files))
	}

	// Verify both files are found
	foundPaths := make(map[string]bool)
	for _, file := range files {
		foundPaths[file.Path] = true
	}

	if !foundPaths[file1] {
		t.Error("Expected to find file1.go")
	}
	if !foundPaths[file2] {
		t.Error("Expected to find file2.go")
	}
}

func TestWalkWithSymlinks(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a real file
	realFile := filepath.Join(tmpDir, "real.go")
	err := os.WriteFile(realFile, []byte("package main"), 0644)
	if err != nil {
		t.Fatalf("Failed to create real file: %v", err)
	}

	// Create a symlink (skip if not supported on this platform)
	symlink := filepath.Join(tmpDir, "link.go")
	err = os.Symlink(realFile, symlink)
	if err != nil {
		t.Skipf("Symlinks not supported: %v", err)
	}

	fw := NewFileWalker([]string{tmpDir}, []string{}, []string{".go"}, false)

	files, err := fw.Walk()
	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	// Should find both the real file and the symlink
	if len(files) < 1 {
		t.Errorf("Expected at least 1 file, got %d", len(files))
	}

	// Verify real file is found
	foundReal := false
	for _, file := range files {
		if file.Path == realFile {
			foundReal = true
			break
		}
	}

	if !foundReal {
		t.Error("Expected to find the real file")
	}
}

func TestWalkWithHiddenFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create hidden file (starting with .)
	hiddenFile := filepath.Join(tmpDir, ".hidden.go")
	normalFile := filepath.Join(tmpDir, "normal.go")

	for _, file := range []string{hiddenFile, normalFile} {
		err := os.WriteFile(file, []byte("package main"), 0644)
		if err != nil {
			t.Fatalf("Failed to create file %s: %v", file, err)
		}
	}

	fw := NewFileWalker([]string{tmpDir}, []string{}, []string{".go"}, false)

	files, err := fw.Walk()
	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	// Should find both hidden and normal files
	if len(files) != 2 {
		t.Errorf("Expected 2 files (including hidden), got %d", len(files))
	}
}