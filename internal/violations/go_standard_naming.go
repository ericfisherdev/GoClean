// Package violations provides detectors for various clean code violations in Go source code.
package violations

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/types"
)

const (
	// Violation types for Go standard naming
	goStandardNamingRule = "go-standard-naming"
)

var commonInitialisms = map[string]struct{}{
	"API": {}, "ASCII": {}, "CPU": {}, "CSS": {}, "DNS": {}, "EOF": {}, "GUID": {},
	"HTML": {}, "HTTP": {}, "HTTPS": {}, "ID": {}, "IP": {}, "JSON": {}, "LHS": {},
	"QPS": {}, "RAM": {}, "RHS": {}, "RPC": {}, "SLA": {}, "SMTP": {}, "SQL": {},
	"SSH": {}, "TCP": {}, "TLS": {}, "TTL": {}, "UDP": {}, "UI": {}, "UID": {},
	"UUID": {}, "URI": {}, "URL": {}, "UTF8": {}, "VM": {}, "XML": {}, "XMPP": {},
	"XSRF": {}, "XSS": {},
}

// GoStandardNamingDetector enforces Go standard naming conventions
type GoStandardNamingDetector struct {
	config        *DetectorConfig
	codeExtractor *CodeExtractor
}

// NewGoStandardNamingDetector creates a new Go standard naming detector
func NewGoStandardNamingDetector(config *DetectorConfig) *GoStandardNamingDetector {
	if config == nil {
		config = DefaultDetectorConfig()
	}
	return &GoStandardNamingDetector{
		config:        config,
		codeExtractor: NewCodeExtractor(),
	}
}

// Name returns the name of this detector
func (d *GoStandardNamingDetector) Name() string {
	return "Go Standard Naming Conventions"
}

// Description returns a description of what this detector checks for
func (d *GoStandardNamingDetector) Description() string {
	return "Enforces Go standard naming conventions including MixedCaps, no underscores, proper exported/unexported casing, and conventional patterns"
}

// Detect analyzes Go files for standard naming convention violations
func (d *GoStandardNamingDetector) Detect(fileInfo *models.FileInfo, astInfo interface{}) []*models.Violation {
	var violations []*models.Violation

	// Only analyze Go files
	if !strings.HasSuffix(fileInfo.Path, ".go") {
		return violations
	}

	// Optionally skip test files
	if d.config != nil && d.config.SkipTestFiles && strings.HasSuffix(fileInfo.Path, "_test.go") {
		return violations
	}

	if astInfo == nil {
		return violations
	}

	// Type assertion to get Go AST info
	goAstInfo, ok := astInfo.(*types.GoASTInfo)
	if !ok || goAstInfo == nil || goAstInfo.AST == nil {
		return violations
	}

	// Check package naming
	violations = append(violations, d.checkPackageNaming(goAstInfo, fileInfo.Path)...)

	// Check function names
	if goAstInfo.Functions != nil {
		for _, function := range goAstInfo.Functions {
			if function != nil {
				violations = append(violations, d.checkFunctionNaming(function, fileInfo.Path)...)
			}
		}
	}

	// Check type names
	if goAstInfo.Types != nil {
		for _, typeInfo := range goAstInfo.Types {
			if typeInfo != nil {
				violations = append(violations, d.checkTypeNaming(typeInfo, fileInfo.Path)...)
			}
		}
	}

	// Check variable names
	if goAstInfo.Variables != nil {
		for _, variable := range goAstInfo.Variables {
			if variable != nil {
				violations = append(violations, d.checkVariableNaming(variable, fileInfo.Path)...)
			}
		}
	}

	// Check constant names
	if goAstInfo.Constants != nil {
		for _, constant := range goAstInfo.Constants {
			if constant != nil {
				violations = append(violations, d.checkConstantNaming(constant, fileInfo.Path)...)
			}
		}
	}

	return violations
}

// checkPackageNaming validates package naming conventions
func (d *GoStandardNamingDetector) checkPackageNaming(goAstInfo *types.GoASTInfo, filePath string) []*models.Violation {
	var violations []*models.Violation

	if goAstInfo.AST == nil || goAstInfo.AST.Name == nil {
		return violations
	}

	packageName := goAstInfo.AST.Name.Name
	
	// Package names should be lowercase
	if !d.isLowercase(packageName) {
		violations = append(violations, &models.Violation{
			Type:        models.ViolationTypeNaming,
			Severity:    models.SeverityMedium,
			Message:     fmt.Sprintf("Package '%s' should use lowercase letters only", packageName),
			File:        filePath,
			Line:        int(goAstInfo.FileSet.Position(goAstInfo.AST.Name.Pos()).Line),
			Column:      int(goAstInfo.FileSet.Position(goAstInfo.AST.Name.Pos()).Column),
			Rule:        goStandardNamingRule,
			Suggestion:  "Use lowercase letters for package names",
			CodeSnippet: fmt.Sprintf("package %s", packageName),
		})
	}

	// Package names should not contain underscores
	if strings.Contains(packageName, "_") {
		violations = append(violations, &models.Violation{
			Type:        models.ViolationTypeNaming,
			Severity:    models.SeverityMedium,
			Message:     fmt.Sprintf("Package '%s' should not contain underscores", packageName),
			File:        filePath,
			Line:        int(goAstInfo.FileSet.Position(goAstInfo.AST.Name.Pos()).Line),
			Column:      int(goAstInfo.FileSet.Position(goAstInfo.AST.Name.Pos()).Column),
			Rule:        goStandardNamingRule,
			Suggestion:  "Remove underscores from package names and use short, descriptive names",
			CodeSnippet: fmt.Sprintf("package %s", packageName),
		})
	}

	return violations
}

// checkFunctionNaming validates function naming conventions
func (d *GoStandardNamingDetector) checkFunctionNaming(fn *types.FunctionInfo, filePath string) []*models.Violation {
	var violations []*models.Violation

	// Check MixedCaps convention
	if !d.isMixedCaps(fn.Name) {
		violations = append(violations, &models.Violation{
			Type:        models.ViolationTypeNaming,
			Severity:    d.getSeverityForExported(fn.IsExported, models.SeverityMedium, models.SeverityLow),
			Message:     fmt.Sprintf("Function '%s' should use MixedCaps (camelCase/PascalCase)", fn.Name),
			File:        filePath,
			Line:        fn.StartLine,
			Column:      fn.StartColumn,
			Rule:        goStandardNamingRule,
			Suggestion:  d.suggestMixedCaps(fn.Name, fn.IsExported),
			CodeSnippet: d.generateFunctionSnippet(fn),
		})
	}

	// Check proper exported/unexported casing
	if !d.hasProperExportCasing(fn.Name, fn.IsExported) {
		violations = append(violations, &models.Violation{
			Type:        models.ViolationTypeNaming,
			Severity:    models.SeverityHigh,
			Message:     fmt.Sprintf("Function '%s' has incorrect export casing", fn.Name),
			File:        filePath,
			Line:        fn.StartLine,
			Column:      fn.StartColumn,
			Rule:        goStandardNamingRule,
			Suggestion:  d.suggestCorrectExportCasing(fn.Name, fn.IsExported),
			CodeSnippet: d.generateFunctionSnippet(fn),
		})
	}

	// Check for underscores (not allowed in Go naming)
	if strings.Contains(fn.Name, "_") {
		violations = append(violations, &models.Violation{
			Type:        models.ViolationTypeNaming,
			Severity:    models.SeverityMedium,
			Message:     fmt.Sprintf("Function '%s' should not contain underscores", fn.Name),
			File:        filePath,
			Line:        fn.StartLine,
			Column:      fn.StartColumn,
			Rule:        goStandardNamingRule,
			Suggestion:  "Use MixedCaps instead of underscores",
			CodeSnippet: d.generateFunctionSnippet(fn),
		})
	}

	// Check for inappropriate "get" prefix
	if d.hasInappropriateGetPrefix(fn.Name) {
		violations = append(violations, &models.Violation{
			Type:        models.ViolationTypeNaming,
			Severity:    models.SeverityLow,
			Message:     fmt.Sprintf("Function '%s' should avoid 'get' prefix for getters", fn.Name),
			File:        filePath,
			Line:        fn.StartLine,
			Column:      fn.StartColumn,
			Rule:        goStandardNamingRule,
			Suggestion:  fmt.Sprintf("Consider renaming to '%s'", d.removeGetPrefix(fn.Name)),
			CodeSnippet: d.generateFunctionSnippet(fn),
		})
	}

	return violations
}

// checkTypeNaming validates type naming conventions
func (d *GoStandardNamingDetector) checkTypeNaming(typeInfo *types.TypeInfo, filePath string) []*models.Violation {
	var violations []*models.Violation

	// Check MixedCaps convention
	if !d.isMixedCaps(typeInfo.Name) {
		violations = append(violations, &models.Violation{
			Type:        models.ViolationTypeNaming,
			Severity:    d.getSeverityForExported(typeInfo.IsExported, models.SeverityMedium, models.SeverityLow),
			Message:     fmt.Sprintf("Type '%s' should use MixedCaps (CamelCase)", typeInfo.Name),
			File:        filePath,
			Line:        typeInfo.StartLine,
			Column:      typeInfo.StartColumn,
			Rule:        goStandardNamingRule,
			Suggestion:  d.suggestMixedCaps(typeInfo.Name, typeInfo.IsExported),
			CodeSnippet: fmt.Sprintf("type %s", typeInfo.Name),
		})
	}

	// Check proper exported/unexported casing
	if !d.hasProperExportCasing(typeInfo.Name, typeInfo.IsExported) {
		violations = append(violations, &models.Violation{
			Type:        models.ViolationTypeNaming,
			Severity:    models.SeverityHigh,
			Message:     fmt.Sprintf("Type '%s' has incorrect export casing", typeInfo.Name),
			File:        filePath,
			Line:        typeInfo.StartLine,
			Column:      typeInfo.StartColumn,
			Rule:        goStandardNamingRule,
			Suggestion:  d.suggestCorrectExportCasing(typeInfo.Name, typeInfo.IsExported),
			CodeSnippet: fmt.Sprintf("type %s", typeInfo.Name),
		})
	}

	// Check for underscores
	if strings.Contains(typeInfo.Name, "_") {
		violations = append(violations, &models.Violation{
			Type:        models.ViolationTypeNaming,
			Severity:    models.SeverityMedium,
			Message:     fmt.Sprintf("Type '%s' should not contain underscores", typeInfo.Name),
			File:        filePath,
			Line:        typeInfo.StartLine,
			Column:      typeInfo.StartColumn,
			Rule:        goStandardNamingRule,
			Suggestion:  "Use MixedCaps instead of underscores",
			CodeSnippet: fmt.Sprintf("type %s", typeInfo.Name),
		})
	}

	// Check interface naming convention (-er suffix for single-method interfaces)
	if d.isInterface(typeInfo) && d.shouldHaveErSuffix(typeInfo) {
		violations = append(violations, d.checkInterfaceNaming(typeInfo, filePath)...)
	}

	return violations
}

// checkVariableNaming validates variable naming conventions
func (d *GoStandardNamingDetector) checkVariableNaming(variable *types.VariableInfo, filePath string) []*models.Violation {
	var violations []*models.Violation

	// Check MixedCaps convention
	if !d.isMixedCaps(variable.Name) {
		violations = append(violations, &models.Violation{
			Type:        models.ViolationTypeNaming,
			Severity:    d.getSeverityForExported(variable.IsExported, models.SeverityMedium, models.SeverityLow),
			Message:     fmt.Sprintf("Variable '%s' should use MixedCaps (camelCase/PascalCase)", variable.Name),
			File:        filePath,
			Line:        variable.Line,
			Column:      variable.Column,
			Rule:        goStandardNamingRule,
			Suggestion:  d.suggestMixedCaps(variable.Name, variable.IsExported),
			CodeSnippet: fmt.Sprintf("var %s", variable.Name),
		})
	}

	// Check proper exported/unexported casing
	if !d.hasProperExportCasing(variable.Name, variable.IsExported) {
		violations = append(violations, &models.Violation{
			Type:        models.ViolationTypeNaming,
			Severity:    models.SeverityHigh,
			Message:     fmt.Sprintf("Variable '%s' has incorrect export casing", variable.Name),
			File:        filePath,
			Line:        variable.Line,
			Column:      variable.Column,
			Rule:        goStandardNamingRule,
			Suggestion:  d.suggestCorrectExportCasing(variable.Name, variable.IsExported),
			CodeSnippet: fmt.Sprintf("var %s", variable.Name),
		})
	}

	// Check for underscores
	if strings.Contains(variable.Name, "_") {
		violations = append(violations, &models.Violation{
			Type:        models.ViolationTypeNaming,
			Severity:    models.SeverityMedium,
			Message:     fmt.Sprintf("Variable '%s' should not contain underscores", variable.Name),
			File:        filePath,
			Line:        variable.Line,
			Column:      variable.Column,
			Rule:        goStandardNamingRule,
			Suggestion:  "Use MixedCaps instead of underscores",
			CodeSnippet: fmt.Sprintf("var %s", variable.Name),
		})
	}

	// Check conventional patterns for boolean variables
	if d.isBooleanType(variable.Type) && !d.hasGoodBooleanName(variable.Name) {
		violations = append(violations, &models.Violation{
			Type:        models.ViolationTypeNaming,
			Severity:    models.SeverityLow,
			Message:     fmt.Sprintf("Boolean variable '%s' should use conventional prefixes (Is, Has, Can, Allow)", variable.Name),
			File:        filePath,
			Line:        variable.Line,
			Column:      variable.Column,
			Rule:        goStandardNamingRule,
			Suggestion:  "Consider prefixes like 'is', 'has', 'can', or 'allow' for boolean variables",
			CodeSnippet: fmt.Sprintf("var %s bool", variable.Name),
		})
	}

	// Check for error variable naming convention
	if d.isErrorType(variable.Type) {
		if variable.IsExported {
			// Exported sentinel errors conventionally use 'Err' prefix (e.g., ErrNotFound)
			if !strings.HasPrefix(variable.Name, "Err") {
				violations = append(violations, &models.Violation{
					Type:        models.ViolationTypeNaming,
					Severity:    models.SeverityLow,
					Message:     fmt.Sprintf("Exported error variable '%s' should be prefixed with 'Err'", variable.Name),
					File:        filePath,
					Line:        variable.Line,
					Column:      variable.Column,
					Rule:        goStandardNamingRule,
					Suggestion:  "Use 'ErrX' style for exported sentinel errors (e.g., ErrNotFound); use 'err' for local error variables",
					CodeSnippet: fmt.Sprintf("var %s error", variable.Name),
				})
			}
		} else if variable.Name != "err" {
			// Local errors should be named 'err'
			violations = append(violations, &models.Violation{
				Type:        models.ViolationTypeNaming,
				Severity:    models.SeverityLow,
				Message:     fmt.Sprintf("Local error variable '%s' should be named 'err'", variable.Name),
				File:        filePath,
				Line:        variable.Line,
				Column:      variable.Column,
				Rule:        goStandardNamingRule,
				Suggestion:  "Name local error variables 'err'",
				CodeSnippet: fmt.Sprintf("var %s error", variable.Name),
			})
		}
	}

	return violations
}

// checkConstantNaming validates constant naming conventions
func (d *GoStandardNamingDetector) checkConstantNaming(constant *types.ConstantInfo, filePath string) []*models.Violation {
	var violations []*models.Violation

	if constant.IsExported {
		// Exported constants should use SCREAMING_SNAKE_CASE
		if !d.isScreamingSnakeCase(constant.Name) {
			violations = append(violations, &models.Violation{
				Type:        models.ViolationTypeNaming,
				Severity:    models.SeverityMedium,
				Message:     fmt.Sprintf("Exported constant '%s' should use SCREAMING_SNAKE_CASE", constant.Name),
				File:        filePath,
				Line:        constant.Line,
				Column:      constant.Column,
				Rule:        goStandardNamingRule,
				Suggestion:  fmt.Sprintf("Consider renaming to '%s'", d.toScreamingSnakeCase(constant.Name)),
				CodeSnippet: fmt.Sprintf("const %s", constant.Name),
			})
		}
	} else {
		// Unexported constants should use MixedCaps
		if !d.isMixedCaps(constant.Name) {
			violations = append(violations, &models.Violation{
				Type:        models.ViolationTypeNaming,
				Severity:    models.SeverityLow,
				Message:     fmt.Sprintf("Unexported constant '%s' should use MixedCaps", constant.Name),
				File:        filePath,
				Line:        constant.Line,
				Column:      constant.Column,
				Rule:        goStandardNamingRule,
				Suggestion:  d.suggestMixedCaps(constant.Name, false),
				CodeSnippet: fmt.Sprintf("const %s", constant.Name),
			})
		}
	}

	return violations
}

// Helper methods for validation logic

// isLowercase checks if a string is all lowercase
func (d *GoStandardNamingDetector) isLowercase(s string) bool {
	return strings.ToLower(s) == s
}

// isMixedCaps checks if a name follows MixedCaps convention
func (d *GoStandardNamingDetector) isMixedCaps(name string) bool {
	if name == "" {
		return false
	}

	// Must start with a letter
	if !unicode.IsLetter(rune(name[0])) {
		return false
	}

	// Should not contain underscores or hyphens
	if strings.ContainsAny(name, "_-") {
		return false
	}

	// Should not have consecutive uppercase letters (except for acronyms)
	return !d.hasInvalidConsecutiveCase(name)
}

// hasProperExportCasing checks if the name has correct export casing
func (d *GoStandardNamingDetector) hasProperExportCasing(name string, isExported bool) bool {
	if name == "" {
		return false
	}

	firstChar := rune(name[0])
	if isExported {
		return unicode.IsUpper(firstChar)
	}
	return unicode.IsLower(firstChar)
}

// isScreamingSnakeCase checks if a name is in SCREAMING_SNAKE_CASE
func (d *GoStandardNamingDetector) isScreamingSnakeCase(name string) bool {
	if name == "" {
		return false
	}
	
	// Should be all uppercase with underscores
	pattern := regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)
	return pattern.MatchString(name)
}

// getSeverityForExported returns different severity based on export status
func (d *GoStandardNamingDetector) getSeverityForExported(isExported bool, exportedSeverity, unexportedSeverity models.Severity) models.Severity {
	if isExported {
		return exportedSeverity
	}
	return unexportedSeverity
}

// suggestMixedCaps suggests a MixedCaps version of a name
func (d *GoStandardNamingDetector) suggestMixedCaps(name string, isExported bool) string {
	if name == "" {
		return ""
	}

	// Convert snake_case or kebab-case to MixedCaps
	parts := regexp.MustCompile(`[_-]+`).Split(name, -1)
	result := ""
	
	for i, part := range parts {
		if part == "" {
			continue
		}
		
		if i == 0 && !isExported {
			// First word lowercase for unexported
			result += strings.ToLower(part)
		} else {
			// Capitalize first letter
			result += strings.ToUpper(string(part[0])) + strings.ToLower(part[1:])
		}
	}
	
	return fmt.Sprintf("Consider renaming to '%s'", result)
}

// suggestCorrectExportCasing suggests correct export casing
func (d *GoStandardNamingDetector) suggestCorrectExportCasing(name string, isExported bool) string {
	if name == "" {
		return ""
	}

	if isExported {
		return fmt.Sprintf("Should start with uppercase: '%s'", strings.ToUpper(string(name[0]))+name[1:])
	}
	return fmt.Sprintf("Should start with lowercase: '%s'", strings.ToLower(string(name[0]))+name[1:])
}

// Additional helper methods...

// hasInvalidConsecutiveCase checks for invalid consecutive uppercase letters
func (d *GoStandardNamingDetector) hasInvalidConsecutiveCase(name string) bool {
	// Treat runs of 3+ uppercase letters as valid if they are common initialisms (HTTP, URL, JSON, etc.)
	runStart := -1
	for i, r := range name {
		if unicode.IsUpper(r) {
			if runStart == -1 {
				runStart = i
			}
			continue
		}
		if runStart != -1 {
			if i-runStart >= 3 {
				seq := name[runStart:i]
				if _, ok := commonInitialisms[seq]; !ok {
					return true
				}
			}
			runStart = -1
		}
	}
	// Trailing run
	if runStart != -1 && len(name)-runStart >= 3 {
		seq := name[runStart:]
		if _, ok := commonInitialisms[seq]; !ok {
			return true
		}
	}
	return false
}

// hasInappropriateGetPrefix checks for inappropriate "get" prefix
func (d *GoStandardNamingDetector) hasInappropriateGetPrefix(name string) bool {
	return strings.HasPrefix(strings.ToLower(name), "get") && len(name) > 3
}

// removeGetPrefix removes "get" prefix and adjusts casing
func (d *GoStandardNamingDetector) removeGetPrefix(name string) string {
	if !d.hasInappropriateGetPrefix(name) {
		return name
	}
	
	remaining := name[3:] // Remove "get"
	if len(remaining) == 0 {
		return name
	}
	
	// Keep the same export status
	if unicode.IsUpper(rune(name[0])) {
		return strings.ToUpper(string(remaining[0])) + remaining[1:]
	}
	return strings.ToLower(string(remaining[0])) + remaining[1:]
}

// toScreamingSnakeCase converts a name to SCREAMING_SNAKE_CASE
func (d *GoStandardNamingDetector) toScreamingSnakeCase(name string) string {
	// Simple conversion - insert underscores before uppercase letters and convert to uppercase
	result := ""
	for i, r := range name {
		if unicode.IsUpper(r) && i > 0 {
			result += "_"
		}
		result += strings.ToUpper(string(r))
	}
	return result
}

// isInterface checks if a type is an interface
func (d *GoStandardNamingDetector) isInterface(typeInfo *types.TypeInfo) bool {
	return typeInfo.Kind == "interface"
}

// shouldHaveErSuffix checks if an interface should have -er suffix
func (d *GoStandardNamingDetector) shouldHaveErSuffix(typeInfo *types.TypeInfo) bool {
	// Only single-method interfaces conventionally use -er suffix
	return typeInfo.IsExported && typeInfo.MethodCount == 1 && 
		!strings.HasSuffix(typeInfo.Name, "er") && !strings.HasSuffix(typeInfo.Name, "or")
}

// checkInterfaceNaming validates interface naming patterns
func (d *GoStandardNamingDetector) checkInterfaceNaming(typeInfo *types.TypeInfo, filePath string) []*models.Violation {
	var violations []*models.Violation

	// This is a suggestion rather than a strict rule
	violations = append(violations, &models.Violation{
		Type:        models.ViolationTypeNaming,
		Severity:    models.SeverityLow,
		Message:     fmt.Sprintf("Interface '%s' conventionally uses -er suffix for single-method interfaces", typeInfo.Name),
		File:        filePath,
		Line:        typeInfo.StartLine,
		Column:      typeInfo.StartColumn,
		Rule:        goStandardNamingRule,
		Suggestion:  "Consider adding -er suffix if this is a single-method interface",
		CodeSnippet: fmt.Sprintf("type %s interface", typeInfo.Name),
	})

	return violations
}

// isBooleanType checks if a variable type is boolean
func (d *GoStandardNamingDetector) isBooleanType(varType string) bool {
	return varType == "bool"
}

// hasGoodBooleanName checks if a boolean variable has a conventional name
func (d *GoStandardNamingDetector) hasGoodBooleanName(name string) bool {
	lowername := strings.ToLower(name)
	prefixes := []string{"is", "has", "can", "allow", "should", "will", "enable", "disable"}
	
	for _, prefix := range prefixes {
		if strings.HasPrefix(lowername, prefix) {
			return true
		}
	}
	
	// Some boolean variables are fine without prefixes
	commonBooleans := []string{"ok", "found", "done", "valid", "ready", "active", "enabled", "disabled"}
	for _, common := range commonBooleans {
		if lowername == common {
			return true
		}
	}
	
	return false
}

// isErrorType checks if a variable type is error
func (d *GoStandardNamingDetector) isErrorType(varType string) bool {
	return varType == "error" || strings.HasSuffix(varType, "Error")
}

// generateFunctionSnippet generates a code snippet for function violations
func (d *GoStandardNamingDetector) generateFunctionSnippet(fn *types.FunctionInfo) string {
	return fmt.Sprintf("func %s(...)", fn.Name)
}