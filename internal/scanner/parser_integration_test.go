package scanner

import (
	"testing"

	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/testutils"
)

func TestParseFile_GoFileUsesAST(t *testing.T) {
	parser := NewParser(true)
	
	// Create a Go file for testing
	tempDir := testutils.CreateTempDir(t)
	goCode := `package test

import "fmt"

// TestFunction demonstrates a function with parameters
func TestFunction(a int, b string) (int, error) {
	if a > 0 {
		return a + 1, nil
	}
	return 0, fmt.Errorf("invalid input")
}

type Person struct {
	Name string
	Age  int
}`

	filePath := testutils.CreateTestFile(t, tempDir, "test.go", goCode)
	
	// Create file info
	fileInfo := &models.FileInfo{
		Path:     filePath,
		Language: "Go",
	}
	
	// Parse the file
	result, err := parser.ParseFile(fileInfo)
	if err != nil {
		t.Fatalf("Failed to parse Go file: %v", err)
	}
	
	// Verify AST info is available
	if result.ASTInfo == nil {
		t.Error("Expected AST info to be available for Go file")
	}
	
	// Cast and verify AST info
	astInfo, ok := result.ASTInfo.(*GoASTInfo)
	if !ok {
		t.Error("Expected ASTInfo to be of type *GoASTInfo")
	} else {
		// Verify AST parsing extracted correct information
		if astInfo.PackageName != "test" {
			t.Errorf("Expected package name 'test', got %s", astInfo.PackageName)
		}
		
		if len(astInfo.Functions) != 1 {
			t.Errorf("Expected 1 function, got %d", len(astInfo.Functions))
		}
		
		if len(astInfo.Types) != 1 {
			t.Errorf("Expected 1 type, got %d", len(astInfo.Types))
		}
		
		if len(astInfo.Imports) != 1 {
			t.Errorf("Expected 1 import, got %d", len(astInfo.Imports))
		}
		
		// Verify function details
		fn := astInfo.Functions[0]
		if fn.Name != "TestFunction" {
			t.Errorf("Expected function name 'TestFunction', got %s", fn.Name)
		}
		
		if len(fn.Parameters) != 2 {
			t.Errorf("Expected 2 parameters, got %d", len(fn.Parameters))
		}
		
		if fn.Complexity < 2 {
			t.Errorf("Expected complexity >= 2, got %d", fn.Complexity)
		}
	}
	
	// Verify metrics are correctly extracted
	if result.Metrics.FunctionCount != 1 {
		t.Errorf("Expected function count 1, got %d", result.Metrics.FunctionCount)
	}
	
	if result.Metrics.ClassCount != 1 { // Person struct
		t.Errorf("Expected class count 1, got %d", result.Metrics.ClassCount)
	}
}

func TestParseFile_NonGoFileUsesLineByLine(t *testing.T) {
	parser := NewParser(true)
	
	// Create a non-Go file for testing
	tempDir := testutils.CreateTempDir(t)
	jsCode := `function testFunction() {
    console.log("Hello, world!");
    return 42;
}`

	filePath := testutils.CreateTestFile(t, tempDir, "test.js", jsCode)
	
	// Create file info
	fileInfo := &models.FileInfo{
		Path:     filePath,
		Language: "JavaScript",
	}
	
	// Parse the file
	result, err := parser.ParseFile(fileInfo)
	if err != nil {
		t.Fatalf("Failed to parse JavaScript file: %v", err)
	}
	
	// Verify AST info is NOT available (should use line-by-line parsing)
	if result.ASTInfo != nil {
		t.Error("Expected AST info to be nil for non-Go file")
	}
	
	// Verify basic metrics are still extracted
	if result.Metrics.TotalLines == 0 {
		t.Error("Expected total lines to be > 0")
	}
	
	if result.Metrics.FunctionCount != 1 {
		t.Errorf("Expected function count 1, got %d", result.Metrics.FunctionCount)
	}
}

func TestParseFile_InvalidGoFileFallsBackToLineByLine(t *testing.T) {
	parser := NewParser(false) // Disable verbose to avoid cluttering test output
	
	// Create an invalid Go file that will fail AST parsing
	tempDir := testutils.CreateTempDir(t)
	invalidGoCode := `package test

func InvalidFunction( {
    // Missing closing parenthesis
}`

	filePath := testutils.CreateTestFile(t, tempDir, "invalid.go", invalidGoCode)
	
	// Create file info
	fileInfo := &models.FileInfo{
		Path:     filePath,
		Language: "Go",
	}
	
	// Parse the file - should fall back to line-by-line parsing
	result, err := parser.ParseFile(fileInfo)
	if err != nil {
		t.Fatalf("Failed to parse invalid Go file (should have fallen back): %v", err)
	}
	
	// Verify AST info is not available (fell back to line parsing)
	if result.ASTInfo != nil {
		t.Error("Expected AST info to be nil when AST parsing fails")
	}
	
	// Verify basic metrics are still extracted via line parsing
	if result.Metrics.TotalLines == 0 {
		t.Error("Expected total lines to be > 0")
	}
}