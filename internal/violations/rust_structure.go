// Package violations provides detectors for various clean code violations in Rust source code.
package violations

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/types"
)

// Rust-specific structure analysis thresholds
const (
	RustMaxStructFields     = 12  // Maximum number of fields in a struct
	RustMaxTraitMethods     = 8   // Maximum number of methods in a trait
	RustMaxImplMethods      = 20  // Maximum number of methods in an impl block
	RustMaxEnumVariants     = 15  // Maximum number of variants in an enum
	RustMaxModuleItems      = 50  // Maximum number of items in a module
	RustMaxFileLines        = 1000 // Maximum lines in a single file
	
	// Severity calculation thresholds
	RustHighComplexityThreshold   = 2.0  // Multiplier for high severity
	RustMediumComplexityThreshold = 1.5  // Multiplier for medium severity
)

// RustStructureDetector detects structural issues in Rust code
type RustStructureDetector struct {
	config        *DetectorConfig
	codeExtractor *CodeExtractor
	// Compiled regex patterns for structural analysis
	usePattern      *regexp.Regexp
	modPattern      *regexp.Regexp
	implPattern     *regexp.Regexp
	structPattern   *regexp.Regexp
	enumPattern     *regexp.Regexp
	traitPattern    *regexp.Regexp
}

// NewRustStructureDetector creates a new Rust structure detector
func NewRustStructureDetector(config *DetectorConfig) *RustStructureDetector {
	if config == nil {
		config = DefaultDetectorConfig()
	}

	return &RustStructureDetector{
		config:        config,
		codeExtractor: NewCodeExtractor(),
		// Rust-specific patterns for structural analysis
		usePattern:    regexp.MustCompile(`^\s*use\s+`),
		modPattern:    regexp.MustCompile(`^\s*(pub\s+)?mod\s+\w+`),
		implPattern:   regexp.MustCompile(`^\s*impl\s*(<[^>]*>)?\s*\w+`),
		structPattern: regexp.MustCompile(`^\s*(pub\s+)?struct\s+\w+`),
		enumPattern:   regexp.MustCompile(`^\s*(pub\s+)?enum\s+\w+`),
		traitPattern:  regexp.MustCompile(`^\s*(pub\s+)?trait\s+\w+`),
	}
}

// Name returns the name of this detector
func (d *RustStructureDetector) Name() string {
	return "Rust Code Structure Analysis"
}

// Description returns a description of what this detector checks for
func (d *RustStructureDetector) Description() string {
	return "Detects structural issues in Rust code including large modules, complex traits, and file organization problems"
}

// Detect analyzes the provided Rust file information and returns violations
func (d *RustStructureDetector) Detect(fileInfo *models.FileInfo, astInfo interface{}) []*models.Violation {
	var violations []*models.Violation

	if astInfo == nil {
		return violations
	}

	// Type assertion to get types.RustASTInfo
	rustAstInfo, ok := astInfo.(*types.RustASTInfo)
	if !ok || rustAstInfo == nil {
		return violations
	}

	// Read the file content for comprehensive analysis
	content, err := d.readFileContent(fileInfo.Path)
	if err != nil {
		return violations
	}

	// Check file-level structure
	violations = append(violations, d.checkFileStructure(fileInfo.Path, content)...)

	// Check struct complexity
	violations = append(violations, d.checkStructComplexity(rustAstInfo, fileInfo.Path)...)

	// Check enum complexity
	violations = append(violations, d.checkEnumComplexity(rustAstInfo, fileInfo.Path)...)

	// Check trait complexity
	violations = append(violations, d.checkTraitComplexity(rustAstInfo, fileInfo.Path)...)

	// Check implementation block complexity
	violations = append(violations, d.checkImplComplexity(rustAstInfo, fileInfo.Path)...)

	// Check module organization
	violations = append(violations, d.checkModuleOrganization(rustAstInfo, fileInfo.Path, content)...)

	return violations
}

// readFileContent reads the content of a file
func (d *RustStructureDetector) readFileContent(filePath string) (string, error) {
	if d.codeExtractor == nil {
		return "", fmt.Errorf("code extractor not available")
	}

	content, err := d.codeExtractor.ExtractSnippet(filePath, 1, -1) // Read entire file
	if err != nil {
		return "", err
	}

	return content, nil
}

// checkFileStructure checks overall file structure and organization
func (d *RustStructureDetector) checkFileStructure(filePath, content string) []*models.Violation {
	var violations []*models.Violation

	lines := strings.Split(content, "\n")
	lineCount := len(lines)

	// Check if file is too large
	if lineCount > RustMaxFileLines {
		violations = append(violations, &models.Violation{
			Type:        models.ViolationTypeClassSize,
			Severity:    d.getRustFileSizeSeverity(lineCount),
			Message:     fmt.Sprintf("Rust file is too large (%d lines, max: %d)", lineCount, RustMaxFileLines),
			File:        filePath,
			Line:        1,
			Column:      1,
			Rule:        "rust-file-size",
			Suggestion:  "Consider splitting this large file into smaller, more focused modules",
			CodeSnippet: d.generateFileSummary(lines),
		})
	}

	// Check for excessive use statements (might indicate unclear module organization)
	useCount := d.countPatternMatches(lines, d.usePattern)
	if useCount > 25 {
		violations = append(violations, &models.Violation{
			Type:        models.ViolationTypeStructure,
			Severity:    models.SeverityMedium,
			Message:     fmt.Sprintf("Excessive use statements (%d) may indicate unclear module organization", useCount),
			File:        filePath,
			Line:        1,
			Column:      1,
			Rule:        "rust-excessive-imports",
			Suggestion:  "Consider reorganizing modules or using re-exports to reduce dependency complexity",
			CodeSnippet: fmt.Sprintf("File contains %d use statements", useCount),
		})
	}

	return violations
}

// checkStructComplexity analyzes struct complexity
func (d *RustStructureDetector) checkStructComplexity(rustAstInfo *types.RustASTInfo, filePath string) []*models.Violation {
	var violations []*models.Violation

	if rustAstInfo.Structs == nil {
		return violations
	}

	for _, structInfo := range rustAstInfo.Structs {
		if structInfo == nil {
			continue
		}

		// Check field count
		if structInfo.FieldCount > RustMaxStructFields {
			violations = append(violations, &models.Violation{
				Type:        models.ViolationTypeClassSize,
				Severity:    d.getRustStructComplexitySeverity(structInfo.FieldCount),
				Message:     fmt.Sprintf("Struct '%s' has too many fields (%d, max: %d)", structInfo.Name, structInfo.FieldCount, RustMaxStructFields),
				File:        filePath,
				Line:        structInfo.StartLine,
				Column:      structInfo.StartColumn,
				Rule:        "rust-struct-complexity",
				Suggestion:  "Consider breaking down large struct into smaller, more focused structs or using composition",
				CodeSnippet: fmt.Sprintf("struct %s { /* %d fields */ }", structInfo.Name, structInfo.FieldCount),
			})
		}

		// Check if struct is too long (lines)
		structLines := structInfo.EndLine - structInfo.StartLine + 1
		if structLines > d.config.MaxClassLines {
			violations = append(violations, &models.Violation{
				Type:        models.ViolationTypeClassSize,
				Severity:    d.getRustLineCountSeverity(structLines),
				Message:     fmt.Sprintf("Struct '%s' definition is too long (%d lines, max: %d)", structInfo.Name, structLines, d.config.MaxClassLines),
				File:        filePath,
				Line:        structInfo.StartLine,
				Column:      structInfo.StartColumn,
				Rule:        "rust-struct-length",
				Suggestion:  "Consider simplifying the struct definition or using type aliases for complex field types",
				CodeSnippet: fmt.Sprintf("struct %s spans %d lines", structInfo.Name, structLines),
			})
		}
	}

	return violations
}

// checkEnumComplexity analyzes enum complexity
func (d *RustStructureDetector) checkEnumComplexity(rustAstInfo *types.RustASTInfo, filePath string) []*models.Violation {
	var violations []*models.Violation

	if rustAstInfo.Enums == nil {
		return violations
	}

	for _, enumInfo := range rustAstInfo.Enums {
		if enumInfo == nil {
			continue
		}

		// Check variant count
		if enumInfo.VariantCount > RustMaxEnumVariants {
			violations = append(violations, &models.Violation{
				Type:        models.ViolationTypeClassSize,
				Severity:    d.getRustEnumComplexitySeverity(enumInfo.VariantCount),
				Message:     fmt.Sprintf("Enum '%s' has too many variants (%d, max: %d)", enumInfo.Name, enumInfo.VariantCount, RustMaxEnumVariants),
				File:        filePath,
				Line:        enumInfo.StartLine,
				Column:      enumInfo.StartColumn,
				Rule:        "rust-enum-complexity",
				Suggestion:  "Consider splitting large enum into smaller enums or using trait objects for behavior-based variants",
				CodeSnippet: fmt.Sprintf("enum %s { /* %d variants */ }", enumInfo.Name, enumInfo.VariantCount),
			})
		}

		// Check if enum definition is too long
		enumLines := enumInfo.EndLine - enumInfo.StartLine + 1
		if enumLines > d.config.MaxClassLines/2 { // Enums should be more compact
			violations = append(violations, &models.Violation{
				Type:        models.ViolationTypeClassSize,
				Severity:    models.SeverityMedium,
				Message:     fmt.Sprintf("Enum '%s' definition is too long (%d lines)", enumInfo.Name, enumLines),
				File:        filePath,
				Line:        enumInfo.StartLine,
				Column:      enumInfo.StartColumn,
				Rule:        "rust-enum-length",
				Suggestion:  "Consider simplifying enum variants or extracting complex variant data into separate structs",
				CodeSnippet: fmt.Sprintf("enum %s spans %d lines", enumInfo.Name, enumLines),
			})
		}
	}

	return violations
}

// checkTraitComplexity analyzes trait complexity
func (d *RustStructureDetector) checkTraitComplexity(rustAstInfo *types.RustASTInfo, filePath string) []*models.Violation {
	var violations []*models.Violation

	if rustAstInfo.Traits == nil {
		return violations
	}

	for _, traitInfo := range rustAstInfo.Traits {
		if traitInfo == nil {
			continue
		}

		// Check method count
		if traitInfo.MethodCount > RustMaxTraitMethods {
			violations = append(violations, &models.Violation{
				Type:        models.ViolationTypeClassSize,
				Severity:    d.getRustTraitComplexitySeverity(traitInfo.MethodCount),
				Message:     fmt.Sprintf("Trait '%s' has too many methods (%d, max: %d)", traitInfo.Name, traitInfo.MethodCount, RustMaxTraitMethods),
				File:        filePath,
				Line:        traitInfo.StartLine,
				Column:      traitInfo.StartColumn,
				Rule:        "rust-trait-complexity",
				Suggestion:  "Consider splitting large trait into smaller, more focused traits using trait composition",
				CodeSnippet: fmt.Sprintf("trait %s { /* %d methods */ }", traitInfo.Name, traitInfo.MethodCount),
			})
		}

		// Check if trait definition is too long
		traitLines := traitInfo.EndLine - traitInfo.StartLine + 1
		if traitLines > d.config.MaxClassLines {
			violations = append(violations, &models.Violation{
				Type:        models.ViolationTypeClassSize,
				Severity:    models.SeverityMedium,
				Message:     fmt.Sprintf("Trait '%s' definition is too long (%d lines)", traitInfo.Name, traitLines),
				File:        filePath,
				Line:        traitInfo.StartLine,
				Column:      traitInfo.StartColumn,
				Rule:        "rust-trait-length",
				Suggestion:  "Consider simplifying trait methods or providing default implementations",
				CodeSnippet: fmt.Sprintf("trait %s spans %d lines", traitInfo.Name, traitLines),
			})
		}
	}

	return violations
}

// checkImplComplexity analyzes implementation block complexity
func (d *RustStructureDetector) checkImplComplexity(rustAstInfo *types.RustASTInfo, filePath string) []*models.Violation {
	var violations []*models.Violation

	if rustAstInfo.Impls == nil {
		return violations
	}

	for _, implInfo := range rustAstInfo.Impls {
		if implInfo == nil {
			continue
		}

		// Check method count
		if implInfo.MethodCount > RustMaxImplMethods {
			implType := "impl"
			if implInfo.TraitName != "" {
				implType = fmt.Sprintf("impl %s for", implInfo.TraitName)
			}

			violations = append(violations, &models.Violation{
				Type:        models.ViolationTypeClassSize,
				Severity:    d.getRustImplComplexitySeverity(implInfo.MethodCount),
				Message:     fmt.Sprintf("%s %s has too many methods (%d, max: %d)", implType, implInfo.TargetType, implInfo.MethodCount, RustMaxImplMethods),
				File:        filePath,
				Line:        implInfo.StartLine,
				Column:      implInfo.StartColumn,
				Rule:        "rust-impl-complexity",
				Suggestion:  "Consider splitting large implementation into multiple impl blocks or extracting functionality into separate modules",
				CodeSnippet: fmt.Sprintf("%s %s { /* %d methods */ }", implType, implInfo.TargetType, implInfo.MethodCount),
			})
		}

		// Check if impl block is too long
		implLines := implInfo.EndLine - implInfo.StartLine + 1
		if implLines > d.config.MaxClassLines {
			violations = append(violations, &models.Violation{
				Type:        models.ViolationTypeClassSize,
				Severity:    models.SeverityMedium,
				Message:     fmt.Sprintf("Implementation for '%s' is too long (%d lines)", implInfo.TargetType, implLines),
				File:        filePath,
				Line:        implInfo.StartLine,
				Column:      implInfo.StartColumn,
				Rule:        "rust-impl-length",
				Suggestion:  "Consider breaking down large implementation into smaller, more focused impl blocks",
				CodeSnippet: fmt.Sprintf("impl %s spans %d lines", implInfo.TargetType, implLines),
			})
		}
	}

	return violations
}

// checkModuleOrganization analyzes module organization and structure
func (d *RustStructureDetector) checkModuleOrganization(rustAstInfo *types.RustASTInfo, filePath, content string) []*models.Violation {
	var violations []*models.Violation

	lines := strings.Split(content, "\n")

	// Count different types of items in the file
	itemCounts := d.countFileItems(lines)
	totalItems := itemCounts["struct"] + itemCounts["enum"] + itemCounts["trait"] + itemCounts["impl"] + itemCounts["fn"]

	// Check if file has too many items (indicates it might need to be split)
	if totalItems > RustMaxModuleItems {
		violations = append(violations, &models.Violation{
			Type:        models.ViolationTypeStructure,
			Severity:    models.SeverityMedium,
			Message:     fmt.Sprintf("File contains too many items (%d, max: %d)", totalItems, RustMaxModuleItems),
			File:        filePath,
			Line:        1,
			Column:      1,
			Rule:        "rust-module-organization",
			Suggestion:  "Consider splitting this file into multiple modules for better organization",
			CodeSnippet: d.generateItemSummary(itemCounts),
		})
	}

	// Check for inconsistent organization (e.g., functions mixed with struct definitions)
	organizationIssues := d.checkOrganizationPatterns(lines)
	for _, issue := range organizationIssues {
		violations = append(violations, &models.Violation{
			Type:        models.ViolationTypeStructure,
			Severity:    models.SeverityLow,
			Message:     issue.message,
			File:        filePath,
			Line:        issue.line,
			Column:      1,
			Rule:        "rust-organization-pattern",
			Suggestion:  issue.suggestion,
			CodeSnippet: issue.snippet,
		})
	}

	return violations
}

// OrganizationIssue represents an issue with code organization
type OrganizationIssue struct {
	message    string
	line       int
	suggestion string
	snippet    string
}

// checkOrganizationPatterns checks for inconsistent organization patterns
func (d *RustStructureDetector) checkOrganizationPatterns(lines []string) []OrganizationIssue {
	var issues []OrganizationIssue

	// Track the order of different item types
	itemOrder := []string{}
	itemLines := []int{}

	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)

		if d.structPattern.MatchString(trimmed) {
			itemOrder = append(itemOrder, "struct")
			itemLines = append(itemLines, lineNum)
		} else if d.enumPattern.MatchString(trimmed) {
			itemOrder = append(itemOrder, "enum")
			itemLines = append(itemLines, lineNum)
		} else if d.traitPattern.MatchString(trimmed) {
			itemOrder = append(itemOrder, "trait")
			itemLines = append(itemLines, lineNum)
		} else if d.implPattern.MatchString(trimmed) {
			itemOrder = append(itemOrder, "impl")
			itemLines = append(itemLines, lineNum)
		} else if strings.HasPrefix(trimmed, "fn ") || strings.HasPrefix(trimmed, "pub fn ") {
			itemOrder = append(itemOrder, "function")
			itemLines = append(itemLines, lineNum)
		}
	}

	// Check for mixed organization (functions scattered among type definitions)
	if len(itemOrder) > 5 {
		functionIndices := []int{}
		for i, item := range itemOrder {
			if item == "function" {
				functionIndices = append(functionIndices, i)
			}
		}

		// If functions are scattered (not grouped together), suggest reorganization
		if len(functionIndices) > 2 && d.areIndicesScattered(functionIndices) {
			firstFunction := functionIndices[0]
			issues = append(issues, OrganizationIssue{
				message:    "Functions are scattered throughout the file instead of being grouped together",
				line:       itemLines[firstFunction],
				suggestion: "Consider grouping related functions together or organizing by functionality",
				snippet:    "Functions mixed with type definitions",
			})
		}
	}

	return issues
}

// areIndicesScattered checks if indices are scattered (not consecutive)
func (d *RustStructureDetector) areIndicesScattered(indices []int) bool {
	if len(indices) < 3 {
		return false
	}

	gaps := 0
	for i := 1; i < len(indices); i++ {
		if indices[i]-indices[i-1] > 2 { // Gap larger than 2
			gaps++
		}
	}

	return gaps > 1 // Multiple large gaps indicate scattering
}

// countFileItems counts different types of items in the file
func (d *RustStructureDetector) countFileItems(lines []string) map[string]int {
	counts := map[string]int{
		"struct": 0,
		"enum":   0,
		"trait":  0,
		"impl":   0,
		"fn":     0,
		"mod":    0,
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if d.structPattern.MatchString(trimmed) {
			counts["struct"]++
		} else if d.enumPattern.MatchString(trimmed) {
			counts["enum"]++
		} else if d.traitPattern.MatchString(trimmed) {
			counts["trait"]++
		} else if d.implPattern.MatchString(trimmed) {
			counts["impl"]++
		} else if strings.HasPrefix(trimmed, "fn ") || strings.HasPrefix(trimmed, "pub fn ") {
			counts["fn"]++
		} else if d.modPattern.MatchString(trimmed) {
			counts["mod"]++
		}
	}

	return counts
}

// countPatternMatches counts how many lines match a given pattern
func (d *RustStructureDetector) countPatternMatches(lines []string, pattern *regexp.Regexp) int {
	count := 0
	for _, line := range lines {
		if pattern.MatchString(line) {
			count++
		}
	}
	return count
}

// generateFileSummary generates a summary of file contents
func (d *RustStructureDetector) generateFileSummary(lines []string) string {
	lineCount := len(lines)
	if lineCount <= 5 {
		return strings.Join(lines, "\n")
	}

	// Show first 3 and last 2 lines
	summary := strings.Join(lines[:3], "\n") + "\n...\n" + strings.Join(lines[lineCount-2:], "\n")
	return summary
}

// generateItemSummary generates a summary of items in the file
func (d *RustStructureDetector) generateItemSummary(counts map[string]int) string {
	var parts []string
	for itemType, count := range counts {
		if count > 0 {
			parts = append(parts, fmt.Sprintf("%d %s(s)", count, itemType))
		}
	}
	return strings.Join(parts, ", ")
}

// Severity calculation methods

func (d *RustStructureDetector) getRustFileSizeSeverity(lineCount int) models.Severity {
	highThreshold := int(float64(RustMaxFileLines) * RustHighComplexityThreshold)
	mediumThreshold := int(float64(RustMaxFileLines) * RustMediumComplexityThreshold)
	
	if lineCount > highThreshold {
		return models.SeverityHigh
	}
	if lineCount > mediumThreshold {
		return models.SeverityMedium
	}
	return models.SeverityLow
}

func (d *RustStructureDetector) getRustStructComplexitySeverity(fieldCount int) models.Severity {
	highThreshold := int(float64(RustMaxStructFields) * RustHighComplexityThreshold)
	mediumThreshold := int(float64(RustMaxStructFields) * RustMediumComplexityThreshold)
	
	if fieldCount > highThreshold {
		return models.SeverityHigh
	}
	if fieldCount > mediumThreshold {
		return models.SeverityMedium
	}
	return models.SeverityLow
}

func (d *RustStructureDetector) getRustEnumComplexitySeverity(variantCount int) models.Severity {
	if variantCount > 30 { // RustMaxEnumVariants (15) * 2.0
		return models.SeverityHigh
	}
	if variantCount > 22 { // RustMaxEnumVariants (15) * 1.5 (rounded)
		return models.SeverityMedium
	}
	return models.SeverityLow
}

func (d *RustStructureDetector) getRustTraitComplexitySeverity(methodCount int) models.Severity {
	highThreshold := int(float64(RustMaxTraitMethods) * RustHighComplexityThreshold)
	mediumThreshold := int(float64(RustMaxTraitMethods) * RustMediumComplexityThreshold)
	
	if methodCount > highThreshold {
		return models.SeverityHigh
	}
	if methodCount > mediumThreshold {
		return models.SeverityMedium
	}
	return models.SeverityLow
}

func (d *RustStructureDetector) getRustImplComplexitySeverity(methodCount int) models.Severity {
	highThreshold := int(float64(RustMaxImplMethods) * RustHighComplexityThreshold)
	mediumThreshold := int(float64(RustMaxImplMethods) * RustMediumComplexityThreshold)
	
	if methodCount > highThreshold {
		return models.SeverityHigh
	}
	if methodCount > mediumThreshold {
		return models.SeverityMedium
	}
	return models.SeverityLow
}

func (d *RustStructureDetector) getRustLineCountSeverity(lineCount int) models.Severity {
	threshold := d.config.MaxClassLines
	highThreshold := int(float64(threshold) * RustHighComplexityThreshold)
	mediumThreshold := int(float64(threshold) * RustMediumComplexityThreshold)
	
	if lineCount > highThreshold {
		return models.SeverityHigh
	}
	if lineCount > mediumThreshold {
		return models.SeverityMedium
	}
	return models.SeverityLow
}