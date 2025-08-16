package violations
import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/types"
)
func TestDuplicationDetector_Detect_NoDuplication(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewDuplicationDetector(config)
	
	source := `
package main

func uniqueFunction1() {
    fmt.Println("First function")
    doSomething()
    return true
}

func uniqueFunction2() {
    fmt.Println("Second function")
    doSomethingElse()
    return false
}
`
	
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, "test.go", source, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}
	
	fileInfo := &models.FileInfo{
		Path:     "/test/example.go",
		Language: "Go",
	}
	astInfo := &types.GoASTInfo{
		AST:     astFile,
		FileSet: fset,
	}
	
	violations := detector.Detect(fileInfo, astInfo)
	if len(violations) != 0 {
		t.Errorf("Expected no violations, got %d", len(violations))
	}
}
func TestDuplicationDetector_Detect_WithDuplication(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewDuplicationDetector(config)
	
	// First file with duplicate code
	source1 := `
package main

func processData() {
    data := getData()
    if data == nil {
        return nil
    }
    result := transform(data)
    return result
}
`
	
	fset1 := token.NewFileSet()
	astFile1, err := parser.ParseFile(fset1, "file1.go", source1, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse first file: %v", err)
	}
	
	fileInfo1 := &models.FileInfo{
		Path:     "/test/file1.go",
		Language: "Go",
	}
	astInfo1 := &types.GoASTInfo{
		AST:     astFile1,
		FileSet: fset1,
	}
	
	// Process first file - should have no violations (first occurrence)
	violations1 := detector.Detect(fileInfo1, astInfo1)
	if len(violations1) != 0 {
		t.Errorf("Expected no violations for first occurrence, got %d", len(violations1))
	}
	
	// Second file with same code (different function name)
	source2 := `
package main

func handleData() {
    data := getData()
    if data == nil {
        return nil
    }
    result := transform(data)
    return result
}
`
	
	fset2 := token.NewFileSet()
	astFile2, err := parser.ParseFile(fset2, "file2.go", source2, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse second file: %v", err)
	}
	
	fileInfo2 := &models.FileInfo{
		Path:     "/test/file2.go",
		Language: "Go",
	}
	astInfo2 := &types.GoASTInfo{
		AST:     astFile2,
		FileSet: fset2,
	}
	
	// Process second file - should detect duplication
	violations2 := detector.Detect(fileInfo2, astInfo2)
	if len(violations2) != 1 {
		t.Errorf("Expected 1 duplication violation, got %d", len(violations2))
		return // Prevent index out of bounds
	}
	
	violation := violations2[0]
	if violation.Type != models.ViolationTypeDuplication {
		t.Errorf("Expected duplication violation type, got %s", violation.Type)
	}
	if violation.File != "/test/file2.go" {
		t.Errorf("Expected violation in file2.go, got %s", violation.File)
	}
}
func TestDuplicationDetector_Reset(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewDuplicationDetector(config)
	
	source := `
package main

func testFunction() {
    fmt.Println("Test function")
    data := getData()
    processData(data)
    return true
}
`
	
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, "test.go", source, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}
	
	fileInfo := &models.FileInfo{
		Path:     "/test/example.go",
		Language: "Go",
	}
	astInfo := &types.GoASTInfo{
		AST:     astFile,
		FileSet: fset,
	}
	
	// Process file to populate cache
	detector.Detect(fileInfo, astInfo)
	// Check cache is not empty
	if len(detector.hashCache) == 0 {
		t.Error("Expected cache to be populated")
	}
	// Reset cache
	detector.Reset()
	// Check cache is empty
	if len(detector.hashCache) != 0 {
		t.Error("Expected cache to be empty after reset")
	}
}
func TestDuplicationDetector_IgnoreSmallFunctions(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewDuplicationDetector(config)
	
	// Small function (less than 5 lines)
	source := `
package main

func smallFunction() {
    return true
}
`
	
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, "test.go", source, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}
	
	fileInfo := &models.FileInfo{
		Path:     "/test/example.go",
		Language: "Go",
	}
	astInfo := &types.GoASTInfo{
		AST:     astFile,
		FileSet: fset,
	}
	
	violations := detector.Detect(fileInfo, astInfo)
	// Should not detect violations for small functions
	if len(violations) != 0 {
		t.Errorf("Expected no violations for small function, got %d", len(violations))
	}
}