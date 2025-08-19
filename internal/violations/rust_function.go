// Package violations provides detectors for various clean code violations in Rust source code.
package violations

import (
	"fmt"
	"strings"

	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/types"
)

// RustFunctionDetector detects function-related violations in Rust code
type RustFunctionDetector struct {
	config        *DetectorConfig
	codeExtractor *CodeExtractor
}

// NewRustFunctionDetector creates a new Rust function violation detector
func NewRustFunctionDetector(config *DetectorConfig) *RustFunctionDetector {
	if config == nil {
		config = DefaultDetectorConfig()
	}
	return &RustFunctionDetector{
		config:        config,
		codeExtractor: NewCodeExtractor(),
	}
}

// Name returns the name of this detector
func (d *RustFunctionDetector) Name() string {
	return "Rust Function Analysis"
}

// Description returns a description of what this detector checks for
func (d *RustFunctionDetector) Description() string {
	return "Detects Rust functions that are too long, complex, or have too many parameters"
}

// Detect analyzes Rust functions and returns violations
func (d *RustFunctionDetector) Detect(fileInfo *models.FileInfo, astInfo interface{}) []*models.Violation {
	var violations []*models.Violation

	if astInfo == nil {
		return violations
	}

	// Type assertion to get types.RustASTInfo
	rustAstInfo, ok := astInfo.(*types.RustASTInfo)
	if !ok || rustAstInfo == nil {
		// Not a Rust file or invalid AST info
		return violations
	}

	// Check if Functions slice is nil
	if rustAstInfo.Functions == nil {
		return violations
	}

	// Analyze each function
	for _, function := range rustAstInfo.Functions {
		if function != nil {
			violations = append(violations, d.checkRustFunction(function, fileInfo.Path)...)
		}
	}

	return violations
}

// checkRustFunction analyzes a single Rust function for violations
func (d *RustFunctionDetector) checkRustFunction(fn *types.RustFunctionInfo, filePath string) []*models.Violation {
	var violations []*models.Violation

	// Check function length
	if fn.LineCount > d.config.MaxFunctionLines {
		codeSnippet := d.extractCodeSnippet(filePath, fn.StartLine, fn.EndLine)
		violations = append(violations, &models.Violation{
			Type:        models.ViolationTypeFunctionLength,
			Severity:    d.getSeverityForFunctionLength(fn.LineCount),
			Message:     fmt.Sprintf("Rust function '%s' is too long (%d lines, max: %d)", fn.Name, fn.LineCount, d.config.MaxFunctionLines),
			File:        filePath,
			Line:        fn.StartLine,
			Column:      fn.StartColumn,
			EndLine:     fn.EndLine,
			Rule:        "rust-function-length",
			Suggestion:  d.getRustFunctionLengthSuggestion(fn.Name, fn.LineCount),
			CodeSnippet: codeSnippet,
		})
	}

	// Check cyclomatic complexity
	if fn.Complexity > d.config.MaxCyclomaticComplexity {
		codeSnippet := d.extractCodeSnippet(filePath, fn.StartLine, fn.EndLine)
		violations = append(violations, &models.Violation{
			Type:        models.ViolationTypeCyclomaticComplexity,
			Severity:    d.getSeverityForComplexity(fn.Complexity),
			Message:     fmt.Sprintf("Rust function '%s' has high cyclomatic complexity (%d, max: %d)", fn.Name, fn.Complexity, d.config.MaxCyclomaticComplexity),
			File:        filePath,
			Line:        fn.StartLine,
			Column:      fn.StartColumn,
			EndLine:     fn.EndLine,
			Rule:        "rust-cyclomatic-complexity",
			Suggestion:  d.getRustComplexitySuggestion(fn.Name, fn.Complexity),
			CodeSnippet: codeSnippet,
		})
	}

	// Check parameter count
	paramCount := len(fn.Parameters)
	if paramCount > d.config.MaxParameters {
		codeSnippet := d.extractCodeSnippet(filePath, fn.StartLine, fn.StartLine)
		violations = append(violations, &models.Violation{
			Type:        models.ViolationTypeParameterCount,
			Severity:    d.getSeverityForParameterCount(paramCount),
			Message:     fmt.Sprintf("Rust function '%s' has too many parameters (%d, max: %d)", fn.Name, paramCount, d.config.MaxParameters),
			File:        filePath,
			Line:        fn.StartLine,
			Column:      fn.StartColumn,
			Rule:        "rust-parameter-count",
			Suggestion:  d.getRustParameterCountSuggestion(fn.Name, paramCount),
			CodeSnippet: codeSnippet,
		})
	}

	// Check for missing documentation on public functions
	if d.config.RequireCommentsForPublic && fn.IsPublic && !fn.HasDocComments {
		codeSnippet := d.extractCodeSnippet(filePath, fn.StartLine, fn.StartLine)
		violations = append(violations, &models.Violation{
			Type:        models.ViolationTypeMissingDocumentation,
			Severity:    models.SeverityMedium,
			Message:     fmt.Sprintf("Public Rust function '%s' is missing documentation", fn.Name),
			File:        filePath,
			Line:        fn.StartLine,
			Column:      fn.StartColumn,
			Rule:        "rust-missing-documentation",
			Suggestion:  fmt.Sprintf("Add doc comments (///) describing what function '%s' does", fn.Name),
			CodeSnippet: codeSnippet,
		})
	}

	// Check for unsafe functions without documentation
	if fn.IsUnsafe && !fn.HasDocComments {
		codeSnippet := d.extractCodeSnippet(filePath, fn.StartLine, fn.StartLine)
		violations = append(violations, &models.Violation{
			Type:        models.ViolationTypeMissingDocumentation,
			Severity:    models.SeverityHigh,
			Message:     fmt.Sprintf("Unsafe Rust function '%s' must have documentation explaining safety requirements", fn.Name),
			File:        filePath,
			Line:        fn.StartLine,
			Column:      fn.StartColumn,
			Rule:        "rust-unsafe-missing-documentation",
			Suggestion:  fmt.Sprintf("Add # Safety section in doc comments explaining why this unsafe function is sound"),
			CodeSnippet: codeSnippet,
		})
	}

	// Check for overly long async functions (they tend to be more complex)
	if fn.IsAsync && fn.LineCount > d.config.MaxFunctionLines/2 {
		codeSnippet := d.extractCodeSnippet(filePath, fn.StartLine, fn.EndLine)
		violations = append(violations, &models.Violation{
			Type:        models.ViolationTypeFunctionLength,
			Severity:    models.SeverityMedium,
			Message:     fmt.Sprintf("Async Rust function '%s' is complex (%d lines). Consider breaking it down", fn.Name, fn.LineCount),
			File:        filePath,
			Line:        fn.StartLine,
			Column:      fn.StartColumn,
			EndLine:     fn.EndLine,
			Rule:        "rust-async-complexity",
			Suggestion:  d.getRustAsyncComplexitySuggestion(fn.Name, fn.LineCount),
			CodeSnippet: codeSnippet,
		})
	}

	return violations
}

// generateRustFunctionSignature creates a code snippet showing the Rust function signature
func (d *RustFunctionDetector) generateRustFunctionSignature(fn *types.RustFunctionInfo) string {
	var signature strings.Builder

	// Add visibility modifier
	if fn.Visibility != "" && fn.Visibility != "private" {
		signature.WriteString(fn.Visibility + " ")
	}

	// Add function modifiers
	if fn.IsAsync {
		signature.WriteString("async ")
	}
	if fn.IsUnsafe {
		signature.WriteString("unsafe ")
	}
	if fn.IsConst {
		signature.WriteString("const ")
	}

	signature.WriteString("fn ")
	signature.WriteString(fn.Name)
	signature.WriteString("(")

	// Add parameters
	for i, param := range fn.Parameters {
		if i > 0 {
			signature.WriteString(", ")
		}
		if param.IsMutable {
			signature.WriteString("mut ")
		}
		if param.Name != "" {
			signature.WriteString(param.Name)
			signature.WriteString(": ")
		}
		if param.IsRef {
			signature.WriteString("&")
		}
		signature.WriteString(param.Type)
	}

	signature.WriteString(")")

	// Add return type
	if fn.ReturnType != "" && fn.ReturnType != "()" {
		signature.WriteString(" -> ")
		signature.WriteString(fn.ReturnType)
	}

	return signature.String()
}

// Severity calculation methods (reusing logic but can be customized for Rust)

func (d *RustFunctionDetector) getSeverityForFunctionLength(lineCount int) models.Severity {
	if lineCount > d.config.MaxFunctionLines*2 {
		return models.SeverityHigh
	}
	if lineCount > int(float64(d.config.MaxFunctionLines)*1.5) {
		return models.SeverityMedium
	}
	return models.SeverityLow
}

func (d *RustFunctionDetector) getSeverityForComplexity(complexity int) models.Severity {
	if complexity > d.config.MaxCyclomaticComplexity*2 {
		return models.SeverityHigh
	}
	if complexity > int(float64(d.config.MaxCyclomaticComplexity)*1.5) {
		return models.SeverityMedium
	}
	return models.SeverityLow
}

func (d *RustFunctionDetector) getSeverityForParameterCount(paramCount int) models.Severity {
	if paramCount > d.config.MaxParameters*2 {
		return models.SeverityHigh
	}
	if paramCount > int(float64(d.config.MaxParameters)*1.5) {
		return models.SeverityMedium
	}
	return models.SeverityLow
}

// Rust-specific suggestion generation methods

func (d *RustFunctionDetector) getRustFunctionLengthSuggestion(funcName string, lineCount int) string {
	return fmt.Sprintf("Consider breaking down function '%s' (%d lines) into smaller, more focused functions. "+
		"Use Rust's powerful type system and traits to create composable abstractions.", funcName, lineCount)
}

func (d *RustFunctionDetector) getRustComplexitySuggestion(funcName string, complexity int) string {
	return fmt.Sprintf("Function '%s' has cyclomatic complexity of %d. "+
		"Consider using pattern matching, early returns with the ? operator, "+
		"or extracting complex logic into separate functions or methods.", funcName, complexity)
}

func (d *RustFunctionDetector) getRustParameterCountSuggestion(funcName string, paramCount int) string {
	return fmt.Sprintf("Function '%s' has %d parameters. "+
		"Consider using a builder pattern, grouping related parameters into structs, "+
		"or using Rust's Default trait for optional parameters.", funcName, paramCount)
}

func (d *RustFunctionDetector) getRustAsyncComplexitySuggestion(funcName string, lineCount int) string {
	return fmt.Sprintf("Async function '%s' is %d lines long. "+
		"Consider breaking it into smaller async functions, using async blocks for organization, "+
		"or extracting non-async logic into separate synchronous functions.", funcName, lineCount)
}

// extractCodeSnippet extracts a code snippet for the violation with context
func (d *RustFunctionDetector) extractCodeSnippet(filePath string, startLine, endLine int) string {
	if d.codeExtractor == nil {
		return d.generateFallbackSnippet(startLine, endLine)
	}

	snippet, err := d.codeExtractor.ExtractSnippet(filePath, startLine, endLine)
	if err != nil {
		return d.generateFallbackSnippet(startLine, endLine)
	}

	return snippet
}

// generateFallbackSnippet creates a simple snippet when file reading fails
func (d *RustFunctionDetector) generateFallbackSnippet(startLine, endLine int) string {
	if endLine <= startLine {
		return fmt.Sprintf("Line %d: <code snippet unavailable>", startLine)
	}
	return fmt.Sprintf("Lines %d-%d: <code snippet unavailable>", startLine, endLine)
}