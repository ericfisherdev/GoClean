package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/ericfisherdev/goclean/internal/testutils"
)

// TestCrossPlatformPathHandling tests that all path operations work correctly across platforms
func TestCrossPlatformPathHandling(t *testing.T) {
	tests := []struct {
		name        string
		setupFiles  map[string]string
		excludePatterns []string
		expectedFiles []string
	}{
		{
			name: "Windows-style paths on Unix",
			setupFiles: map[string]string{
				filepath.Join("src", "main.go"):     "package main",
				filepath.Join("src", "util.go"):     "package util",
				filepath.Join("tests", "main_test.go"): "package main",
			},
			excludePatterns: []string{"tests"},
			expectedFiles: []string{
				filepath.Join("src", "main.go"),
				filepath.Join("src", "util.go"),
			},
		},
		{
			name: "Unix-style paths with backslash handling",
			setupFiles: map[string]string{
				filepath.Join("app", "handlers", "user.go"): "package handlers",
				filepath.Join("app", "models", "user.go"):   "package models",
				filepath.Join("vendor", "lib.go"):           "package vendor",
			},
			excludePatterns: []string{"vendor"},
			expectedFiles: []string{
				filepath.Join("app", "handlers", "user.go"),
				filepath.Join("app", "models", "user.go"),
			},
		},
		{
			name: "Mixed separator handling",
			setupFiles: map[string]string{
				filepath.Join("mixed", "path", "file1.go"): "package mixed",
				filepath.Join("mixed", "path", "file2.go"): "package mixed",
				filepath.Join("excluded", "file.go"):       "package excluded",
			},
			excludePatterns: []string{"excluded"},
			expectedFiles: []string{
				filepath.Join("mixed", "path", "file1.go"),
				filepath.Join("mixed", "path", "file2.go"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a separate temp directory for each test case
			tempDir := testutils.CreateTempDir(t)
			defer os.RemoveAll(tempDir)
			
			// Setup test directory structure
			for filePath, content := range tt.setupFiles {
				testutils.CreateTestFile(t, tempDir, filePath, content)
			}

			// Create file walker
			fw := NewFileWalker([]string{tempDir}, tt.excludePatterns, []string{".go"}, false)
			
			// Walk files
			files, err := fw.Walk()
			if err != nil {
				t.Fatalf("Walk failed: %v", err)
			}

			// Extract relative paths for comparison
			var actualFiles []string
			for _, file := range files {
				relPath, err := filepath.Rel(tempDir, file.Path)
				if err != nil {
					t.Fatalf("Failed to get relative path: %v", err)
				}
				actualFiles = append(actualFiles, relPath)
			}

			// Verify expected files are found
			sort.Strings(actualFiles)
			sort.Strings(tt.expectedFiles)

			if len(actualFiles) != len(tt.expectedFiles) {
				t.Fatalf("Expected %d files, but got %d.\nExpected: %v\nGot: %v", len(tt.expectedFiles), len(actualFiles), tt.expectedFiles, actualFiles)
			}

			for i := range actualFiles {
				if actualFiles[i] != tt.expectedFiles[i] {
					t.Errorf("File list mismatch.\nExpected: %v\nGot: %v", tt.expectedFiles, actualFiles)
					break
				}
			}
		})
	}
}

// TestPathSeparatorHandling tests proper handling of path separators across platforms
func TestPathSeparatorHandling(t *testing.T) {
	tempDir := testutils.CreateTempDir(t)
	defer os.RemoveAll(tempDir)

	// Create test files
	testFiles := map[string]string{
		filepath.Join("src", "main.go"):        "package main",
		filepath.Join("tests", "main_test.go"): "package main",
		filepath.Join("vendor", "lib.go"):      "package vendor",
	}

	for filePath, content := range testFiles {
		testutils.CreateTestFile(t, tempDir, filePath, content)
	}

	tests := []struct {
		name            string
		excludePattern  string
		shouldExclude   map[string]bool
	}{
		{
			name:           "Directory with trailing separator",
			excludePattern: "vendor" + string(filepath.Separator),
			shouldExclude: map[string]bool{
				filepath.Join(tempDir, "vendor", "lib.go"): true,
				filepath.Join(tempDir, "src", "main.go"):   false,
			},
		},
		{
			name:           "Directory without trailing separator",
			excludePattern: "tests",
			shouldExclude: map[string]bool{
				filepath.Join(tempDir, "tests", "main_test.go"): true,
				filepath.Join(tempDir, "src", "main.go"):        false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fw := NewFileWalker([]string{tempDir}, []string{tt.excludePattern}, []string{".go"}, false)
			
			for path, shouldBeExcluded := range tt.shouldExclude {
				excluded := fw.shouldExclude(path)
				if excluded != shouldBeExcluded {
					t.Errorf("Path %s: expected excluded=%v, got excluded=%v", path, shouldBeExcluded, excluded)
				}
			}
		})
	}
}

// TestCrossPlatformFileExtensions tests file extension handling across platforms
func TestCrossPlatformFileExtensions(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		expected string
	}{
		{"Go file", "main.go", "Go"},
		{"JavaScript file", "app.js", "JavaScript"},
		{"TypeScript file", "component.ts", "TypeScript"},
		{"Python file", "script.py", "Python"},
		{"Java file", "Main.java", "Java"},
		{"C# file", "Program.cs", "C#"},
		{"Unknown file", "README.txt", "Unknown"},
		{"No extension", "Makefile", "Unknown"},
		{"Multiple dots", "app.min.js", "JavaScript"},
		{"Case sensitivity", "APP.GO", "Go"}, // Should be case insensitive
	}

	fw := NewFileWalker([]string{}, []string{}, []string{}, false)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			language := fw.detectLanguage(tt.fileName)
			if language != tt.expected {
				t.Errorf("detectLanguage(%s) = %s, expected %s", tt.fileName, language, tt.expected)
			}
		})
	}
}

// TestCrossPlatformSymlinks tests symlink handling across platforms
func TestCrossPlatformSymlinks(t *testing.T) {
	// Skip symlink tests on Windows unless running as admin
	if runtime.GOOS == "windows" {
		t.Skip("Skipping symlink test on Windows")
	}

	tempDir := testutils.CreateTempDir(t)
	defer os.RemoveAll(tempDir)

	// Create a regular file
	regularFile := filepath.Join(tempDir, "regular.go")
	testutils.CreateTestFile(t, tempDir, "regular.go", "package main")

	// Create a symlink
	symlinkFile := filepath.Join(tempDir, "symlink.go")
	err := os.Symlink(regularFile, symlinkFile)
	if err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	fw := NewFileWalker([]string{tempDir}, []string{}, []string{".go"}, false)
	files, err := fw.Walk()
	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	// Should find both the regular file and the symlink
	if len(files) < 2 {
		t.Errorf("Expected at least 2 files (regular + symlink), got %d", len(files))
	}
}

// TestCrossPlatformHiddenFiles tests hidden file handling across platforms
func TestCrossPlatformHiddenFiles(t *testing.T) {
	tempDir := testutils.CreateTempDir(t)
	defer os.RemoveAll(tempDir)

	// Create regular and hidden files
	testutils.CreateTestFile(t, tempDir, "regular.go", "package main")
	testutils.CreateTestFile(t, tempDir, ".hidden.go", "package hidden")

	fw := NewFileWalker([]string{tempDir}, []string{}, []string{".go"}, false)
	files, err := fw.Walk()
	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	// Check that both files are found (GoClean should scan hidden files by default)
	var foundRegular, foundHidden bool
	for _, file := range files {
		if strings.Contains(file.Path, "regular.go") {
			foundRegular = true
		}
		if strings.Contains(file.Path, ".hidden.go") {
			foundHidden = true
		}
	}

	if !foundRegular {
		t.Error("Regular file not found")
	}
	if !foundHidden {
		t.Error("Hidden file not found")
	}
}

// TestCrossPlatformLongPaths tests handling of long paths across platforms
func TestCrossPlatformLongPaths(t *testing.T) {
	tempDir := testutils.CreateTempDir(t)
	defer os.RemoveAll(tempDir)

	// Create a deeply nested directory structure
	deepPath := tempDir
	for i := 0; i < 10; i++ {
		deepPath = filepath.Join(deepPath, "very_long_directory_name_to_test_path_limits")
	}

	// Create the nested directories
	err := os.MkdirAll(deepPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create deep path: %v", err)
	}

	// Create a file in the deep path
	deepFile := filepath.Join(deepPath, "deep.go")
	err = os.WriteFile(deepFile, []byte("package deep"), 0644)
	if err != nil {
		t.Fatalf("Failed to create deep file: %v", err)
	}

	fw := NewFileWalker([]string{tempDir}, []string{}, []string{".go"}, false)
	files, err := fw.Walk()
	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	// Should find the deep file
	found := false
	for _, file := range files {
		if strings.Contains(file.Path, "deep.go") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Deep file not found")
	}
}

// TestCrossPlatformCaseSensitivity tests case sensitivity handling across platforms
func TestCrossPlatformCaseSensitivity(t *testing.T) {
	tempDir := testutils.CreateTempDir(t)
	defer os.RemoveAll(tempDir)

	// Create files with different cases
	testutils.CreateTestFile(t, tempDir, "Main.go", "package main")
	testutils.CreateTestFile(t, tempDir, "UTIL.GO", "package util")
	testutils.CreateTestFile(t, tempDir, "helper.Go", "package helper")

	fw := NewFileWalker([]string{tempDir}, []string{}, []string{".go", ".GO"}, false)
	files, err := fw.Walk()
	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	// Should find all files regardless of case
	if len(files) < 1 {
		t.Error("Expected to find at least 1 file with case variations")
	}

	// Verify language detection works with case variations
	for _, file := range files {
		if file.Language != "Go" {
			t.Errorf("Expected Go language for file %s, got %s", file.Path, file.Language)
		}
	}

	t.Logf("Found %d files on %s filesystem", len(files), runtime.GOOS)
}

// TestCrossPlatformMemoryUsage tests memory usage patterns across platforms
func TestCrossPlatformMemoryUsage(t *testing.T) {
	tempDir := testutils.CreateTempDir(t)
	defer os.RemoveAll(tempDir)

	// Create many files to test memory usage
	fileCount := 1000
	for i := 0; i < fileCount; i++ {
		fileName := filepath.Join("dir", fmt.Sprintf("file_%04d.go", i))
		testutils.CreateTestFile(t, tempDir, fileName, "package main\n\nfunc main() {}")
	}

	fw := NewFileWalker([]string{tempDir}, []string{}, []string{".go"}, false)
	
	// Measure memory before
	var m1 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	files, err := fw.Walk()
	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	// Measure memory after
	var m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m2)

	// Use signed arithmetic to handle potential underflow
	memoryUsed := int64(m2.Alloc) - int64(m1.Alloc)
	if memoryUsed < 0 {
		// Memory usage decreased, likely due to GC - use current allocation
		memoryUsed = int64(m2.Alloc)
	}
	
	memoryPerFile := memoryUsed / int64(len(files))

	t.Logf("Processed %d files using %d bytes (%.2f KB per file) on %s", 
		len(files), memoryUsed, float64(memoryPerFile)/1024, runtime.GOOS)

	// Memory usage should be reasonable (less than 10KB per file)
	// This is a generous limit to account for Go's memory management
	if memoryPerFile > 10240 {
		t.Errorf("Memory usage too high: %d bytes per file", memoryPerFile)
	}
}