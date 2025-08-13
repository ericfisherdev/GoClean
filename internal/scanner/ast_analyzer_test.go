package scanner

import (
	"path/filepath"
	"testing"

	"github.com/ericfisherdev/goclean/internal/testutils"
	"github.com/ericfisherdev/goclean/internal/types"
)

func TestNewASTAnalyzer(t *testing.T) {
	analyzer := NewASTAnalyzer(true)
	
	if analyzer == nil {
		t.Fatal("Expected analyzer to be created, got nil")
	}
	
	if analyzer.fileSet == nil {
		t.Error("Expected fileSet to be initialized")
	}
	
	if !analyzer.verbose {
		t.Error("Expected verbose to be true")
	}
}

func TestAnalyzeGoFile_BasicFunction(t *testing.T) {
	analyzer := NewASTAnalyzer(false)
	
	// Create a simple Go file for testing
	tempDir := testutils.CreateTempDir(t)
	goCode := `package test

import "fmt"

// TestFunction is a test function
func TestFunction(a int, b string) (int, error) {
	if a > 0 {
		return a + 1, nil
	}
	return 0, fmt.Errorf("invalid input")
}`

	filePath := testutils.CreateTestFile(t, tempDir, "test.go", goCode)
	
	// Analyze the file
	astInfo, err := analyzer.AnalyzeGoFile(filePath)
	if err != nil {
		t.Fatalf("Failed to analyze Go file: %v", err)
	}
	
	// Verify basic information
	if astInfo.PackageName != "test" {
		t.Errorf("Expected package name 'test', got %s", astInfo.PackageName)
	}
	
	if len(astInfo.Functions) != 1 {
		t.Errorf("Expected 1 function, got %d", len(astInfo.Functions))
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
	
	if fn.Parameters[0].Name != "a" || fn.Parameters[0].Type != "int" {
		t.Errorf("Expected parameter 'a int', got '%s %s'", fn.Parameters[0].Name, fn.Parameters[0].Type)
	}
	
	if len(fn.Results) != 2 {
		t.Errorf("Expected 2 return types, got %d", len(fn.Results))
	}
	
	if !fn.IsExported {
		t.Error("Expected function to be exported")
	}
	
	if fn.HasComments != true {
		t.Error("Expected function to have comments")
	}
	
	// Verify complexity (should be at least 2 due to if statement)
	if fn.Complexity < 2 {
		t.Errorf("Expected complexity >= 2, got %d", fn.Complexity)
	}
}

func TestAnalyzeGoFile_Struct(t *testing.T) {
	analyzer := NewASTAnalyzer(false)
	
	tempDir := testutils.CreateTempDir(t)
	goCode := `package test

// Person represents a person
type Person struct {
	Name string
	Age  int
	Email string
}

// Address is unexported
type address struct {
	Street string
	City   string
}`

	filePath := testutils.CreateTestFile(t, tempDir, "types.go", goCode)
	
	astInfo, err := analyzer.AnalyzeGoFile(filePath)
	if err != nil {
		t.Fatalf("Failed to analyze Go file: %v", err)
	}
	
	if len(astInfo.Types) != 2 {
		t.Errorf("Expected 2 types, got %d", len(astInfo.Types))
	}
	
	// Find Person struct
	var personType *types.TypeInfo
	for _, typeInfo := range astInfo.Types {
		if typeInfo.Name == "Person" {
			personType = typeInfo
			break
		}
	}
	
	if personType == nil {
		t.Fatal("Expected to find Person type")
	}
	
	if personType.Kind != "struct" {
		t.Errorf("Expected Person to be struct, got %s", personType.Kind)
	}
	
	if !personType.IsExported {
		t.Error("Expected Person to be exported")
	}
	
	if personType.FieldCount != 3 {
		t.Errorf("Expected Person to have 3 fields, got %d", personType.FieldCount)
	}
	
	// Check unexported type
	var addressType *types.TypeInfo
	for _, typeInfo := range astInfo.Types {
		if typeInfo.Name == "address" {
			addressType = typeInfo
			break
		}
	}
	
	if addressType == nil {
		t.Fatal("Expected to find address type")
	}
	
	if addressType.IsExported {
		t.Error("Expected address to be unexported")
	}
}

func TestAnalyzeGoFile_Interface(t *testing.T) {
	analyzer := NewASTAnalyzer(false)
	
	tempDir := testutils.CreateTempDir(t)
	goCode := `package test

// Writer interface for writing data
type Writer interface {
	Write(data []byte) error
	Close() error
}`

	filePath := testutils.CreateTestFile(t, tempDir, "interface.go", goCode)
	
	astInfo, err := analyzer.AnalyzeGoFile(filePath)
	if err != nil {
		t.Fatalf("Failed to analyze Go file: %v", err)
	}
	
	if len(astInfo.Types) != 1 {
		t.Errorf("Expected 1 type, got %d", len(astInfo.Types))
	}
	
	writer := astInfo.Types[0]
	if writer.Name != "Writer" {
		t.Errorf("Expected type name 'Writer', got %s", writer.Name)
	}
	
	if writer.Kind != "interface" {
		t.Errorf("Expected Writer to be interface, got %s", writer.Kind)
	}
	
	if writer.MethodCount != 2 {
		t.Errorf("Expected Writer to have 2 methods, got %d", writer.MethodCount)
	}
}

func TestAnalyzeGoFile_Methods(t *testing.T) {
	analyzer := NewASTAnalyzer(false)
	
	tempDir := testutils.CreateTempDir(t)
	goCode := `package test

type Counter struct {
	value int
}

// Increment increases the counter
func (c *Counter) Increment() {
	c.value++
}

// Value returns the current value
func (c Counter) Value() int {
	return c.value
}`

	filePath := testutils.CreateTestFile(t, tempDir, "methods.go", goCode)
	
	astInfo, err := analyzer.AnalyzeGoFile(filePath)
	if err != nil {
		t.Fatalf("Failed to analyze Go file: %v", err)
	}
	
	if len(astInfo.Functions) != 2 {
		t.Errorf("Expected 2 functions, got %d", len(astInfo.Functions))
	}
	
	// Check both methods
	for _, fn := range astInfo.Functions {
		if !fn.IsMethod {
			t.Errorf("Expected %s to be a method", fn.Name)
		}
		
		if fn.ReceiverType == "" {
			t.Errorf("Expected %s to have receiver type", fn.Name)
		}
	}
}

func TestAnalyzeGoFile_VariablesAndConstants(t *testing.T) {
	analyzer := NewASTAnalyzer(false)
	
	tempDir := testutils.CreateTempDir(t)
	goCode := `package test

// Public constants
const (
	MaxRetries = 3
	TimeoutMS  = 5000
)

// private constant
const defaultValue = "test"

// Public variables
var (
	GlobalCounter int
	IsEnabled     bool
)

// private variable
var internal string`

	filePath := testutils.CreateTestFile(t, tempDir, "vars.go", goCode)
	
	astInfo, err := analyzer.AnalyzeGoFile(filePath)
	if err != nil {
		t.Fatalf("Failed to analyze Go file: %v", err)
	}
	
	if len(astInfo.Constants) != 3 {
		t.Errorf("Expected 3 constants, got %d", len(astInfo.Constants))
	}
	
	if len(astInfo.Variables) != 3 {
		t.Errorf("Expected 3 variables, got %d", len(astInfo.Variables))
	}
	
	// Check exported vs unexported
	exportedConsts := 0
	for _, c := range astInfo.Constants {
		if c.IsExported {
			exportedConsts++
		}
	}
	
	if exportedConsts != 2 {
		t.Errorf("Expected 2 exported constants, got %d", exportedConsts)
	}
}

func TestAnalyzeGoFile_ComplexityCalculation(t *testing.T) {
	analyzer := NewASTAnalyzer(false)
	
	tempDir := testutils.CreateTempDir(t)
	goCode := `package test

func ComplexFunction(x int) int {
	if x > 0 {          // +1
		if x < 100 {    // +1
			for i := 0; i < x; i++ {  // +1
				if i%2 == 0 {         // +1
					return i
				}
			}
		}
	} else if x < 0 {   // +1
		return -1
	}
	
	switch x {          // +1
	case 0:             // +1
		return 1
	case 1:             // +1
		return 2
	default:
		return x
	}
}`

	filePath := testutils.CreateTestFile(t, tempDir, "complex.go", goCode)
	
	astInfo, err := analyzer.AnalyzeGoFile(filePath)
	if err != nil {
		t.Fatalf("Failed to analyze Go file: %v", err)
	}
	
	if len(astInfo.Functions) != 1 {
		t.Errorf("Expected 1 function, got %d", len(astInfo.Functions))
	}
	
	fn := astInfo.Functions[0]
	// Base complexity is 1, plus 8 control flow statements = 9
	expectedComplexity := 9
	if fn.Complexity != expectedComplexity {
		t.Errorf("Expected complexity %d, got %d", expectedComplexity, fn.Complexity)
	}
}

func TestAnalyzeGoFile_ImportAnalysis(t *testing.T) {
	analyzer := NewASTAnalyzer(false)
	
	tempDir := testutils.CreateTempDir(t)
	goCode := `package test

import (
	"fmt"
	"os"
	. "strings"
	custom "github.com/user/package"
)`

	filePath := testutils.CreateTestFile(t, tempDir, "imports.go", goCode)
	
	astInfo, err := analyzer.AnalyzeGoFile(filePath)
	if err != nil {
		t.Fatalf("Failed to analyze Go file: %v", err)
	}
	
	if len(astInfo.Imports) != 4 {
		t.Errorf("Expected 4 imports, got %d", len(astInfo.Imports))
	}
	
	// Check for dot import
	var dotImport *types.ImportInfo
	for _, imp := range astInfo.Imports {
		if imp.Alias == "." {
			dotImport = imp
			break
		}
	}
	
	if dotImport == nil {
		t.Error("Expected to find dot import")
	} else if dotImport.Path != "strings" {
		t.Errorf("Expected dot import path 'strings', got %s", dotImport.Path)
	}
	
	// Check for alias import
	var aliasImport *types.ImportInfo
	for _, imp := range astInfo.Imports {
		if imp.Alias == "custom" {
			aliasImport = imp
			break
		}
	}
	
	if aliasImport == nil {
		t.Error("Expected to find alias import")
	} else if aliasImport.Path != "github.com/user/package" {
		t.Errorf("Expected alias import path 'github.com/user/package', got %s", aliasImport.Path)
	}
}

func TestAnalyzeGoFile_InvalidFile(t *testing.T) {
	analyzer := NewASTAnalyzer(false)
	
	tempDir := testutils.CreateTempDir(t)
	invalidCode := `package test

func InvalidFunction( {
	// Missing closing parenthesis and proper syntax
}`

	filePath := testutils.CreateTestFile(t, tempDir, "invalid.go", invalidCode)
	
	_, err := analyzer.AnalyzeGoFile(filePath)
	if err == nil {
		t.Error("Expected error when analyzing invalid Go file")
	}
}

func TestAnalyzeGoFile_NonExistentFile(t *testing.T) {
	analyzer := NewASTAnalyzer(false)
	
	tempDir := testutils.CreateTempDir(t)
	nonExistentPath := filepath.Join(tempDir, "does-not-exist.go")
	
	_, err := analyzer.AnalyzeGoFile(nonExistentPath)
	if err == nil {
		t.Error("Expected error when analyzing non-existent file")
	}
}

func TestExtractTypeName(t *testing.T) {
	// This would require creating AST nodes manually, which is complex
	// For now, we'll test it indirectly through the main analysis functions
	// In a full implementation, we'd create mock AST nodes to test edge cases
}

func TestCalculateCyclomaticComplexity_BaseCase(t *testing.T) {
	analyzer := NewASTAnalyzer(false)
	
	tempDir := testutils.CreateTempDir(t)
	goCode := `package test

func SimpleFunction() {
	return
}`

	filePath := testutils.CreateTestFile(t, tempDir, "simple.go", goCode)
	
	astInfo, err := analyzer.AnalyzeGoFile(filePath)
	if err != nil {
		t.Fatalf("Failed to analyze Go file: %v", err)
	}
	
	if len(astInfo.Functions) != 1 {
		t.Errorf("Expected 1 function, got %d", len(astInfo.Functions))
	}
	
	fn := astInfo.Functions[0]
	if fn.Complexity != 1 {
		t.Errorf("Expected base complexity 1, got %d", fn.Complexity)
	}
}