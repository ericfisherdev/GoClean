// Package scanner provides file scanning and pattern recognition capabilities for GoClean.
// It includes test file detection patterns for multiple programming languages and frameworks.
package scanner

import (
	"path/filepath"
	"strings"
)

// TestFilePatterns defines patterns for different testing frameworks
type TestFilePatterns struct {
	// Go test files
	GoTestSuffixes []string
	
	// JavaScript/TypeScript test files
	JSTestPatterns []string
	
	// PHP test files
	PHPTestPatterns []string
	
	// Python test files
	PythonTestPatterns []string
	
	// Java test files
	JavaTestPatterns []string
	
	// C#/.NET test files
	DotNetTestPatterns []string
	
	// Additional test directories
	TestDirectories []string
}

// Pattern matching constants
const (
	TestDirSeparator = "/"
)

// DefaultTestPatterns returns the default test file patterns for common frameworks
func DefaultTestPatterns() *TestFilePatterns {
	return &TestFilePatterns{
		GoTestSuffixes: []string{
			"_test.go",
			"_bench_test.go",
			"_example_test.go",
		},
		JSTestPatterns: []string{
			".test.js", ".test.ts", ".test.jsx", ".test.tsx",
			".spec.js", ".spec.ts", ".spec.jsx", ".spec.tsx",
			"__tests__/*.js", "__tests__/*.ts",
			"*.test.mjs", "*.spec.mjs",
		},
		PHPTestPatterns: []string{
			"*Test.php",
			"*test.php", 
			"Test*.php",
			"tests/*.php",
			"Tests/*.php",
		},
		PythonTestPatterns: []string{
			"test_*.py",
			"*_test.py",
			"test*.py",
			"tests/*.py",
			"testing/*.py",
		},
		JavaTestPatterns: []string{
			"*Test.java",
			"*Tests.java", 
			"*TestCase.java",
			"src/test/**/*.java",
		},
		DotNetTestPatterns: []string{
			"*Test.cs",
			"*Tests.cs",
			"*.Test.cs",
			"*.Tests.cs",
			"*Spec.cs",
			"*.Spec.cs",
		},
		TestDirectories: []string{
			"test/", "tests/", "testing/", "spec/", "specs/",
			"__test__/", "__tests__/", "__testing__/",
			"t/", // Perl convention
			"tst/", "unittest/", "integrationtest/",
		},
	}
}

// IsTestFile determines if a file is a test file based on its path and name
func (tp *TestFilePatterns) IsTestFile(filePath string) bool {
	fileName := filepath.Base(filePath)
	dir := filepath.Dir(filePath)
	
	// Check Go test patterns
	for _, suffix := range tp.GoTestSuffixes {
		if strings.HasSuffix(fileName, suffix) {
			return true
		}
	}
	
	// Check JavaScript/TypeScript patterns
	for _, pattern := range tp.JSTestPatterns {
		if matched, _ := filepath.Match(pattern, fileName); matched {
			return true
		}
		if strings.Contains(filePath, strings.TrimSuffix(pattern, "/*")) {
			return true
		}
	}
	
	// Check PHP patterns
	for _, pattern := range tp.PHPTestPatterns {
		if matched, _ := filepath.Match(pattern, fileName); matched {
			return true
		}
	}
	
	// Check Python patterns
	for _, pattern := range tp.PythonTestPatterns {
		if matched, _ := filepath.Match(pattern, fileName); matched {
			return true
		}
	}
	
	// Check Java patterns
	for _, pattern := range tp.JavaTestPatterns {
		if matched, _ := filepath.Match(pattern, fileName); matched {
			return true
		}
		if strings.Contains(filePath, "src/test/") {
			return true
		}
	}
	
	// Check .NET patterns
	for _, pattern := range tp.DotNetTestPatterns {
		if matched, _ := filepath.Match(pattern, fileName); matched {
			return true
		}
	}
	
	// Check test directories
	for _, testDir := range tp.TestDirectories {
		if strings.Contains(dir+TestDirSeparator, TestDirSeparator+testDir) {
			return true
		}
	}
	
	return false
}

// AddCustomPatterns allows adding custom test patterns to the existing set
func (tp *TestFilePatterns) AddCustomPatterns(customPatterns []string) {
	// Add custom patterns to Go test suffixes for now
	// This could be enhanced to categorize by language type
	tp.GoTestSuffixes = append(tp.GoTestSuffixes, customPatterns...)
}

// GetAllPatterns returns all patterns as a flat slice for debugging
func (tp *TestFilePatterns) GetAllPatterns() []string {
	var allPatterns []string
	
	allPatterns = append(allPatterns, tp.GoTestSuffixes...)
	allPatterns = append(allPatterns, tp.JSTestPatterns...)
	allPatterns = append(allPatterns, tp.PHPTestPatterns...)
	allPatterns = append(allPatterns, tp.PythonTestPatterns...)
	allPatterns = append(allPatterns, tp.JavaTestPatterns...)
	allPatterns = append(allPatterns, tp.DotNetTestPatterns...)
	allPatterns = append(allPatterns, tp.TestDirectories...)
	
	return allPatterns
}