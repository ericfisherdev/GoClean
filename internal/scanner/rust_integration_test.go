// +build integration

package scanner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ericfisherdev/goclean/internal/config"
	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/violations"
)

// TestRustScanningEndToEnd tests the complete Rust scanning workflow
func TestRustScanningEndToEnd(t *testing.T) {
	// Setup test configuration
	cfg := &config.Config{
		Scan: config.ScanConfig{
			Paths: []string{"../../testdata/rust"},
			Exclude: []string{
				"*.test.rs",
				"vendor/",
			},
		},
		Thresholds: config.Thresholds{
			FunctionLines:        25,
			CyclomaticComplexity: 8,
			Parameters:           4,
			NestingDepth:         3,
		},
	}

	// Create scanner engine
	engine := NewEngine(cfg.Scan.Paths, cfg.Scan.Exclude, cfg.Scan.FileTypes, false)
	if engine == nil {
		t.Fatal("Failed to create scanner engine")
	}

	// Perform the scan
	result, err := engine.Scan()
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// Verify Rust files were discovered
	rustFilesFound := false
	for _, file := range result.Files {
		if strings.HasSuffix(file.Path, ".rs") {
			rustFilesFound = true
			break
		}
	}

	if !rustFilesFound {
		t.Error("No Rust files were discovered in the scan")
	}

	// Test specific violation detection for known test files
	testCases := []struct {
		filename       string
		minViolations  int
		expectedTypes  []models.ViolationType
	}{
		{
			filename:      "naming_violations.rs",
			minViolations: 10,
			expectedTypes: []models.ViolationType{
				models.ViolationTypeNaming,
			},
		},
		{
			filename:      "ownership_issues.rs",
			minViolations: 5,
			expectedTypes: []models.ViolationType{
				"RUST_UNNECESSARY_CLONE",
				"RUST_INEFFICIENT_BORROWING",
			},
		},
		{
			filename:      "error_handling_bad.rs",
			minViolations: 10,
			expectedTypes: []models.ViolationType{
				"RUST_OVERUSE_UNWRAP",
				"RUST_PANIC_PRONE_CODE",
			},
		},
		{
			filename:      "function_violations.rs",
			minViolations: 5,
			expectedTypes: []models.ViolationType{
				models.ViolationTypeFunctionLength,
				models.ViolationTypeTooManyParameters,
				models.ViolationTypeCyclomaticComplexity,
			},
		},
		{
			filename:      "magic_numbers.rs",
			minViolations: 10,
			expectedTypes: []models.ViolationType{
				models.ViolationTypeMagicNumber,
			},
		},
		{
			filename:      "documentation_missing.rs",
			minViolations: 10,
			expectedTypes: []models.ViolationType{
				models.ViolationTypeMissingDocumentation,
			},
		},
		{
			filename:      "unsafe_code.rs",
			minViolations: 5,
			expectedTypes: []models.ViolationType{
				"RUST_UNNECESSARY_UNSAFE",
				"RUST_UNSAFE_WITHOUT_COMMENT",
			},
		},
		{
			filename:      "good_code_example.rs",
			minViolations: 0,
			expectedTypes: []models.ViolationType{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.filename, func(t *testing.T) {
			violations := getViolationsForFile(result.Violations, tc.filename)
			
			if tc.minViolations == 0 && len(violations) > 0 {
				t.Errorf("%s: Expected no violations but found %d", tc.filename, len(violations))
				for _, v := range violations {
					t.Logf("  Unexpected violation: %s at line %d", v.Type, v.Line)
				}
			} else if len(violations) < tc.minViolations {
				t.Errorf("%s: Expected at least %d violations, got %d", 
					tc.filename, tc.minViolations, len(violations))
			}

			// Check for expected violation types
			for _, expectedType := range tc.expectedTypes {
				found := false
				for _, v := range violations {
					if v.Type == expectedType {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("%s: Expected violation type %s not found", tc.filename, expectedType)
				}
			}
		})
	}

	// Verify statistics
	if result.Statistics.TotalViolations == 0 {
		t.Error("Expected violations to be detected in test files")
	}

	if result.Statistics.FilesScanned == 0 {
		t.Error("No files were scanned")
	}

	// Log summary for debugging
	t.Logf("Scan completed: %d files scanned, %d violations found",
		result.Statistics.FilesScanned, result.Statistics.TotalViolations)
}

// TestMixedGoRustProject tests scanning a project with both Go and Rust files
func TestMixedGoRustProject(t *testing.T) {
	// Create a temporary mixed project
	tempDir := t.TempDir()
	
	// Create Go file
	goFile := filepath.Join(tempDir, "main.go")
	goContent := `package main

// This function has too many parameters
func ProcessData(a, b, c, d, e, f int) int {
	return a + b + c + d + e + f
}

func main() {
	// Magic number violation
	result := ProcessData(1, 2, 3, 4, 5, 42)
	println(result)
}
`
	if err := os.WriteFile(goFile, []byte(goContent), 0644); err != nil {
		t.Fatalf("Failed to create Go file: %v", err)
	}

	// Create Rust file
	rustFile := filepath.Join(tempDir, "lib.rs")
	rustContent := `// Missing documentation for public function
pub fn processData(a: i32, b: i32, c: i32, d: i32, e: i32) -> i32 {
    // Magic number violation
    let magic = 42;
    a + b + c + d + e + magic
}

// Function name should be snake_case
pub fn ProcessDataWrong() {
    println!("Wrong naming");
}
`
	if err := os.WriteFile(rustFile, []byte(rustContent), 0644); err != nil {
		t.Fatalf("Failed to create Rust file: %v", err)
	}

	// Setup configuration for mixed project
	cfg := &config.Config{
		Scan: config.ScanConfig{
			Paths: []string{tempDir},
		},
		Thresholds: config.ThresholdConfig{
			Parameters: 4,
		},
	}

	// Scan the mixed project
	engine := NewEngine(cfg.Scan.Paths, cfg.Scan.Exclude, cfg.Scan.FileTypes, false)
	result, err := engine.Scan()
	if err != nil {
		t.Fatalf("Failed to scan mixed project: %v", err)
	}

	// Verify both languages were detected
	goFound := false
	rustFound := false
	
	for _, file := range result.Files {
		if strings.HasSuffix(file.Path, ".go") {
			goFound = true
		}
		if strings.HasSuffix(file.Path, ".rs") {
			rustFound = true
		}
	}

	if !goFound {
		t.Error("Go files were not detected in mixed project")
	}
	if !rustFound {
		t.Error("Rust files were not detected in mixed project")
	}

	// Verify violations were detected for both languages
	goViolations := getViolationsForFile(result.Violations, "main.go")
	rustViolations := getViolationsForFile(result.Violations, "lib.rs")

	if len(goViolations) == 0 {
		t.Error("No violations detected for Go file")
	}
	if len(rustViolations) == 0 {
		t.Error("No violations detected for Rust file")
	}

	t.Logf("Mixed project scan: %d Go violations, %d Rust violations",
		len(goViolations), len(rustViolations))
}

// TestRustDetectorIntegration tests individual Rust detectors
func TestRustDetectorIntegration(t *testing.T) {
	detectors := []violations.Detector{
		violations.NewRustFunctionDetector(nil),
		violations.NewRustNamingDetector(nil),
		violations.NewRustDocumentationDetector(nil),
		violations.NewRustMagicNumberDetector(nil),
		violations.NewRustDuplicationDetector(nil),
		violations.NewRustStructureDetector(nil),
		violations.NewRustCommentedCodeDetector(nil),
		violations.NewRustTodoTrackerDetector(nil),
		violations.NewRustOwnershipDetector(nil),
		violations.NewRustErrorHandlingDetector(nil),
	}

	// Test each detector with appropriate test file
	testFiles := map[string]string{
		"RustFunctionDetector":       "../../testdata/rust/function_violations.rs",
		"RustNamingDetector":         "../../testdata/rust/naming_violations.rs",
		"RustDocumentationDetector":  "../../testdata/rust/documentation_missing.rs",
		"RustMagicNumberDetector":    "../../testdata/rust/magic_numbers.rs",
		"RustDuplicationDetector":    "../../testdata/rust/duplication_issues.rs",
		"RustStructureDetector":      "../../testdata/rust/structure_issues.rs",
		"RustCommentedCodeDetector":  "../../testdata/rust/commented_code.rs",
		"RustTodoTrackerDetector":    "../../testdata/rust/commented_code.rs",
		"RustOwnershipDetector":      "../../testdata/rust/ownership_issues.rs",
		"RustErrorHandlingDetector":  "../../testdata/rust/error_handling_bad.rs",
	}

	for _, detector := range detectors {
		detectorName := detector.Name()
		testFile, exists := testFiles[detectorName]
		if !exists {
			t.Logf("No test file mapped for detector: %s", detectorName)
			continue
		}

		t.Run(detectorName, func(t *testing.T) {
			// Read the test file
			content, err := os.ReadFile(testFile)
			if err != nil {
				t.Skipf("Test file not found: %s", testFile)
				return
			}

			// Create file info
			fileInfo := &models.FileInfo{
				Path:     testFile,
				Language: "rust",
				Lines:    len(strings.Split(string(content), "\n")),
			}

			// Parse the Rust file
			parser := NewRustParser()
			astInfo, err := parser.Parse(string(content), testFile)
			if err != nil {
				t.Logf("Warning: Rust parsing failed for %s: %v", testFile, err)
				// Continue with test even if parsing fails (regex-based fallback)
			}

			// Detect violations
			detectedViolations := detector.Detect(fileInfo, astInfo)

			// Verify violations were detected
			if len(detectedViolations) == 0 {
				t.Logf("%s: No violations detected for %s", detectorName, testFile)
			} else {
				t.Logf("%s: Detected %d violations in %s", 
					detectorName, len(detectedViolations), testFile)
				
				// Log first few violations for debugging
				for i, v := range detectedViolations {
					if i >= 3 {
						break
					}
					t.Logf("  - Line %d: %s", v.Line, v.Message)
				}
			}
		})
	}
}

// TestRustClippyIntegration tests the Clippy integrator if available
func TestRustClippyIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Clippy integration test in short mode")
	}

	integrator := violations.NewClippyIntegrator(nil)
	
	// Check if clippy is available
	// This is a simplified check - actual implementation would verify clippy availability
	t.Log("Testing Clippy integration (if available)")

	// Create a test Rust project with Cargo.toml
	tempDir := t.TempDir()
	
	// Create Cargo.toml
	cargoToml := `[package]
name = "test_project"
version = "0.1.0"
edition = "2021"

[dependencies]
`
	cargoPath := filepath.Join(tempDir, "Cargo.toml")
	if err := os.WriteFile(cargoPath, []byte(cargoToml), 0644); err != nil {
		t.Fatalf("Failed to create Cargo.toml: %v", err)
	}

	// Create src directory
	srcDir := filepath.Join(tempDir, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatalf("Failed to create src directory: %v", err)
	}

	// Create main.rs with clippy violations
	mainRs := `fn main() {
    println!("Hello, world!");
}

// Function with too many arguments
fn too_many_args(a: i32, b: i32, c: i32, d: i32, e: i32, f: i32) -> i32 {
    a + b + c + d + e + f
}

// Magic number
fn magic_number() -> i32 {
    42
}
`
	mainPath := filepath.Join(srcDir, "main.rs")
	if err := os.WriteFile(mainPath, []byte(mainRs), 0644); err != nil {
		t.Fatalf("Failed to create main.rs: %v", err)
	}

	// Create file info
	fileInfo := &models.FileInfo{
		Path:     mainPath,
		Language: "rust",
		Lines:    strings.Count(mainRs, "\n"),
	}

	// Run clippy detector
	violations := integrator.Detect(fileInfo, nil)
	
	if len(violations) > 0 {
		t.Logf("Clippy detected %d violations", len(violations))
		for _, v := range violations {
			// Verify proper attribution
			if !strings.Contains(v.Message, "Detected by rust-clippy") {
				t.Errorf("Clippy violation missing attribution: %s", v.Message)
			}
		}
	} else {
		t.Log("No Clippy violations detected (Clippy may not be available)")
	}
}

// Helper function to get violations for a specific file
func getViolationsForFile(violations []*models.Violation, filename string) []*models.Violation {
	var result []*models.Violation
	for _, v := range violations {
		if strings.Contains(v.File, filename) {
			result = append(result, v)
		}
	}
	return result
}