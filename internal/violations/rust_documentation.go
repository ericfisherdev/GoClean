// Package violations provides detectors for various clean code violations in Rust source code.
package violations

import (
	"fmt"

	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/types"
)

// RustDocumentationDetector checks for missing or poor quality documentation in Rust code
type RustDocumentationDetector struct {
	config        *DetectorConfig
	codeExtractor *CodeExtractor
}

// NewRustDocumentationDetector creates a new Rust documentation detector
func NewRustDocumentationDetector(config *DetectorConfig) *RustDocumentationDetector {
	if config == nil {
		config = DefaultDetectorConfig()
	}
	return &RustDocumentationDetector{
		config:        config,
		codeExtractor: NewCodeExtractor(),
	}
}

// Name returns the name of this detector
func (d *RustDocumentationDetector) Name() string {
	return "Rust Documentation Quality"
}

// Description returns a description of what this detector checks for
func (d *RustDocumentationDetector) Description() string {
	return "Checks for missing or poor quality documentation in Rust code including structs, enums, traits, and modules"
}

// Detect analyzes the provided Rust file information and returns violations
func (d *RustDocumentationDetector) Detect(fileInfo *models.FileInfo, astInfo interface{}) []*models.Violation {
	var violations []*models.Violation

	if astInfo == nil {
		return violations
	}

	// Type assertion to get types.RustASTInfo
	rustAstInfo, ok := astInfo.(*types.RustASTInfo)
	if !ok || rustAstInfo == nil {
		return violations
	}

	// Check public functions (excluding those already checked by RustFunctionDetector)
	violations = append(violations, d.checkPublicFunctions(rustAstInfo, fileInfo.Path)...)

	// Check public structs
	violations = append(violations, d.checkPublicStructs(rustAstInfo, fileInfo.Path)...)

	// Check public enums
	violations = append(violations, d.checkPublicEnums(rustAstInfo, fileInfo.Path)...)

	// Check public traits
	violations = append(violations, d.checkPublicTraits(rustAstInfo, fileInfo.Path)...)

	// Check public modules
	violations = append(violations, d.checkPublicModules(rustAstInfo, fileInfo.Path)...)

	// Check public constants
	violations = append(violations, d.checkPublicConstants(rustAstInfo, fileInfo.Path)...)

	// Check public macros
	violations = append(violations, d.checkPublicMacros(rustAstInfo, fileInfo.Path)...)

	return violations
}

// checkPublicFunctions checks documentation for public functions
// Note: This provides additional checks beyond what RustFunctionDetector does
func (d *RustDocumentationDetector) checkPublicFunctions(rustAstInfo *types.RustASTInfo, filePath string) []*models.Violation {
	var violations []*models.Violation

	if !d.config.RequireCommentsForPublic || rustAstInfo.Functions == nil {
		return violations
	}

	for _, fn := range rustAstInfo.Functions {
		if fn == nil || !fn.IsPublic {
			continue
		}

		// Check for missing documentation
		if !fn.HasDocComments {
			codeSnippet := d.extractCodeSnippet(filePath, fn.StartLine, fn.StartLine)
			violations = append(violations, &models.Violation{
				Type:        models.ViolationTypeMissingDocumentation,
				Severity:    d.getSeverityForMissingDocumentation("function", fn.IsUnsafe),
				Message:     d.getMissingDocumentationMessage("function", fn.Name, fn.Visibility),
				File:        filePath,
				Line:        fn.StartLine,
				Column:      fn.StartColumn,
				Rule:        "rust-missing-function-documentation",
				Suggestion:  d.getFunctionDocumentationSuggestion(fn.Name, fn.IsUnsafe),
				CodeSnippet: codeSnippet,
			})
		}
	}

	return violations
}

// checkPublicStructs checks documentation for public structs
func (d *RustDocumentationDetector) checkPublicStructs(rustAstInfo *types.RustASTInfo, filePath string) []*models.Violation {
	var violations []*models.Violation

	if !d.config.RequireCommentsForPublic || rustAstInfo.Structs == nil {
		return violations
	}

	for _, structInfo := range rustAstInfo.Structs {
		if structInfo == nil || !structInfo.IsPublic {
			continue
		}

		if !structInfo.HasDocComments {
			codeSnippet := d.extractCodeSnippet(filePath, structInfo.StartLine, structInfo.StartLine)
			violations = append(violations, &models.Violation{
				Type:        models.ViolationTypeMissingDocumentation,
				Severity:    models.SeverityMedium,
				Message:     d.getMissingDocumentationMessage("struct", structInfo.Name, structInfo.Visibility),
				File:        filePath,
				Line:        structInfo.StartLine,
				Column:      structInfo.StartColumn,
				Rule:        "rust-missing-struct-documentation",
				Suggestion:  d.getStructDocumentationSuggestion(structInfo.Name),
				CodeSnippet: codeSnippet,
			})
		}
	}

	return violations
}

// checkPublicEnums checks documentation for public enums
func (d *RustDocumentationDetector) checkPublicEnums(rustAstInfo *types.RustASTInfo, filePath string) []*models.Violation {
	var violations []*models.Violation

	if !d.config.RequireCommentsForPublic || rustAstInfo.Enums == nil {
		return violations
	}

	for _, enumInfo := range rustAstInfo.Enums {
		if enumInfo == nil || !enumInfo.IsPublic {
			continue
		}

		if !enumInfo.HasDocComments {
			codeSnippet := d.extractCodeSnippet(filePath, enumInfo.StartLine, enumInfo.StartLine)
			violations = append(violations, &models.Violation{
				Type:        models.ViolationTypeMissingDocumentation,
				Severity:    models.SeverityMedium,
				Message:     d.getMissingDocumentationMessage("enum", enumInfo.Name, enumInfo.Visibility),
				File:        filePath,
				Line:        enumInfo.StartLine,
				Column:      enumInfo.StartColumn,
				Rule:        "rust-missing-enum-documentation",
				Suggestion:  d.getEnumDocumentationSuggestion(enumInfo.Name),
				CodeSnippet: codeSnippet,
			})
		}
	}

	return violations
}

// checkPublicTraits checks documentation for public traits
func (d *RustDocumentationDetector) checkPublicTraits(rustAstInfo *types.RustASTInfo, filePath string) []*models.Violation {
	var violations []*models.Violation

	if !d.config.RequireCommentsForPublic || rustAstInfo.Traits == nil {
		return violations
	}

	for _, traitInfo := range rustAstInfo.Traits {
		if traitInfo == nil || !traitInfo.IsPublic {
			continue
		}

		if !traitInfo.HasDocComments {
			codeSnippet := d.extractCodeSnippet(filePath, traitInfo.StartLine, traitInfo.StartLine)
			violations = append(violations, &models.Violation{
				Type:        models.ViolationTypeMissingDocumentation,
				Severity:    models.SeverityMedium,
				Message:     d.getMissingDocumentationMessage("trait", traitInfo.Name, traitInfo.Visibility),
				File:        filePath,
				Line:        traitInfo.StartLine,
				Column:      traitInfo.StartColumn,
				Rule:        "rust-missing-trait-documentation",
				Suggestion:  d.getTraitDocumentationSuggestion(traitInfo.Name),
				CodeSnippet: codeSnippet,
			})
		}
	}

	return violations
}

// checkPublicModules checks documentation for public modules
func (d *RustDocumentationDetector) checkPublicModules(rustAstInfo *types.RustASTInfo, filePath string) []*models.Violation {
	var violations []*models.Violation

	if !d.config.RequireCommentsForPublic || rustAstInfo.Modules == nil {
		return violations
	}

	for _, moduleInfo := range rustAstInfo.Modules {
		if moduleInfo == nil || !moduleInfo.IsPublic {
			continue
		}

		if !moduleInfo.HasDocComments {
			codeSnippet := d.extractCodeSnippet(filePath, moduleInfo.StartLine, moduleInfo.StartLine)
			violations = append(violations, &models.Violation{
				Type:        models.ViolationTypeMissingDocumentation,
				Severity:    models.SeverityLow,
				Message:     d.getMissingDocumentationMessage("module", moduleInfo.Name, moduleInfo.Visibility),
				File:        filePath,
				Line:        moduleInfo.StartLine,
				Column:      moduleInfo.StartColumn,
				Rule:        "rust-missing-module-documentation",
				Suggestion:  d.getModuleDocumentationSuggestion(moduleInfo.Name),
				CodeSnippet: codeSnippet,
			})
		}
	}

	return violations
}

// checkPublicConstants checks documentation for public constants
func (d *RustDocumentationDetector) checkPublicConstants(rustAstInfo *types.RustASTInfo, filePath string) []*models.Violation {
	var violations []*models.Violation

	if !d.config.RequireCommentsForPublic || rustAstInfo.Constants == nil {
		return violations
	}

	for _, constantInfo := range rustAstInfo.Constants {
		if constantInfo == nil || !constantInfo.IsPublic {
			continue
		}

		if !constantInfo.HasDocComments {
			codeSnippet := d.extractCodeSnippet(filePath, constantInfo.Line, constantInfo.Line)
			violations = append(violations, &models.Violation{
				Type:        models.ViolationTypeMissingDocumentation,
				Severity:    models.SeverityLow,
				Message:     d.getMissingDocumentationMessage("constant", constantInfo.Name, constantInfo.Visibility),
				File:        filePath,
				Line:        constantInfo.Line,
				Column:      constantInfo.Column,
				Rule:        "rust-missing-constant-documentation",
				Suggestion:  d.getConstantDocumentationSuggestion(constantInfo.Name),
				CodeSnippet: codeSnippet,
			})
		}
	}

	return violations
}

// checkPublicMacros checks documentation for public macros
func (d *RustDocumentationDetector) checkPublicMacros(rustAstInfo *types.RustASTInfo, filePath string) []*models.Violation {
	var violations []*models.Violation

	if !d.config.RequireCommentsForPublic || rustAstInfo.Macros == nil {
		return violations
	}

	for _, macroInfo := range rustAstInfo.Macros {
		if macroInfo == nil || !macroInfo.IsPublic {
			continue
		}

		if !macroInfo.HasDocComments {
			codeSnippet := d.extractCodeSnippet(filePath, macroInfo.StartLine, macroInfo.StartLine)
			violations = append(violations, &models.Violation{
				Type:        models.ViolationTypeMissingDocumentation,
				Severity:    models.SeverityMedium,
				Message:     d.getMissingDocumentationMessage("macro", macroInfo.Name, "pub"),
				File:        filePath,
				Line:        macroInfo.StartLine,
				Column:      macroInfo.StartColumn,
				Rule:        "rust-missing-macro-documentation",
				Suggestion:  d.getMacroDocumentationSuggestion(macroInfo.Name, macroInfo.MacroType),
				CodeSnippet: codeSnippet,
			})
		}
	}

	return violations
}

// Helper methods for generating messages and suggestions

func (d *RustDocumentationDetector) getMissingDocumentationMessage(itemType, name, visibility string) string {
	return fmt.Sprintf("Public Rust %s '%s' (%s) is missing documentation", itemType, name, visibility)
}

func (d *RustDocumentationDetector) getSeverityForMissingDocumentation(itemType string, isUnsafe bool) models.Severity {
	if isUnsafe {
		return models.SeverityHigh
	}
	switch itemType {
	case "function", "struct", "enum", "trait", "macro":
		return models.SeverityMedium
	case "module", "constant":
		return models.SeverityLow
	default:
		return models.SeverityLow
	}
}

func (d *RustDocumentationDetector) getFunctionDocumentationSuggestion(name string, isUnsafe bool) string {
	suggestion := fmt.Sprintf("Add doc comments (///) describing what function '%s' does, its parameters, and return value", name)
	if isUnsafe {
		suggestion += ". Include a # Safety section explaining the safety requirements"
	}
	return suggestion
}

func (d *RustDocumentationDetector) getStructDocumentationSuggestion(name string) string {
	return fmt.Sprintf("Add doc comments (///) describing what struct '%s' represents, its purpose, and usage examples", name)
}

func (d *RustDocumentationDetector) getEnumDocumentationSuggestion(name string) string {
	return fmt.Sprintf("Add doc comments (///) describing what enum '%s' represents and when each variant should be used", name)
}

func (d *RustDocumentationDetector) getTraitDocumentationSuggestion(name string) string {
	return fmt.Sprintf("Add doc comments (///) describing what trait '%s' provides, its contract, and implementation requirements", name)
}

func (d *RustDocumentationDetector) getModuleDocumentationSuggestion(name string) string {
	return fmt.Sprintf("Add doc comments (//!) at the top of module '%s' describing its purpose and contents", name)
}

func (d *RustDocumentationDetector) getConstantDocumentationSuggestion(name string) string {
	return fmt.Sprintf("Add doc comments (///) describing what constant '%s' represents and when it should be used", name)
}

func (d *RustDocumentationDetector) getMacroDocumentationSuggestion(name, macroType string) string {
	return fmt.Sprintf("Add doc comments (///) describing what %s macro '%s' does, its syntax, and usage examples", macroType, name)
}

// extractCodeSnippet extracts a code snippet for the violation with context
func (d *RustDocumentationDetector) extractCodeSnippet(filePath string, startLine, endLine int) string {
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
func (d *RustDocumentationDetector) generateFallbackSnippet(startLine, endLine int) string {
	if endLine <= startLine {
		return fmt.Sprintf("Line %d: <code snippet unavailable>", startLine)
	}
	return fmt.Sprintf("Lines %d-%d: <code snippet unavailable>", startLine, endLine)
}