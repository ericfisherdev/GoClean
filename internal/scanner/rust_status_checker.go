package scanner

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// RustStatusChecker provides utilities for checking Rust parsing status
type RustStatusChecker struct {
	manager *RustParserManager
}

// NewRustStatusChecker creates a new status checker
func NewRustStatusChecker() *RustStatusChecker {
	return &RustStatusChecker{
		manager: GetGlobalParserManager(false),
	}
}

// getFloat64FromMap safely extracts a float64 value from a map
func getFloat64FromMap(m map[string]interface{}, key string) float64 {
	val, ok := m[key]
	if !ok {
		return 0.0
	}
	
	switch v := val.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case int32:
		return float64(v)
	case uint:
		return float64(v)
	case uint64:
		return float64(v)
	case uint32:
		return float64(v)
	default:
		return 0.0
	}
}

// getBoolFromMap safely extracts a bool value from a map
func getBoolFromMap(m map[string]interface{}, key string) bool {
	val, ok := m[key]
	if !ok {
		return false
	}
	
	boolVal, ok := val.(bool)
	if !ok {
		return false
	}
	return boolVal
}

// CheckStatus performs a comprehensive status check
func (c *RustStatusChecker) CheckStatus() *RustStatusReport {
	capabilities := c.manager.GetCapabilities()
	status := c.manager.GetStatus()

	report := &RustStatusReport{
		Timestamp:       time.Now(),
		ParserType:      capabilities.ParserType,
		FallbackReason:  capabilities.FallbackReason,
		CGOEnabled:      getBoolFromMap(status, "cgo_enabled"),
		RustAvailable:   getBoolFromMap(status, "rust_available"),
		AccuracyLevel:   capabilities.AccuracyLevel,
		PerformanceLevel: capabilities.PerformanceLevel,
		Features:        c.getFeatureStatus(capabilities),
		Recommendations: c.generateRecommendations(capabilities, status),
		TechnicalDetails: status,
	}

	// Test parsing capability
	c.testParsingCapability(report)

	return report
}

// getFeatureStatus returns the status of various features
func (c *RustStatusChecker) getFeatureStatus(capabilities *RustParserCapabilities) map[string]bool {
	return map[string]bool{
		"syntax_validation":   capabilities.HasSyntaxValidation,
		"expression_parsing":  capabilities.HasExpressionParsing,
		"full_ast_parsing":    capabilities.ParserType == "syn-crate",
		"basic_ast_parsing":   capabilities.ParserType == "regex-fallback",
		"performance_optimal": capabilities.PerformanceLevel == "optimal",
		"high_accuracy":       capabilities.AccuracyLevel == "high",
	}
}

// generateRecommendations provides recommendations based on current status
func (c *RustStatusChecker) generateRecommendations(capabilities *RustParserCapabilities, status map[string]interface{}) []string {
	var recommendations []string

	switch capabilities.ParserType {
	case "syn-crate":
		recommendations = append(recommendations, "✅ Optimal Rust parsing is active. No action needed.")

	case "regex-fallback":
		recommendations = append(recommendations,
			"⚠️  Using basic regex parsing. For better accuracy:",
			"  • Install Rust toolchain: curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh",
			"  • Build with CGO enabled: CGO_ENABLED=1 go build",
			"  • Ensure goclean-rust-parser library is built",
		)

	case "no-op-fallback":
		recommendations = append(recommendations,
			"❌ Rust parsing is not available. To enable:",
			"  • Install Rust toolchain: curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh",
			"  • Enable CGO: export CGO_ENABLED=1",
			"  • Build the Rust parser library: cd ../playground/goclean-rust-parser && cargo build --release",
			"  • Ensure library is in the correct path",
			"  • Rebuild GoClean: make build",
		)
	}

	// Add CGO-specific recommendations
	if !getBoolFromMap(status, "cgo_enabled") {
		recommendations = append(recommendations,
			"ℹ️  CGO is disabled. To enable full Rust support:",
			"  • Set CGO_ENABLED=1 environment variable",
			"  • Rebuild with: CGO_ENABLED=1 go build",
		)
	}

	// Add performance recommendations
	successRate := getFloat64FromMap(status, "success_rate")
	if successRate < 90.0 && successRate > 0 {
		recommendations = append(recommendations,
			fmt.Sprintf("⚠️  Low success rate (%.1f%%). Consider checking:", successRate),
			"  • File permissions and paths",
			"  • Rust code syntax validity",
			"  • Parser library stability",
		)
	}

	return recommendations
}

// testParsingCapability tests parsing with a simple Rust example
func (c *RustStatusChecker) testParsingCapability(report *RustStatusReport) {
	testCode := `
fn test_function() -> i32 {
    42
}

struct TestStruct {
    field: String,
}
`

	start := time.Now()
	astInfo, err := c.manager.ParseRustFile([]byte(testCode), "status_test.rs")
	duration := time.Since(start)

	report.TestResult = &ParseTestResult{
		Success:     err == nil,
		Duration:    duration,
		Error:       "",
		FunctionsFound: 0,
		StructsFound:   0,
	}

	if err != nil {
		report.TestResult.Error = err.Error()
	} else if astInfo != nil {
		report.TestResult.FunctionsFound = len(astInfo.Functions)
		report.TestResult.StructsFound = len(astInfo.Structs)
	}
}

// RustStatusReport contains comprehensive status information
type RustStatusReport struct {
	Timestamp        time.Time              `json:"timestamp"`
	ParserType       string                 `json:"parser_type"`
	FallbackReason   string                 `json:"fallback_reason,omitempty"`
	CGOEnabled       bool                   `json:"cgo_enabled"`
	RustAvailable    bool                   `json:"rust_available"`
	AccuracyLevel    string                 `json:"accuracy_level"`
	PerformanceLevel string                 `json:"performance_level"`
	Features         map[string]bool        `json:"features"`
	Recommendations  []string               `json:"recommendations"`
	TestResult       *ParseTestResult       `json:"test_result,omitempty"`
	TechnicalDetails map[string]interface{} `json:"technical_details,omitempty"`
}

// ParseTestResult contains results from parsing capability test
type ParseTestResult struct {
	Success        bool          `json:"success"`
	Duration       time.Duration `json:"duration"`
	Error          string        `json:"error,omitempty"`
	FunctionsFound int           `json:"functions_found"`
	StructsFound   int           `json:"structs_found"`
}

// PrintReport prints a human-readable status report
func (c *RustStatusChecker) PrintReport(report *RustStatusReport) {
	fmt.Printf("\n🦀 GoClean Rust Support Status Report\n")
	fmt.Printf("=====================================\n\n")

	// Basic status
	fmt.Printf("📋 Basic Information:\n")
	fmt.Printf("  • Timestamp: %s\n", report.Timestamp.Format(time.RFC3339))
	fmt.Printf("  • Parser Type: %s\n", report.ParserType)
	fmt.Printf("  • CGO Enabled: %v\n", report.CGOEnabled)
	fmt.Printf("  • Rust Available: %v\n", report.RustAvailable)
	fmt.Printf("  • Accuracy Level: %s\n", report.AccuracyLevel)
	fmt.Printf("  • Performance Level: %s\n", report.PerformanceLevel)

	if report.FallbackReason != "" {
		fmt.Printf("  • Fallback Reason: %s\n", report.FallbackReason)
	}

	// Features
	fmt.Printf("\n🔧 Feature Support:\n")
	for feature, supported := range report.Features {
		status := "❌"
		if supported {
			status = "✅"
		}
		fmt.Printf("  %s %s\n", status, strings.ReplaceAll(feature, "_", " "))
	}

	// Test results
	if report.TestResult != nil {
		fmt.Printf("\n🧪 Parse Test Results:\n")
		if report.TestResult.Success {
			fmt.Printf("  ✅ Test passed in %v\n", report.TestResult.Duration)
			fmt.Printf("  • Found %d functions, %d structs\n", 
				report.TestResult.FunctionsFound, report.TestResult.StructsFound)
		} else {
			fmt.Printf("  ❌ Test failed: %s\n", report.TestResult.Error)
		}
	}

	// Recommendations
	fmt.Printf("\n💡 Recommendations:\n")
	for _, rec := range report.Recommendations {
		fmt.Printf("  %s\n", rec)
	}

	fmt.Printf("\n")
}

// PrintJSONReport prints the report in JSON format
func (c *RustStatusChecker) PrintJSONReport(report *RustStatusReport) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

// PrintCompactReport prints a compact one-line status
func (c *RustStatusChecker) PrintCompactReport(report *RustStatusReport) {
	var statusIcon string
	switch report.ParserType {
	case "syn-crate":
		statusIcon = "✅"
	case "regex-fallback":
		statusIcon = "⚠️ "
	default:
		statusIcon = "❌"
	}

	fmt.Printf("%s Rust parsing: %s (%s accuracy, %s performance)\n",
		statusIcon, report.ParserType, report.AccuracyLevel, report.PerformanceLevel)

	if report.TestResult != nil && report.TestResult.Success {
		fmt.Printf("   Test: ✅ %v (%d functions, %d structs found)\n",
			report.TestResult.Duration, report.TestResult.FunctionsFound, report.TestResult.StructsFound)
	} else if report.TestResult != nil {
		fmt.Printf("   Test: ❌ %s\n", report.TestResult.Error)
	}
}

// GetRecommendationsForCI returns CI-friendly recommendations
func (c *RustStatusChecker) GetRecommendationsForCI(report *RustStatusReport) []string {
	var ciRecommendations []string

	if report.ParserType != "syn-crate" {
		ciRecommendations = append(ciRecommendations,
			"Consider enabling optimal Rust parsing for better code analysis:",
			"1. Install Rust toolchain in CI environment",
			"2. Set CGO_ENABLED=1 in build environment",
			"3. Build with rust parser dependencies",
			"4. Cache built libraries for faster subsequent builds",
		)
	}

	if !report.RustAvailable {
		ciRecommendations = append(ciRecommendations,
			"Rust parsing is completely unavailable in this build.",
			"Consider creating separate build configurations for different environments.",
		)
	}

	return ciRecommendations
}

// GetUpgradeInstructions returns detailed upgrade instructions
func (c *RustStatusChecker) GetUpgradeInstructions() []string {
	return []string{
		"📋 Complete Rust Support Upgrade Instructions:",
		"",
		"1️⃣  Install Rust toolchain:",
		"   curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh",
		"   source ~/.cargo/env",
		"",
		"2️⃣  Verify Rust installation:",
		"   rustc --version",
		"   cargo --version",
		"",
		"3️⃣  Build the Rust parser library:",
		"   cd ../playground/goclean-rust-parser",
		"   cargo build --release --crate-type cdylib",
		"",
		"4️⃣  Set environment variables:",
		"   export CGO_ENABLED=1",
		"   export LIBRARY_PATH=/path/to/rust/library",
		"",
		"5️⃣  Rebuild GoClean:",
		"   make clean",
		"   make build",
		"",
		"6️⃣  Verify installation:",
		"   ./bin/goclean --rust-status",
		"",
		"🔍 Troubleshooting:",
		"   • Check library paths with: ldd ./bin/goclean",
		"   • Verify CGO linking with: go build -x",
		"   • Test manually with: go test ./internal/scanner -run RustSyn",
	}
}

// CheckEnvironment performs environment checks for Rust support
func (c *RustStatusChecker) CheckEnvironment() *EnvironmentReport {
	report := &EnvironmentReport{
		Timestamp: time.Now(),
		Checks:    make(map[string]EnvironmentCheck),
	}

	// Check CGO
	report.Checks["cgo"] = EnvironmentCheck{
		Name:        "CGO Support",
		Available:   CGOEnabled,
		Description: "C/Go interoperability for Rust library integration",
		Requirement: "Required for optimal Rust parsing",
	}

	// Check for Rust toolchain (attempt to detect)
	report.Checks["rust_toolchain"] = c.checkRustToolchain()

	// Check library availability
	report.Checks["rust_library"] = c.checkRustLibrary()

	// Check parser functionality
	report.Checks["parser_function"] = c.checkParserFunction()

	return report
}

// EnvironmentReport contains environment check results
type EnvironmentReport struct {
	Timestamp time.Time                   `json:"timestamp"`
	Checks    map[string]EnvironmentCheck `json:"checks"`
	Overall   string                      `json:"overall_status"`
}

// EnvironmentCheck represents a single environment check
type EnvironmentCheck struct {
	Name        string `json:"name"`
	Available   bool   `json:"available"`
	Description string `json:"description"`
	Requirement string `json:"requirement"`
	Details     string `json:"details,omitempty"`
}

// checkRustToolchain attempts to detect Rust toolchain
func (c *RustStatusChecker) checkRustToolchain() EnvironmentCheck {
	// This is a basic check - in a real implementation, you might try to execute rustc --version
	return EnvironmentCheck{
		Name:        "Rust Toolchain",
		Available:   false, // Conservatively assume not available
		Description: "Rust compiler and cargo package manager",
		Requirement: "Required for building Rust parser library",
		Details:     "Cannot detect Rust from Go code - manual verification needed",
	}
}

// checkRustLibrary checks for the presence of the Rust parser library
func (c *RustStatusChecker) checkRustLibrary() EnvironmentCheck {
	// Try to initialize syn parser to test library availability
	_, err := NewRustSynParser(false)
	available := err == nil

	check := EnvironmentCheck{
		Name:        "Rust Parser Library",
		Available:   available,
		Description: "goclean-rust-parser library for syn crate integration",
		Requirement: "Required for optimal Rust parsing accuracy",
	}

	if available {
		check.Details = "Library loaded successfully"
	} else {
		check.Details = fmt.Sprintf("Library not available: %v", err)
	}

	return check
}

// checkParserFunction tests basic parser functionality
func (c *RustStatusChecker) checkParserFunction() EnvironmentCheck {
	testCode := `fn test() {}`
	_, err := c.manager.ParseRustFile([]byte(testCode), "env_test.rs")

	check := EnvironmentCheck{
		Name:        "Parser Functionality",
		Available:   err == nil,
		Description: "Basic Rust code parsing capability",
		Requirement: "Required for any Rust code analysis",
	}

	if err == nil {
		check.Details = "Basic parsing test successful"
	} else {
		check.Details = fmt.Sprintf("Parsing test failed: %v", err)
	}

	return check
}

// PrintEnvironmentReport prints environment check results
func (c *RustStatusChecker) PrintEnvironmentReport(report *EnvironmentReport) {
	fmt.Printf("\n🔍 GoClean Rust Environment Check\n")
	fmt.Printf("=================================\n\n")

	allGood := true
	for _, check := range report.Checks {
		if !check.Available {
			allGood = false
		}

		status := "❌"
		if check.Available {
			status = "✅"
		}

		fmt.Printf("%s %s\n", status, check.Name)
		fmt.Printf("   %s\n", check.Description)
		fmt.Printf("   Requirement: %s\n", check.Requirement)
		if check.Details != "" {
			fmt.Printf("   Details: %s\n", check.Details)
		}
		fmt.Printf("\n")
	}

	if allGood {
		fmt.Printf("🎉 All environment checks passed! Rust support should be fully functional.\n")
	} else {
		fmt.Printf("⚠️  Some environment checks failed. See recommendations above.\n")
	}
}