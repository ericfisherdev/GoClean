// Package violations provides detectors for various clean code violations in Rust source code.
package violations

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/types"
)

// Rust-specific constants for magic number detection
const (
	// Numeric ranges for acceptable values
	rustSmallNumberLimit = 10
	rustPiLowerBound     = 3.1
	rustPiUpperBound     = 3.15
	rustELowerBound      = 2.7
	rustEUpperBound      = 2.72
)

// RustMagicNumberDetector detects hardcoded magic numbers in Rust code
type RustMagicNumberDetector struct {
	config        *DetectorConfig
	codeExtractor *CodeExtractor
	// Compiled regex patterns for efficient matching
	numericPattern    *regexp.Regexp
	integerPattern    *regexp.Regexp
	floatPattern      *regexp.Regexp
	hexPattern        *regexp.Regexp
	binaryPattern     *regexp.Regexp
	constantPattern   *regexp.Regexp
}

// NewRustMagicNumberDetector creates a new Rust magic number detector
func NewRustMagicNumberDetector(config *DetectorConfig) *RustMagicNumberDetector {
	if config == nil {
		config = DefaultDetectorConfig()
	}

	return &RustMagicNumberDetector{
		config:        config,
		codeExtractor: NewCodeExtractor(),
		// Rust numeric literal patterns
		numericPattern:  regexp.MustCompile(`\b(\d+(?:\.\d+)?(?:[eE][+-]?\d+)?(?:f32|f64|u8|u16|u32|u64|u128|i8|i16|i32|i64|i128|usize|isize)?)\b`),
		integerPattern:  regexp.MustCompile(`\b(\d+(?:u8|u16|u32|u64|u128|i8|i16|i32|i64|i128|usize|isize)?)\b`),
		floatPattern:    regexp.MustCompile(`\b(\d+\.\d+(?:[eE][+-]?\d+)?(?:f32|f64)?)\b`),
		hexPattern:      regexp.MustCompile(`\b(0x[0-9a-fA-F]+(?:u8|u16|u32|u64|u128|i8|i16|i32|i64|i128|usize|isize)?)\b`),
		binaryPattern:   regexp.MustCompile(`\b(0b[01]+(?:u8|u16|u32|u64|u128|i8|i16|i32|i64|i128|usize|isize)?)\b`),
		constantPattern: regexp.MustCompile(`(?i)\b(const\s+[A-Z_][A-Z0-9_]*\s*:\s*\w+\s*=|static\s+[A-Z_][A-Z0-9_]*\s*:\s*\w+\s*=)`),
	}
}

// Name returns the name of this detector
func (d *RustMagicNumberDetector) Name() string {
	return "Rust Magic Number Detector"
}

// Description returns a description of what this detector checks for
func (d *RustMagicNumberDetector) Description() string {
	return "Detects hardcoded numeric literals in Rust code that should be named constants"
}

// Detect analyzes the provided Rust file information and returns violations
func (d *RustMagicNumberDetector) Detect(fileInfo *models.FileInfo, astInfo interface{}) []*models.Violation {
	var violations []*models.Violation

	if astInfo == nil {
		return violations
	}

	// Type assertion to get types.RustASTInfo
	rustAstInfo, ok := astInfo.(*types.RustASTInfo)
	if !ok || rustAstInfo == nil {
		return violations
	}

	// Read the file content for regex-based analysis
	// This is simpler than full AST analysis for magic numbers
	content, err := d.readFileContent(fileInfo.Path)
	if err != nil {
		return violations
	}

	// Analyze numeric literals in the file content
	violations = append(violations, d.analyzeNumericLiterals(content, fileInfo.Path)...)

	return violations
}

// readFileContent reads the content of a file
func (d *RustMagicNumberDetector) readFileContent(filePath string) (string, error) {
	if d.codeExtractor == nil {
		return "", fmt.Errorf("code extractor not available")
	}

	// Use the code extractor to read the file
	content, err := d.codeExtractor.ExtractSnippet(filePath, 1, -1) // Read entire file
	if err != nil {
		return "", err
	}

	return content, nil
}

// analyzeNumericLiterals analyzes numeric literals in the file content
func (d *RustMagicNumberDetector) analyzeNumericLiterals(content, filePath string) []*models.Violation {
	var violations []*models.Violation

	lines := strings.Split(content, "\n")

	for lineNum, line := range lines {
		lineNumber := lineNum + 1

		// Skip comment lines and constant declarations
		if d.shouldSkipLine(line) {
			continue
		}

		// Find all numeric literals in the line
		violations = append(violations, d.findNumericLiteralsInLine(line, lineNumber, filePath)...)
	}

	return violations
}

// shouldSkipLine determines if a line should be skipped from magic number analysis
func (d *RustMagicNumberDetector) shouldSkipLine(line string) bool {
	trimmed := strings.TrimSpace(line)

	// Skip comment lines
	if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") {
		return true
	}

	// Skip constant and static declarations
	if d.constantPattern.MatchString(trimmed) {
		return true
	}

	// Skip version numbers and other specific patterns
	if strings.Contains(trimmed, "version") || 
	   strings.Contains(trimmed, "= \"") || 
	   strings.Contains(trimmed, "= '") {
		return true
	}

	// Skip test files with test data
	if strings.Contains(trimmed, "#[test]") || 
	   strings.Contains(trimmed, "assert_") ||
	   strings.Contains(trimmed, "expect(") {
		return true
	}

	return false
}

// findNumericLiteralsInLine finds numeric literals in a line and checks if they are magic numbers
func (d *RustMagicNumberDetector) findNumericLiteralsInLine(line string, lineNumber int, filePath string) []*models.Violation {
	var violations []*models.Violation

	// Find all numeric patterns in the line
	allMatches := d.findAllNumericMatches(line)

	for _, match := range allMatches {
		if violation := d.checkRustMagicNumber(match, lineNumber, filePath, line); violation != nil {
			violations = append(violations, violation)
		}
	}

	return violations
}

// findAllNumericMatches finds all numeric literal matches in a line
func (d *RustMagicNumberDetector) findAllNumericMatches(line string) []NumericMatch {
	var matches []NumericMatch

	// Check integer literals
	for _, match := range d.integerPattern.FindAllStringSubmatch(line, -1) {
		if len(match) > 1 {
			matches = append(matches, NumericMatch{
				Value:   match[1],
				Type:    "integer",
				Content: line,
			})
		}
	}

	// Check float literals
	for _, match := range d.floatPattern.FindAllStringSubmatch(line, -1) {
		if len(match) > 1 {
			matches = append(matches, NumericMatch{
				Value:   match[1],
				Type:    "float",
				Content: line,
			})
		}
	}

	// Check hex literals
	for _, match := range d.hexPattern.FindAllStringSubmatch(line, -1) {
		if len(match) > 1 {
			matches = append(matches, NumericMatch{
				Value:   match[1],
				Type:    "hex",
				Content: line,
			})
		}
	}

	// Check binary literals
	for _, match := range d.binaryPattern.FindAllStringSubmatch(line, -1) {
		if len(match) > 1 {
			matches = append(matches, NumericMatch{
				Value:   match[1],
				Type:    "binary",
				Content: line,
			})
		}
	}

	return matches
}

// NumericMatch represents a found numeric literal
type NumericMatch struct {
	Value   string
	Type    string
	Content string
}

// checkRustMagicNumber checks if a numeric literal is a magic number
func (d *RustMagicNumberDetector) checkRustMagicNumber(match NumericMatch, lineNumber int, filePath, lineContent string) *models.Violation {
	value := match.Value

	// Check if it's in an acceptable context
	if d.isInRustAcceptableContext(match, lineContent) {
		return nil
	}

	// Parse and check the numeric value
	if d.isRustAcceptableValue(value, match.Type) {
		return nil
	}

	// Create violation
	return &models.Violation{
		Type:        models.ViolationTypeMagicNumber,
		Severity:    d.getRustMagicNumberSeverity(value, match.Type),
		Message:     fmt.Sprintf("Magic number '%s' detected in Rust code", value),
		File:        filePath,
		Line:        lineNumber,
		Column:      strings.Index(lineContent, value) + 1,
		Rule:        "rust-magic-number",
		Suggestion:  d.getRustMagicNumberSuggestion(value, match.Type),
		CodeSnippet: strings.TrimSpace(lineContent),
	}
}

// isInRustAcceptableContext checks if a numeric literal is in an acceptable context
func (d *RustMagicNumberDetector) isInRustAcceptableContext(match NumericMatch, lineContent string) bool {
	line := strings.ToLower(lineContent)

	// Array/vector indexing
	if strings.Contains(line, "[") && strings.Contains(line, "]") {
		return true
	}

	// Range operations
	if strings.Contains(line, "..") || strings.Contains(line, "...") {
		return true
	}

	// Match arms and patterns
	if strings.Contains(line, "match") || strings.Contains(line, "=>") {
		return true
	}

	// Size and capacity hints
	if strings.Contains(line, "with_capacity") || 
	   strings.Contains(line, "reserve") ||
	   strings.Contains(line, "resize") {
		return true
	}

	// Time and duration context
	if strings.Contains(line, "duration") || 
	   strings.Contains(line, "timeout") ||
	   strings.Contains(line, "sleep") ||
	   strings.Contains(line, "delay") {
		return true
	}

	// Network and port context
	if strings.Contains(line, "port") || 
	   strings.Contains(line, "address") ||
	   strings.Contains(line, "socket") {
		return true
	}

	// Buffer sizes and memory allocation
	if strings.Contains(line, "buffer") || 
	   strings.Contains(line, "vec!") ||
	   strings.Contains(line, "allocate") {
		return true
	}

	// Bit manipulation context
	if strings.Contains(line, "<<") || 
	   strings.Contains(line, ">>") ||
	   strings.Contains(line, "&") ||
	   strings.Contains(line, "|") ||
	   strings.Contains(line, "^") {
		return true
	}

	// Variable names that suggest the number is descriptive
	descriptivePatterns := []string{
		"count", "size", "len", "length", "capacity", "limit", "max", "min",
		"width", "height", "depth", "radius", "diameter", "area", "volume",
		"weight", "mass", "density", "pressure", "temperature", "speed",
		"frequency", "rate", "ratio", "percentage", "factor", "multiplier",
		"offset", "index", "position", "coordinate", "threshold", "tolerance",
	}

	for _, pattern := range descriptivePatterns {
		if strings.Contains(line, pattern) {
			return true
		}
	}

	return false
}

// isRustAcceptableValue checks if a numeric value is commonly acceptable in Rust
func (d *RustMagicNumberDetector) isRustAcceptableValue(value, valueType string) bool {
	switch valueType {
	case "integer":
		return d.isRustAcceptableInt(value)
	case "float":
		return d.isRustAcceptableFloat(value)
	case "hex":
		return d.isRustAcceptableHex(value)
	case "binary":
		return d.isRustAcceptableBinary(value)
	default:
		return false
	}
}

// isRustAcceptableInt checks if an integer value is commonly acceptable
func (d *RustMagicNumberDetector) isRustAcceptableInt(value string) bool {
	// Remove type suffix for parsing
	cleanValue := d.removeRustTypeSuffix(value)
	
	intVal, err := strconv.Atoi(cleanValue)
	if err != nil {
		return false
	}

	// Common acceptable values
	acceptableInts := []int{
		-1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, // Very common small values
		16, 24, 32, 60, 64, // Common bit sizes and time values
		100, 128, 256, 512, 1024, 2048, 4096, 8192, // Powers of 2 and 100
	}

	for _, acceptable := range acceptableInts {
		if intVal == acceptable {
			return true
		}
	}

	// Check if it's a small number (common for array indices, small loops)
	if intVal >= 0 && intVal <= rustSmallNumberLimit {
		return true
	}

	// Check if it's a power of 2 (common in systems programming)
	if d.isRustPowerOfTwo(intVal) {
		return true
	}

	// Check if it's a power of 10
	if d.isRustPowerOfTen(intVal) {
		return true
	}

	return false
}

// isRustAcceptableFloat checks if a float value is commonly acceptable
func (d *RustMagicNumberDetector) isRustAcceptableFloat(value string) bool {
	// Remove type suffix for parsing
	cleanValue := d.removeRustTypeSuffix(value)
	
	floatVal, err := strconv.ParseFloat(cleanValue, 64)
	if err != nil {
		return false
	}

	// Common acceptable float values
	acceptableFloats := []float64{
		0.0, 1.0, 2.0, 3.0, 4.0, 5.0, // Basic values
		0.5, 0.25, 0.75, 0.1, 0.01, 0.001, // Common fractions and decimals
		3.14, 3.14159, 3.141592653589793, // Pi approximations
		2.71828, 2.718281828459045, // e approximations
	}

	for _, acceptable := range acceptableFloats {
		if floatVal == acceptable {
			return true
		}
	}

	// Check for close approximations of Pi
	if floatVal > rustPiLowerBound && floatVal < rustPiUpperBound {
		return true
	}

	// Check for close approximations of e
	if floatVal > rustELowerBound && floatVal < rustEUpperBound {
		return true
	}

	return false
}

// isRustAcceptableHex checks if a hex value is commonly acceptable
func (d *RustMagicNumberDetector) isRustAcceptableHex(value string) bool {
	// Common hex values that are typically acceptable
	acceptableHex := []string{
		"0x0", "0x1", "0x2", "0x4", "0x8", "0x10", "0x20", "0x40", "0x80",
		"0x100", "0x200", "0x400", "0x800", "0x1000", "0x2000", "0x4000", "0x8000",
		"0xff", "0xffff", "0xffffffff", "0xffffffffffffffff", // Common bit masks
	}

	cleanValue := strings.ToLower(d.removeRustTypeSuffix(value))
	for _, acceptable := range acceptableHex {
		if cleanValue == acceptable {
			return true
		}
	}

	return false
}

// isRustAcceptableBinary checks if a binary value is commonly acceptable
func (d *RustMagicNumberDetector) isRustAcceptableBinary(value string) bool {
	// Common binary values that are typically acceptable (bit patterns)
	acceptableBinary := []string{
		"0b0", "0b1", "0b10", "0b100", "0b1000", "0b10000", "0b100000", "0b1000000", "0b10000000",
		"0b11", "0b111", "0b1111", "0b11111", "0b111111", "0b1111111", "0b11111111", // Common bit masks
	}

	cleanValue := strings.ToLower(d.removeRustTypeSuffix(value))
	for _, acceptable := range acceptableBinary {
		if cleanValue == acceptable {
			return true
		}
	}

	return false
}

// removeRustTypeSuffix removes Rust type suffixes from numeric literals
func (d *RustMagicNumberDetector) removeRustTypeSuffix(value string) string {
	suffixes := []string{"u8", "u16", "u32", "u64", "u128", "i8", "i16", "i32", "i64", "i128", "usize", "isize", "f32", "f64"}
	
	for _, suffix := range suffixes {
		if strings.HasSuffix(value, suffix) {
			return strings.TrimSuffix(value, suffix)
		}
	}
	
	return value
}

// isRustPowerOfTwo checks if a number is a power of 2
func (d *RustMagicNumberDetector) isRustPowerOfTwo(value int) bool {
	if value <= 0 {
		return false
	}
	return (value & (value - 1)) == 0
}

// isRustPowerOfTen checks if a number is a power of 10
func (d *RustMagicNumberDetector) isRustPowerOfTen(value int) bool {
	if value <= 0 {
		return false
	}
	
	temp := value
	for temp > 1 {
		if temp%10 != 0 {
			return false
		}
		temp = temp / 10
	}
	return temp == 1
}

// getRustMagicNumberSeverity determines the severity of a magic number violation
func (d *RustMagicNumberDetector) getRustMagicNumberSeverity(value, valueType string) models.Severity {
	// Hex and binary literals are often more problematic than decimal
	if valueType == "hex" || valueType == "binary" {
		return models.SeverityMedium
	}

	// Large numbers are typically more problematic
	if valueType == "integer" {
		cleanValue := d.removeRustTypeSuffix(value)
		if intVal, err := strconv.Atoi(cleanValue); err == nil {
			if intVal > 1000000 {
				return models.SeverityMedium
			}
		}
	}

	return models.SeverityLow
}

// getRustMagicNumberSuggestion provides a Rust-specific suggestion for magic number violations
func (d *RustMagicNumberDetector) getRustMagicNumberSuggestion(value, valueType string) string {
	baseMsg := "Consider extracting this value to a named constant using 'const' or 'static'"

	switch valueType {
	case "hex":
		return baseMsg + ". For hex literals, use descriptive names like 'const FLAG_MASK: u32 = " + value + ";'"
	case "binary":
		return baseMsg + ". For binary literals, use descriptive names like 'const BIT_PATTERN: u8 = " + value + ";'"
	case "float":
		return baseMsg + ". For float literals, use descriptive names like 'const COEFFICIENT: f64 = " + value + ";'"
	default:
		return baseMsg + ". Use SCREAMING_SNAKE_CASE for constant names in Rust"
	}
}