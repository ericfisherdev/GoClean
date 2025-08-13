package scanner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ericfisherdev/goclean/internal/models"
)

func TestNewParser(t *testing.T) {
	parser := NewParser(true)
	if parser == nil {
		t.Fatal("Expected parser to be created, got nil")
	}
	if !parser.verbose {
		t.Error("Expected verbose to be true")
	}

	parser = NewParser(false)
	if parser.verbose {
		t.Error("Expected verbose to be false")
	}
}

func TestParseGoFile(t *testing.T) {
	tmpDir := t.TempDir()
	goFile := filepath.Join(tmpDir, "test.go")

	content := `package main

import "fmt"

// main is the entry point
func main() {
    fmt.Println("Hello, World!")
    helper()
}

// helper does something useful
func helper() int {
    // This is a comment
    return 42
}

type TestStruct struct {
    Value int
}
`

	err := os.WriteFile(goFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test Go file: %v", err)
	}

	fileInfo := &models.FileInfo{
		Path:      goFile,
		Name:      "test.go",
		Extension: ".go",
		Language:  "Go",
	}

	parser := NewParser(false)
	result, err := parser.ParseFile(fileInfo)
	if err != nil {
		t.Fatalf("Failed to parse Go file: %v", err)
	}

	if result == nil {
		t.Fatal("Expected parse result, got nil")
	}

	if !fileInfo.Scanned {
		t.Error("Expected file to be marked as scanned")
	}

	if result.Metrics.TotalLines == 0 {
		t.Error("Expected total lines to be counted")
	}

	if result.Metrics.FunctionCount != 2 {
		t.Errorf("Expected 2 functions, got %d", result.Metrics.FunctionCount)
	}

	if result.Metrics.ClassCount != 1 {
		t.Errorf("Expected 1 struct (class), got %d", result.Metrics.ClassCount)
	}

	if result.Metrics.CommentLines == 0 {
		t.Error("Expected comment lines to be counted")
	}

	if result.Metrics.CodeLines == 0 {
		t.Error("Expected code lines to be counted")
	}
}

func TestParseJavaScriptFile(t *testing.T) {
	tmpDir := t.TempDir()
	jsFile := filepath.Join(tmpDir, "test.js")

	content := `// JavaScript test file
function greet(name) {
    console.log("Hello, " + name);
}

const arrowFunc = (x, y) => {
    return x + y;
}

class Calculator {
    constructor() {
        this.result = 0;
    }
    
    add(value) {
        this.result += value;
        return this;
    }
}

// Another comment
function multiply(a, b) {
    return a * b;
}
`

	err := os.WriteFile(jsFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test JS file: %v", err)
	}

	fileInfo := &models.FileInfo{
		Path:      jsFile,
		Name:      "test.js",
		Extension: ".js",
		Language:  "JavaScript",
	}

	parser := NewParser(false)
	result, err := parser.ParseFile(fileInfo)
	if err != nil {
		t.Fatalf("Failed to parse JavaScript file: %v", err)
	}

	if result.Metrics.FunctionCount != 3 {
		t.Errorf("Expected 3 functions, got %d", result.Metrics.FunctionCount)
	}

	if result.Metrics.ClassCount != 1 {
		t.Errorf("Expected 1 class, got %d", result.Metrics.ClassCount)
	}

	if result.Metrics.CommentLines == 0 {
		t.Error("Expected comment lines to be counted")
	}
}

func TestParsePythonFile(t *testing.T) {
	tmpDir := t.TempDir()
	pyFile := filepath.Join(tmpDir, "test.py")

	content := `# Python test file
def greet(name):
    print(f"Hello, {name}")

class TestClass:
    def __init__(self):
        self.value = 42
    
    def get_value(self):
        # Return the stored value
        return self.value

def helper_function():
    pass
`

	err := os.WriteFile(pyFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test Python file: %v", err)
	}

	fileInfo := &models.FileInfo{
		Path:      pyFile,
		Name:      "test.py",
		Extension: ".py",
		Language:  "Python",
	}

	parser := NewParser(false)
	result, err := parser.ParseFile(fileInfo)
	if err != nil {
		t.Fatalf("Failed to parse Python file: %v", err)
	}

	if result.Metrics.FunctionCount != 4 { // greet, __init__, get_value, helper_function
		t.Errorf("Expected 4 functions, got %d", result.Metrics.FunctionCount)
	}

	if result.Metrics.ClassCount != 1 {
		t.Errorf("Expected 1 class, got %d", result.Metrics.ClassCount)
	}
}

func TestParseNonExistentFile(t *testing.T) {
	fileInfo := &models.FileInfo{
		Path:      "/non/existent/file.go",
		Name:      "file.go",
		Extension: ".go",
		Language:  "Go",
	}

	parser := NewParser(false)
	result, err := parser.ParseFile(fileInfo)

	if err == nil {
		t.Error("Expected error for non-existent file")
	}

	if result != nil {
		t.Error("Expected nil result for non-existent file")
	}

	if fileInfo.Scanned {
		t.Error("Expected file to not be marked as scanned")
	}
}

func TestIsCommentLine(t *testing.T) {
	parser := NewParser(false)

	testCases := []struct {
		line     string
		language string
		expected bool
	}{
		// Go/C-style comments
		{"// This is a comment", "Go", true},
		{"/* This is a comment */", "Go", true},
		{"* This is a comment", "Go", true},
		{"func main() {}", "Go", false},
		
		// Python comments
		{"# This is a comment", "Python", true},
		{"def test():", "Python", false},
		
		// JavaScript comments
		{"// JavaScript comment", "JavaScript", true},
		{"/* Block comment */", "JavaScript", true},
		{"function test() {}", "JavaScript", false},
		
		// PHP comments
		{"// PHP comment", "PHP", true},
		{"# PHP comment", "PHP", true},
		{"/* PHP comment */", "PHP", true},
		{"<?php", "PHP", false},
	}

	for _, tc := range testCases {
		result := parser.isCommentLine(tc.line, tc.language)
		if result != tc.expected {
			t.Errorf("isCommentLine(%q, %q) = %v, expected %v", tc.line, tc.language, result, tc.expected)
		}
	}
}

func TestIsFunctionDeclaration(t *testing.T) {
	parser := NewParser(false)

	testCases := []struct {
		line     string
		language string
		expected bool
	}{
		// Go functions
		{"func main() {", "Go", true},
		{"func helper(x int) string {", "Go", true},
		{"var x = 5", "Go", false},
		
		// JavaScript functions
		{"function test() {", "JavaScript", true},
		{"const func = () => {", "JavaScript", true},
		{"var x = 5;", "JavaScript", false},
		
		// Python functions
		{"def greet(name):", "Python", true},
		{"    def helper():", "Python", true},
		{"class Test:", "Python", false},
		
		// Java functions
		{"public void test() {", "Java", true},
		{"private static int helper(int x) {", "Java", true},
		{"public class Test {", "Java", false},
		
		// C++ functions
		{"int main() {", "C++", true},
		{"void helper(int x) {", "C++", true},
		{"class Test {", "C++", false},
	}

	for _, tc := range testCases {
		result := parser.isFunctionDeclaration(tc.line, tc.language)
		if result != tc.expected {
			t.Errorf("isFunctionDeclaration(%q, %q) = %v, expected %v", tc.line, tc.language, result, tc.expected)
		}
	}
}

func TestIsClassDeclaration(t *testing.T) {
	parser := NewParser(false)

	testCases := []struct {
		line     string
		language string
		expected bool
	}{
		// Go structs
		{"type TestStruct struct {", "Go", true},
		{"func main() {", "Go", false},
		
		// JavaScript classes
		{"class Calculator {", "JavaScript", true},
		{"function test() {", "JavaScript", false},
		
		// Python classes
		{"class TestClass:", "Python", true},
		{"def test():", "Python", false},
		
		// Java classes
		{"public class Test {", "Java", true},
		{"private void method() {", "Java", false},
		
		// C++ classes
		{"class Calculator {", "C++", true},
		{"int main() {", "C++", false},
	}

	for _, tc := range testCases {
		result := parser.isClassDeclaration(tc.line, tc.language)
		if result != tc.expected {
			t.Errorf("isClassDeclaration(%q, %q) = %v, expected %v", tc.line, tc.language, result, tc.expected)
		}
	}
}

func TestAnalyzeLine(t *testing.T) {
	parser := NewParser(false)
	metrics := &models.FileMetrics{}
	fileInfo := &models.FileInfo{Language: "Go"}

	// Test blank line
	parser.analyzeLine("", 1, metrics, fileInfo)
	if metrics.BlankLines != 1 {
		t.Errorf("Expected 1 blank line, got %d", metrics.BlankLines)
	}

	// Test comment line
	parser.analyzeLine("// This is a comment", 2, metrics, fileInfo)
	if metrics.CommentLines != 1 {
		t.Errorf("Expected 1 comment line, got %d", metrics.CommentLines)
	}

	// Test code line
	parser.analyzeLine("func main() {", 3, metrics, fileInfo)
	if metrics.CodeLines != 1 {
		t.Errorf("Expected 1 code line, got %d", metrics.CodeLines)
	}
	if metrics.FunctionCount != 1 {
		t.Errorf("Expected 1 function, got %d", metrics.FunctionCount)
	}

	// Test struct line
	parser.analyzeLine("type Test struct {", 4, metrics, fileInfo)
	if metrics.CodeLines != 2 {
		t.Errorf("Expected 2 code lines, got %d", metrics.CodeLines)
	}
	if metrics.ClassCount != 1 {
		t.Errorf("Expected 1 struct, got %d", metrics.ClassCount)
	}
}

func TestLooksLikeFunctionSignature(t *testing.T) {
	parser := NewParser(false)

	testCases := []struct {
		line     string
		expected bool
	}{
		{"int main(int argc, char** argv) {", true},
		{"void helper() {", true},
		{"static int calculate(int x, int y) {", true},
		{"#define MACRO(x) (x*2)", false},
		{"if (condition) {", true}, // Basic parser sees "if" as identifier + parentheses
		{"printf(\"hello\");", true}, // Basic parser sees "printf" as identifier + parentheses
		{"", false},
		{"int", false},
	}

	for _, tc := range testCases {
		result := parser.looksLikeFunctionSignature(tc.line)
		if result != tc.expected {
			t.Errorf("looksLikeFunctionSignature(%q) = %v, expected %v", tc.line, result, tc.expected)
		}
	}
}

func TestParseFileMetrics(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "metrics.go")

	content := `package main

import "fmt"

// This is a comment
func main() {
    fmt.Println("Hello")
    
    helper()
}

/* Multi-line
   comment */
func helper() {
    // Another comment
}

type TestStruct struct {
    Value int
}
`

	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	fileInfo := &models.FileInfo{
		Path:      testFile,
		Name:      "metrics.go",
		Extension: ".go",
		Language:  "Go",
	}

	parser := NewParser(false)
	result, err := parser.ParseFile(fileInfo)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	// Verify comprehensive metrics
	if result.Metrics.TotalLines == 0 {
		t.Error("Expected total lines to be counted")
	}

	if result.Metrics.BlankLines == 0 {
		t.Error("Expected blank lines to be counted")
	}

	if result.Metrics.CommentLines == 0 {
		t.Error("Expected comment lines to be counted")
	}

	if result.Metrics.CodeLines == 0 {
		t.Error("Expected code lines to be counted")
	}

	// Verify the sum
	total := result.Metrics.BlankLines + result.Metrics.CommentLines + result.Metrics.CodeLines
	if total != result.Metrics.TotalLines {
		t.Errorf("Line counts don't add up: blank(%d) + comment(%d) + code(%d) = %d, expected %d",
			result.Metrics.BlankLines, result.Metrics.CommentLines, result.Metrics.CodeLines,
			total, result.Metrics.TotalLines)
	}

	// Verify function and struct counting
	if result.Metrics.FunctionCount != 2 {
		t.Errorf("Expected 2 functions, got %d", result.Metrics.FunctionCount)
	}

	if result.Metrics.ClassCount != 1 {
		t.Errorf("Expected 1 struct, got %d", result.Metrics.ClassCount)
	}
}