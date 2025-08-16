// Package violations provides detectors for various clean code violations in Go source code.
package violations

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/types"
)

// FunctionDetector detects function-related violations
type FunctionDetector struct {
	config        *DetectorConfig
	codeExtractor *CodeExtractor
}

// NewFunctionDetector creates a new function violation detector
func NewFunctionDetector(config *DetectorConfig) *FunctionDetector {
	if config == nil {
		config = DefaultDetectorConfig()
	}
	return &FunctionDetector{
		config:        config,
		codeExtractor: NewCodeExtractor(),
	}
}

// Name returns the name of this detector
func (d *FunctionDetector) Name() string {
	return "Function Analysis"
}

// Description returns a description of what this detector checks for
func (d *FunctionDetector) Description() string {
	return "Detects functions that are too long, complex, or have too many parameters"
}

// Detect analyzes functions and returns violations
func (d *FunctionDetector) Detect(fileInfo *models.FileInfo, astInfo interface{}) []*models.Violation {
	var violations []*models.Violation

	if astInfo == nil {
		return violations
	}
	
	// Type assertion to get types.GoASTInfo
	goAstInfo, ok := astInfo.(*types.GoASTInfo)
	if !ok || goAstInfo == nil {
		// For non-Go files, we can't do detailed function analysis
		return violations
	}

	// Check if Functions slice is nil
	if goAstInfo.Functions == nil {
		return violations
	}

	for _, function := range goAstInfo.Functions {
		if function != nil {
			violations = append(violations, d.checkFunction(function, fileInfo.Path)...)
		}
	}

	return violations
}

// checkFunction analyzes a single function for violations
func (d *FunctionDetector) checkFunction(fn *types.FunctionInfo, filePath string) []*models.Violation {
	var violations []*models.Violation

	// Check function length
	if fn.LineCount > d.config.MaxFunctionLines {
		codeSnippet := d.extractCodeSnippet(filePath, fn.StartLine, fn.EndLine)
		violations = append(violations, &models.Violation{
			Type:        models.ViolationTypeFunctionLength,
			Severity:    d.getSeverityForFunctionLength(fn.LineCount),
			Message:     fmt.Sprintf("Function '%s' is too long (%d lines, max: %d)", fn.Name, fn.LineCount, d.config.MaxFunctionLines),
			File:        filePath,
			Line:        fn.StartLine,
			Column:      fn.StartColumn,
			EndLine:     fn.EndLine,
			Rule:        "function-length",
			Suggestion:  d.getFunctionLengthSuggestion(fn.Name, fn.LineCount),
			CodeSnippet: codeSnippet,
		})
	}

	// Check cyclomatic complexity
	if fn.Complexity > d.config.MaxCyclomaticComplexity {
		codeSnippet := d.extractCodeSnippet(filePath, fn.StartLine, fn.EndLine)
		violations = append(violations, &models.Violation{
			Type:        models.ViolationTypeCyclomaticComplexity,
			Severity:    d.getSeverityForComplexity(fn.Complexity),
			Message:     fmt.Sprintf("Function '%s' has high cyclomatic complexity (%d, max: %d)", fn.Name, fn.Complexity, d.config.MaxCyclomaticComplexity),
			File:        filePath,
			Line:        fn.StartLine,
			Column:      fn.StartColumn,
			EndLine:     fn.EndLine,
			Rule:        "cyclomatic-complexity",
			Suggestion:  d.getComplexitySuggestion(fn.Name, fn.Complexity),
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
			Message:     fmt.Sprintf("Function '%s' has too many parameters (%d, max: %d)", fn.Name, paramCount, d.config.MaxParameters),
			File:        filePath,
			Line:        fn.StartLine,
			Column:      fn.StartColumn,
			Rule:        "parameter-count",
			Suggestion:  d.getParameterCountSuggestion(fn.Name, paramCount),
			CodeSnippet: codeSnippet,
		})
	}

	// Check for missing documentation on exported functions
	if d.config.RequireCommentsForPublic && fn.IsExported && !fn.HasComments {
		codeSnippet := d.extractCodeSnippet(filePath, fn.StartLine, fn.StartLine)
		violations = append(violations, &models.Violation{
			Type:        models.ViolationTypeMissingDocumentation,
			Severity:    models.SeverityMedium,
			Message:     fmt.Sprintf("Exported function '%s' is missing documentation", fn.Name),
			File:        filePath,
			Line:        fn.StartLine,
			Column:      fn.StartColumn,
			Rule:        "missing-documentation",
			Suggestion:  fmt.Sprintf("Add a comment describing what function '%s' does", fn.Name),
			CodeSnippet: codeSnippet,
		})
	}

	// Check for nesting depth (if we can analyze the AST)
	if fn.ASTNode != nil {
		nestingDepth := d.calculateNestingDepth(fn.ASTNode)
		if nestingDepth > d.config.MaxNestingDepth {
			codeSnippet := d.extractCodeSnippet(filePath, fn.StartLine, fn.EndLine)
			violations = append(violations, &models.Violation{
				Type:        models.ViolationTypeNestingDepth,
				Severity:    d.getSeverityForNestingDepth(nestingDepth),
				Message:     fmt.Sprintf("Function '%s' has excessive nesting depth (%d, max: %d)", fn.Name, nestingDepth, d.config.MaxNestingDepth),
				File:        filePath,
				Line:        fn.StartLine,
				Column:      fn.StartColumn,
				EndLine:     fn.EndLine,
				Rule:        "nesting-depth",
				Suggestion:  d.getNestingDepthSuggestion(fn.Name, nestingDepth),
				CodeSnippet: codeSnippet,
			})
		}
	}

	return violations
}

// calculateNestingDepth calculates the maximum nesting depth in a function
func (d *FunctionDetector) calculateNestingDepth(fn *ast.FuncDecl) int {
	if fn.Body == nil {
		return 0
	}

	maxDepth := 0
	
	var calculateDepth func(ast.Node, int) int
	calculateDepth = func(n ast.Node, currentDepth int) int {
		if currentDepth > maxDepth {
			maxDepth = currentDepth
		}

		switch node := n.(type) {
		case *ast.BlockStmt:
			for _, stmt := range node.List {
				calculateDepth(stmt, currentDepth)
			}
		case *ast.IfStmt:
			calculateDepth(node.Body, currentDepth+1)
			if node.Else != nil {
				calculateDepth(node.Else, currentDepth+1)
			}
		case *ast.ForStmt:
			calculateDepth(node.Body, currentDepth+1)
		case *ast.RangeStmt:
			calculateDepth(node.Body, currentDepth+1)
		case *ast.SwitchStmt:
			calculateDepth(node.Body, currentDepth+1)
		case *ast.TypeSwitchStmt:
			calculateDepth(node.Body, currentDepth+1)
		case *ast.SelectStmt:
			calculateDepth(node.Body, currentDepth+1)
		default:
			// For other node types, inspect children without increasing depth
			ast.Inspect(n, func(child ast.Node) bool {
				if child != n {
					calculateDepth(child, currentDepth)
				}
				return false // We handle recursion manually
			})
		}
		
		return maxDepth
	}

	return calculateDepth(fn.Body, 0)
}

// generateFunctionSignature creates a code snippet showing the function signature
func (d *FunctionDetector) generateFunctionSignature(fn *types.FunctionInfo) string {
	var signature strings.Builder
	
	if fn.IsMethod && fn.ReceiverType != "" {
		signature.WriteString(fmt.Sprintf("func (%s) ", fn.ReceiverType))
	} else {
		signature.WriteString("func ")
	}
	
	signature.WriteString(fn.Name)
	signature.WriteString("(")
	
	for i, param := range fn.Parameters {
		if i > 0 {
			signature.WriteString(", ")
		}
		if param.Name != "" {
			signature.WriteString(param.Name + " ")
		}
		signature.WriteString(param.Type)
	}
	
	signature.WriteString(")")
	
	if len(fn.Results) > 0 {
		if len(fn.Results) == 1 {
			signature.WriteString(" " + fn.Results[0])
		} else {
			signature.WriteString(" (")
			for i, result := range fn.Results {
				if i > 0 {
					signature.WriteString(", ")
				}
				signature.WriteString(result)
			}
			signature.WriteString(")")
		}
	}
	
	return signature.String()
}

// Severity calculation methods

func (d *FunctionDetector) getSeverityForFunctionLength(lineCount int) models.Severity {
	if lineCount > d.config.MaxFunctionLines*2 {
		return models.SeverityHigh
	}
	if lineCount > int(float64(d.config.MaxFunctionLines)*1.5) {
		return models.SeverityMedium
	}
	return models.SeverityLow
}

func (d *FunctionDetector) getSeverityForComplexity(complexity int) models.Severity {
	if complexity > d.config.MaxCyclomaticComplexity*2 {
		return models.SeverityHigh
	}
	if complexity > int(float64(d.config.MaxCyclomaticComplexity)*1.5) {
		return models.SeverityMedium
	}
	return models.SeverityLow
}

func (d *FunctionDetector) getSeverityForParameterCount(paramCount int) models.Severity {
	if paramCount > d.config.MaxParameters*2 {
		return models.SeverityHigh
	}
	if paramCount > int(float64(d.config.MaxParameters)*1.5) {
		return models.SeverityMedium
	}
	return models.SeverityLow
}

func (d *FunctionDetector) getSeverityForNestingDepth(depth int) models.Severity {
	if depth > d.config.MaxNestingDepth*2 {
		return models.SeverityHigh
	}
	if depth > int(float64(d.config.MaxNestingDepth)*1.5) {
		return models.SeverityMedium
	}
	return models.SeverityLow
}

// Suggestion generation methods

func (d *FunctionDetector) getFunctionLengthSuggestion(funcName string, lineCount int) string {
	return fmt.Sprintf("Consider breaking down function '%s' (%d lines) into smaller, more focused functions. "+
		"Each function should ideally do one thing well.", funcName, lineCount)
}

func (d *FunctionDetector) getComplexitySuggestion(funcName string, complexity int) string {
	return fmt.Sprintf("Function '%s' has cyclomatic complexity of %d. "+
		"Consider extracting complex logic into separate functions, "+
		"using early returns, or simplifying conditional logic.", funcName, complexity)
}

func (d *FunctionDetector) getParameterCountSuggestion(funcName string, paramCount int) string {
	return fmt.Sprintf("Function '%s' has %d parameters. "+
		"Consider grouping related parameters into a struct, "+
		"using options pattern, or splitting the function.", funcName, paramCount)
}

func (d *FunctionDetector) getNestingDepthSuggestion(funcName string, depth int) string {
	return fmt.Sprintf("Function '%s' has nesting depth of %d. "+
		"Consider using early returns, extracting nested logic into separate functions, "+
		"or flattening conditional statements.", funcName, depth)
}

// extractCodeSnippet extracts a code snippet for the violation with context
func (d *FunctionDetector) extractCodeSnippet(filePath string, startLine, endLine int) string {
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
func (d *FunctionDetector) generateFallbackSnippet(startLine, endLine int) string {
	if endLine <= startLine {
		return fmt.Sprintf("Line %d: <code snippet unavailable>", startLine)
	}
	return fmt.Sprintf("Lines %d-%d: <code snippet unavailable>", startLine, endLine)
}