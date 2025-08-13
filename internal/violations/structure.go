package violations

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/types"
)

// StructureDetector detects code structure-related violations
type StructureDetector struct {
	config *DetectorConfig
}

// NewStructureDetector creates a new structure violation detector
func NewStructureDetector(config *DetectorConfig) *StructureDetector {
	if config == nil {
		config = DefaultDetectorConfig()
	}
	return &StructureDetector{
		config: config,
	}
}

// Name returns the name of this detector
func (d *StructureDetector) Name() string {
	return "Code Structure Analysis"
}

// Description returns a description of what this detector checks for
func (d *StructureDetector) Description() string {
	return "Detects large structs, interfaces with too many methods, and other structural issues"
}

// Detect analyzes code structure and returns violations
func (d *StructureDetector) Detect(fileInfo *models.FileInfo, astInfo interface{}) []*models.Violation {
	var violations []*models.Violation

	if astInfo == nil {
		return violations
	}
	
	// Type assertion to get types.GoASTInfo
	goAstInfo, ok := astInfo.(*types.GoASTInfo)
	if !ok || goAstInfo == nil {
		// For non-Go files, we can't do structural analysis
		return violations
	}

	// Check types (structs, interfaces, etc.)
	if goAstInfo.Types != nil {
		for _, typeInfo := range goAstInfo.Types {
			if typeInfo != nil {
				violations = append(violations, d.checkTypeStructure(typeInfo, fileInfo.Path)...)
			}
		}
	}

	// Check for god objects (types with too many methods)
	violations = append(violations, d.checkForGodObjects(goAstInfo, fileInfo.Path)...)

	// Check for magic numbers
	violations = append(violations, d.checkForMagicNumbers(goAstInfo, fileInfo.Path)...)

	return violations
}

// checkTypeStructure analyzes individual types for structural issues
func (d *StructureDetector) checkTypeStructure(typeInfo *types.TypeInfo, filePath string) []*models.Violation {
	var violations []*models.Violation

	// Check struct size (line count)
	lineCount := typeInfo.EndLine - typeInfo.StartLine + 1
	if lineCount > d.config.MaxClassLines {
		violations = append(violations, &models.Violation{
			Type:        models.ViolationTypeClassSize,
			Severity:    d.getSeverityForTypeSize(lineCount),
			Message:     fmt.Sprintf("%s '%s' is too large (%d lines, max: %d)", strings.Title(typeInfo.Kind), typeInfo.Name, lineCount, d.config.MaxClassLines),
			File:        filePath,
			Line:        typeInfo.StartLine,
			Column:      typeInfo.StartColumn,
			EndLine:     typeInfo.EndLine,
			EndColumn:   typeInfo.EndColumn,
			Rule:        "type-size",
			Suggestion:  d.getTypeSizeSuggestion(typeInfo.Name, typeInfo.Kind, lineCount),
			CodeSnippet: d.generateTypeSignature(typeInfo),
		})
	}

	// Check struct field count for excessive fields
	if typeInfo.Kind == "struct" && typeInfo.FieldCount > 15 { // Default threshold for field count
		violations = append(violations, &models.Violation{
			Type:        models.ViolationTypeClassSize,
			Severity:    d.getSeverityForFieldCount(typeInfo.FieldCount),
			Message:     fmt.Sprintf("Struct '%s' has too many fields (%d), consider breaking it down", typeInfo.Name, typeInfo.FieldCount),
			File:        filePath,
			Line:        typeInfo.StartLine,
			Column:      typeInfo.StartColumn,
			EndLine:     typeInfo.EndLine,
			EndColumn:   typeInfo.EndColumn,
			Rule:        "struct-field-count",
			Suggestion:  fmt.Sprintf("Consider breaking struct '%s' into smaller, more focused structs or using composition", typeInfo.Name),
			CodeSnippet: d.generateTypeSignature(typeInfo),
		})
	}

	// Check interface method count
	if typeInfo.Kind == "interface" && typeInfo.MethodCount > 5 { // Default threshold for interface methods
		violations = append(violations, &models.Violation{
			Type:        models.ViolationTypeClassSize,
			Severity:    d.getSeverityForMethodCount(typeInfo.MethodCount),
			Message:     fmt.Sprintf("Interface '%s' has too many methods (%d), violates Interface Segregation Principle", typeInfo.Name, typeInfo.MethodCount),
			File:        filePath,
			Line:        typeInfo.StartLine,
			Column:      typeInfo.StartColumn,
			EndLine:     typeInfo.EndLine,
			EndColumn:   typeInfo.EndColumn,
			Rule:        "interface-method-count",
			Suggestion:  fmt.Sprintf("Consider splitting interface '%s' into smaller, more specific interfaces", typeInfo.Name),
			CodeSnippet: d.generateTypeSignature(typeInfo),
		})
	}

	return violations
}

// checkForGodObjects identifies types that have too many associated methods
func (d *StructureDetector) checkForGodObjects(goAstInfo *types.GoASTInfo, filePath string) []*models.Violation {
	var violations []*models.Violation

	// Group methods by receiver type
	methodsByReceiver := make(map[string][]*types.FunctionInfo)
	receiverLineInfo := make(map[string]*types.TypeInfo)

	for _, fn := range goAstInfo.Functions {
		if fn.IsMethod && fn.ReceiverType != "" {
			cleanReceiverType := d.cleanReceiverType(fn.ReceiverType)
			methodsByReceiver[cleanReceiverType] = append(methodsByReceiver[cleanReceiverType], fn)
		}
	}

	// Find corresponding type info for receivers
	for _, typeInfo := range goAstInfo.Types {
		if _, exists := methodsByReceiver[typeInfo.Name]; exists {
			receiverLineInfo[typeInfo.Name] = typeInfo
		}
	}

	// Check for god objects
	for receiverType, methods := range methodsByReceiver {
		if len(methods) > d.config.MaxMethods {
			typeInfo := receiverLineInfo[receiverType]
			line := 1
			column := 1
			if typeInfo != nil {
				line = typeInfo.StartLine
				column = typeInfo.StartColumn
			}

			violations = append(violations, &models.Violation{
				Type:        models.ViolationTypeClassSize,
				Severity:    d.getSeverityForMethodCount(len(methods)),
				Message:     fmt.Sprintf("Type '%s' has too many methods (%d, max: %d), potential God object", receiverType, len(methods), d.config.MaxMethods),
				File:        filePath,
				Line:        line,
				Column:      column,
				Rule:        "god-object",
				Suggestion:  d.getGodObjectSuggestion(receiverType, len(methods)),
				CodeSnippet: fmt.Sprintf("type %s struct { ... } // %d methods", receiverType, len(methods)),
			})
		}
	}

	return violations
}

// checkForMagicNumbers identifies hardcoded numeric literals that should be constants
func (d *StructureDetector) checkForMagicNumbers(goAstInfo *types.GoASTInfo, filePath string) []*models.Violation {
	var violations []*models.Violation

	// Walk through all functions to find magic numbers
	for _, fn := range goAstInfo.Functions {
		if fn.ASTNode != nil {
			violations = append(violations, d.findMagicNumbers(fn.ASTNode, goAstInfo.FileSet, filePath)...)
		}
	}

	return violations
}

// findMagicNumbers walks the AST to find hardcoded numbers
func (d *StructureDetector) findMagicNumbers(fn *ast.FuncDecl, fileSet *token.FileSet, filePath string) []*models.Violation {
	var violations []*models.Violation

	ast.Inspect(fn, func(n ast.Node) bool {
		if lit, ok := n.(*ast.BasicLit); ok {
			if d.isMagicNumber(lit) {
				pos := fileSet.Position(lit.Pos())
				violations = append(violations, &models.Violation{
					Type:        models.ViolationTypeMagicNumbers,
					Severity:    models.SeverityLow,
					Message:     fmt.Sprintf("Magic number '%s' should be replaced with a named constant", lit.Value),
					File:        filePath,
					Line:        pos.Line,
					Column:      pos.Column,
					Rule:        "magic-numbers",
					Suggestion:  fmt.Sprintf("Replace '%s' with a descriptive constant", lit.Value),
					CodeSnippet: lit.Value,
				})
			}
		}
		return true
	})

	return violations
}

// isMagicNumber determines if a basic literal is a magic number
func (d *StructureDetector) isMagicNumber(lit *ast.BasicLit) bool {
	// Import the token package to get the proper constants
	if lit.Kind != token.INT && lit.Kind != token.FLOAT {
		return false
	}

	// Allow common non-magic numbers
	switch lit.Value {
	case "0", "1", "2", "-1", "0.0", "1.0":
		return false
	}

	// Numbers in certain contexts are usually not magic
	// This is a simplified check - a more sophisticated version would
	// analyze the context (array indexing, common math operations, etc.)
	return true
}

// cleanReceiverType removes pointer indicators and package qualifiers
func (d *StructureDetector) cleanReceiverType(receiverType string) string {
	// Remove pointer indicator
	if strings.HasPrefix(receiverType, "*") {
		receiverType = receiverType[1:]
	}
	
	// Remove package qualifier if present
	if lastDot := strings.LastIndex(receiverType, "."); lastDot != -1 {
		receiverType = receiverType[lastDot+1:]
	}
	
	return receiverType
}

// generateTypeSignature creates a code snippet showing the type signature
func (d *StructureDetector) generateTypeSignature(typeInfo *types.TypeInfo) string {
	switch typeInfo.Kind {
	case "struct":
		return fmt.Sprintf("type %s struct { ... } // %d fields", typeInfo.Name, typeInfo.FieldCount)
	case "interface":
		return fmt.Sprintf("type %s interface { ... } // %d methods", typeInfo.Name, typeInfo.MethodCount)
	default:
		return fmt.Sprintf("type %s %s", typeInfo.Name, typeInfo.Kind)
	}
}

// Severity calculation methods

func (d *StructureDetector) getSeverityForTypeSize(lineCount int) models.Severity {
	if lineCount > d.config.MaxClassLines*2 {
		return models.SeverityHigh
	}
	if lineCount > int(float64(d.config.MaxClassLines)*1.5) {
		return models.SeverityMedium
	}
	return models.SeverityLow
}

func (d *StructureDetector) getSeverityForFieldCount(fieldCount int) models.Severity {
	if fieldCount > 25 {
		return models.SeverityHigh
	}
	if fieldCount > 20 {
		return models.SeverityMedium
	}
	return models.SeverityLow
}

func (d *StructureDetector) getSeverityForMethodCount(methodCount int) models.Severity {
	if methodCount > d.config.MaxMethods*2 {
		return models.SeverityHigh
	}
	if methodCount > int(float64(d.config.MaxMethods)*1.5) {
		return models.SeverityMedium
	}
	return models.SeverityLow
}

// Suggestion generation methods

func (d *StructureDetector) getTypeSizeSuggestion(typeName, kind string, lineCount int) string {
	return fmt.Sprintf("The %s '%s' has %d lines. Consider breaking it into smaller, more focused %ss or using composition to reduce complexity.",
		kind, typeName, lineCount, kind)
}

func (d *StructureDetector) getGodObjectSuggestion(typeName string, methodCount int) string {
	return fmt.Sprintf("Type '%s' has %d methods, suggesting it may have too many responsibilities. "+
		"Consider applying the Single Responsibility Principle by splitting it into smaller, "+
		"more focused types or extracting some functionality into separate components.", typeName, methodCount)
}