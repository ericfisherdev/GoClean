// Package scanner provides file scanning and pattern recognition capabilities for GoClean.
// It includes test file detection patterns for multiple programming languages and frameworks.
package scanner

import (
	"path"
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
	
	// User-provided patterns
	CustomPatterns []string
	
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
			"*.test.js", "*.test.ts", "*.test.jsx", "*.test.tsx",
			"*.spec.js", "*.spec.ts", "*.spec.jsx", "*.spec.tsx",
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
		CustomPatterns: []string{},
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
	normalized := filepath.ToSlash(filePath)
	fileName := filepath.Base(normalized)
	dir := filepath.Dir(normalized)
	
	// Check Go test patterns
	for _, suffix := range tp.GoTestSuffixes {
		if strings.HasSuffix(fileName, suffix) {
			return true
		}
	}
	
	// Check JavaScript/TypeScript patterns
	for _, pattern := range tp.JSTestPatterns {
		// treat patterns as suffixes (e.g., "*.test.js" => ".test.js")
		if strings.HasPrefix(pattern, "*") {
			suffix := strings.TrimPrefix(pattern, "*")
			if strings.HasSuffix(fileName, suffix) {
				return true
			}
		}
	}
	// common JS test directories (handles root-level and nested)
	safeDir := TestDirSeparator + strings.Trim(dir, TestDirSeparator) + TestDirSeparator
	if strings.Contains(safeDir, TestDirSeparator+"__tests__"+TestDirSeparator) ||
		strings.Contains(safeDir, TestDirSeparator+"__test__"+TestDirSeparator) {
		return true
	}
	
	// Check PHP patterns
	for _, pattern := range tp.PHPTestPatterns {
		target := fileName
		if strings.Contains(pattern, "/") {
			target = normalized
		}
		glob := strings.ReplaceAll(pattern, "**", "*")
		if matched, _ := path.Match(glob, target); matched {
			return true
		}
	}
	
	// Check Python patterns
	for _, pattern := range tp.PythonTestPatterns {
		target := fileName
		if strings.Contains(pattern, "/") {
			target = normalized
		}
		glob := strings.ReplaceAll(pattern, "**", "*")
		if matched, _ := path.Match(glob, target); matched {
			return true
		}
	}
	
	// Check Java patterns
	for _, pattern := range tp.JavaTestPatterns {
		target := fileName
		if strings.Contains(pattern, "/") {
			target = normalized
		}
		glob := strings.ReplaceAll(pattern, "**", "*")
		if matched, _ := path.Match(glob, target); matched {
			return true
		}
	}
	
	// Check .NET patterns
	for _, pattern := range tp.DotNetTestPatterns {
		target := fileName
		if strings.Contains(pattern, "/") {
			target = normalized
		}
		glob := strings.ReplaceAll(pattern, "**", "*")
		if matched, _ := path.Match(glob, target); matched {
			return true
		}
	}
	
	// Check test directories (handles root-level and nested)
	safeDir = TestDirSeparator + strings.Trim(dir, TestDirSeparator) + TestDirSeparator
	for _, testDir := range tp.TestDirectories {
		needle := TestDirSeparator + strings.Trim(testDir, TestDirSeparator) + TestDirSeparator
		if strings.Contains(safeDir, needle) {
			return true
		}
	}

	// Check custom user-provided patterns
	for _, pattern := range tp.CustomPatterns {
		// Treat leading-* patterns as suffixes
		if strings.HasPrefix(pattern, "*") && !strings.Contains(pattern, "/") {
			if strings.HasSuffix(fileName, strings.TrimPrefix(pattern, "*")) {
				return true
			}
			continue
		}
		// Otherwise, try glob matching against path or name depending on whether it contains a directory
		target := fileName
		if strings.Contains(pattern, "/") {
			target = normalized
		}
		// Basic '**' fallback: collapse to '*' (documented approximation)
		glob := strings.ReplaceAll(pattern, "**", "*")
		if matched, _ := path.Match(glob, target); matched {
			return true
		}
	}
	
	return false
}

// AddCustomPatterns allows adding custom test patterns to the existing set
func (tp *TestFilePatterns) AddCustomPatterns(customPatterns []string) {
	// Keep as general patterns (suffix or glob)
	tp.CustomPatterns = append(tp.CustomPatterns, customPatterns...)
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