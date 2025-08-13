package violations

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/scanner"
)

// NamingDetector detects naming convention violations
type NamingDetector struct {
	config *DetectorConfig
}

// NewNamingDetector creates a new naming convention detector
func NewNamingDetector(config *DetectorConfig) *NamingDetector {
	if config == nil {
		config = DefaultDetectorConfig()
	}
	return &NamingDetector{
		config: config,
	}
}

// Name returns the name of this detector
func (d *NamingDetector) Name() string {
	return "Naming Convention Analysis"
}

// Description returns a description of what this detector checks for
func (d *NamingDetector) Description() string {
	return "Detects violations of naming conventions including non-descriptive names, inconsistent casing, and inappropriate naming patterns"
}

// Detect analyzes naming conventions and returns violations
func (d *NamingDetector) Detect(fileInfo *models.FileInfo, astInfo *scanner.GoASTInfo) []*models.Violation {
	var violations []*models.Violation

	if astInfo == nil {
		// For non-Go files, we have limited naming analysis
		return violations
	}

	// Check function names
	for _, function := range astInfo.Functions {
		violations = append(violations, d.checkFunctionNaming(function, fileInfo.Path)...)
		// Check parameter names
		violations = append(violations, d.checkParameterNaming(function, fileInfo.Path)...)
	}

	// Check type names
	for _, typeInfo := range astInfo.Types {
		violations = append(violations, d.checkTypeNaming(typeInfo, fileInfo.Path)...)
	}

	// Check variable names
	for _, variable := range astInfo.Variables {
		violations = append(violations, d.checkVariableNaming(variable, fileInfo.Path)...)
	}

	// Check constant names
	for _, constant := range astInfo.Constants {
		violations = append(violations, d.checkConstantNaming(constant, fileInfo.Path)...)
	}

	return violations
}

// checkFunctionNaming analyzes function names for violations
func (d *NamingDetector) checkFunctionNaming(fn *scanner.FunctionInfo, filePath string) []*models.Violation {
	var violations []*models.Violation

	// Check for non-descriptive names
	if d.isNonDescriptiveName(fn.Name) {
		violations = append(violations, &models.Violation{
			Type:        models.ViolationTypeNaming,
			Severity:    models.SeverityMedium,
			Message:     fmt.Sprintf("Function '%s' has a non-descriptive name", fn.Name),
			File:        filePath,
			Line:        fn.StartLine,
			Column:      fn.StartColumn,
			Rule:        "non-descriptive-function-name",
			Suggestion:  fmt.Sprintf("Choose a more descriptive name that clearly indicates what the function does"),
			CodeSnippet: d.generateFunctionNameSnippet(fn),
		})
	}

	// Check for improper casing (Go specific)
	if !d.isProperGoFunctionCase(fn.Name, fn.IsExported) {
		severity := models.SeverityLow
		if fn.IsExported {
			severity = models.SeverityMedium
		}
		
		violations = append(violations, &models.Violation{
			Type:        models.ViolationTypeNaming,
			Severity:    severity,
			Message:     fmt.Sprintf("Function '%s' does not follow Go naming conventions", fn.Name),
			File:        filePath,
			Line:        fn.StartLine,
			Column:      fn.StartColumn,
			Rule:        "go-function-case",
			Suggestion:  d.getGoCasingSuggestion(fn.Name, fn.IsExported, "function"),
			CodeSnippet: d.generateFunctionNameSnippet(fn),
		})
	}

	// Check for inappropriate abbreviations
	if d.hasInappropriateAbbreviation(fn.Name) {
		violations = append(violations, &models.Violation{
			Type:        models.ViolationTypeNaming,
			Severity:    models.SeverityLow,
			Message:     fmt.Sprintf("Function '%s' contains unclear abbreviations", fn.Name),
			File:        filePath,
			Line:        fn.StartLine,
			Column:      fn.StartColumn,
			Rule:        "unclear-abbreviation",
			Suggestion:  "Consider spelling out abbreviations for better readability",
			CodeSnippet: d.generateFunctionNameSnippet(fn),
		})
	}

	return violations
}

// checkParameterNaming analyzes parameter names for violations
func (d *NamingDetector) checkParameterNaming(fn *scanner.FunctionInfo, filePath string) []*models.Violation {
	var violations []*models.Violation

	for _, param := range fn.Parameters {
		if param.Name == "" {
			continue // Unnamed parameter, skip
		}

		// Check for single letter variables (if not allowed)
		if !d.config.AllowSingleLetterVars && d.isSingleLetterVar(param.Name) {
			violations = append(violations, &models.Violation{
				Type:        models.ViolationTypeNaming,
				Severity:    models.SeverityLow,
				Message:     fmt.Sprintf("Parameter '%s' in function '%s' uses single letter naming", param.Name, fn.Name),
				File:        filePath,
				Line:        fn.StartLine,
				Column:      fn.StartColumn,
				Rule:        "single-letter-parameter",
				Suggestion:  fmt.Sprintf("Use a more descriptive name for parameter '%s'", param.Name),
				CodeSnippet: d.generateParameterSnippet(fn, param.Name),
			})
		}

		// Check for non-descriptive parameter names
		if d.isNonDescriptiveName(param.Name) && !d.isCommonShortParam(param.Name) {
			violations = append(violations, &models.Violation{
				Type:        models.ViolationTypeNaming,
				Severity:    models.SeverityLow,
				Message:     fmt.Sprintf("Parameter '%s' in function '%s' has a non-descriptive name", param.Name, fn.Name),
				File:        filePath,
				Line:        fn.StartLine,
				Column:      fn.StartColumn,
				Rule:        "non-descriptive-parameter",
				Suggestion:  fmt.Sprintf("Choose a more descriptive name for parameter '%s'", param.Name),
				CodeSnippet: d.generateParameterSnippet(fn, param.Name),
			})
		}
	}

	return violations
}

// checkTypeNaming analyzes type names for violations
func (d *NamingDetector) checkTypeNaming(typeInfo *scanner.TypeInfo, filePath string) []*models.Violation {
	var violations []*models.Violation

	// Check for non-descriptive type names
	if d.isNonDescriptiveName(typeInfo.Name) {
		violations = append(violations, &models.Violation{
			Type:        models.ViolationTypeNaming,
			Severity:    models.SeverityMedium,
			Message:     fmt.Sprintf("Type '%s' has a non-descriptive name", typeInfo.Name),
			File:        filePath,
			Line:        typeInfo.StartLine,
			Column:      typeInfo.StartColumn,
			Rule:        "non-descriptive-type-name",
			Suggestion:  "Choose a more descriptive type name that clearly indicates its purpose",
			CodeSnippet: fmt.Sprintf("type %s", typeInfo.Name),
		})
	}

	// Check for improper Go type naming
	if !d.isProperGoTypeCase(typeInfo.Name, typeInfo.IsExported) {
		severity := models.SeverityMedium
		if typeInfo.IsExported {
			severity = models.SeverityHigh
		}

		violations = append(violations, &models.Violation{
			Type:        models.ViolationTypeNaming,
			Severity:    severity,
			Message:     fmt.Sprintf("Type '%s' does not follow Go naming conventions", typeInfo.Name),
			File:        filePath,
			Line:        typeInfo.StartLine,
			Column:      typeInfo.StartColumn,
			Rule:        "go-type-case",
			Suggestion:  d.getGoCasingSuggestion(typeInfo.Name, typeInfo.IsExported, "type"),
			CodeSnippet: fmt.Sprintf("type %s", typeInfo.Name),
		})
	}

	return violations
}

// checkVariableNaming analyzes variable names for violations
func (d *NamingDetector) checkVariableNaming(variable *scanner.VariableInfo, filePath string) []*models.Violation {
	var violations []*models.Violation

	// Check for single letter variables (if not allowed)
	if !d.config.AllowSingleLetterVars && d.isSingleLetterVar(variable.Name) {
		violations = append(violations, &models.Violation{
			Type:        models.ViolationTypeNaming,
			Severity:    models.SeverityLow,
			Message:     fmt.Sprintf("Variable '%s' uses single letter naming", variable.Name),
			File:        filePath,
			Line:        variable.Line,
			Column:      variable.Column,
			Rule:        "single-letter-variable",
			Suggestion:  fmt.Sprintf("Use a more descriptive name for variable '%s'", variable.Name),
			CodeSnippet: fmt.Sprintf("var %s", variable.Name),
		})
	}

	// Check for improper Go variable naming
	if !d.isProperGoVariableCase(variable.Name, variable.IsExported) {
		severity := models.SeverityLow
		if variable.IsExported {
			severity = models.SeverityMedium
		}

		violations = append(violations, &models.Violation{
			Type:        models.ViolationTypeNaming,
			Severity:    severity,
			Message:     fmt.Sprintf("Variable '%s' does not follow Go naming conventions", variable.Name),
			File:        filePath,
			Line:        variable.Line,
			Column:      variable.Column,
			Rule:        "go-variable-case",
			Suggestion:  d.getGoCasingSuggestion(variable.Name, variable.IsExported, "variable"),
			CodeSnippet: fmt.Sprintf("var %s", variable.Name),
		})
	}

	return violations
}

// checkConstantNaming analyzes constant names for violations
func (d *NamingDetector) checkConstantNaming(constant *scanner.ConstantInfo, filePath string) []*models.Violation {
	var violations []*models.Violation

	// Check for improper Go constant naming
	if !d.isProperGoConstantCase(constant.Name, constant.IsExported) {
		severity := models.SeverityLow
		if constant.IsExported {
			severity = models.SeverityMedium
		}

		violations = append(violations, &models.Violation{
			Type:        models.ViolationTypeNaming,
			Severity:    severity,
			Message:     fmt.Sprintf("Constant '%s' does not follow Go naming conventions", constant.Name),
			File:        filePath,
			Line:        constant.Line,
			Column:      constant.Column,
			Rule:        "go-constant-case",
			Suggestion:  d.getGoCasingSuggestion(constant.Name, constant.IsExported, "constant"),
			CodeSnippet: fmt.Sprintf("const %s", constant.Name),
		})
	}

	return violations
}

// Helper methods for naming analysis

func (d *NamingDetector) isNonDescriptiveName(name string) bool {
	// Common non-descriptive patterns
	nonDescriptivePatterns := []string{
		"^[a-z]$",           // Single letters
		"^(data|info|item|obj|thing|stuff|temp|tmp)\\d*$", // Generic names
		"^[a-z][0-9]+$",     // Letter followed by numbers
		"^(foo|bar|baz|qux)\\d*$", // Placeholder names
	}

	for _, pattern := range nonDescriptivePatterns {
		if matched, _ := regexp.MatchString(pattern, strings.ToLower(name)); matched {
			return true
		}
	}

	return false
}

func (d *NamingDetector) isSingleLetterVar(name string) bool {
	return len(name) == 1 && unicode.IsLetter(rune(name[0]))
}

func (d *NamingDetector) isCommonShortParam(name string) bool {
	// Common acceptable short parameter names in Go
	commonShortParams := map[string]bool{
		"i": true, "j": true, "k": true, // Loop counters
		"x": true, "y": true, "z": true, // Coordinates/math
		"n": true, "m": true,             // Counts/sizes
		"r": true, "w": true,             // Readers/writers
		"t": true,                        // Time/testing
		"b": true,                        // Bytes/boolean
		"s": true,                        // String
		"v": true,                        // Value
		"ok": true,                       // Boolean results
		"id": true,                       // Identifier
	}
	return commonShortParams[strings.ToLower(name)]
}

func (d *NamingDetector) hasInappropriateAbbreviation(name string) bool {
	// Common problematic abbreviations - check for exact word boundaries
	problematicAbbrevs := []string{
		"mgr", "mng", "mgmt",    // Manager
		"calc", "comp", "proc",  // Calculate, Compute, Process
		"str", "strg",           // String, Storage
		"num", "nbr",            // Number
		"addr",                  // Address
		"cfg", "conf",           // Config
		"btn",                   // Button
		"img",                   // Image
		"req", "res", "resp",    // Request, Response
	}

	lowerName := strings.ToLower(name)
	
	// Check if the abbreviation appears as a complete word/part, not just a substring
	for _, abbrev := range problematicAbbrevs {
		// Use word boundary patterns to avoid false positives
		// Check if abbreviation appears at start, end, or between case changes
		pattern := fmt.Sprintf("(^%s|%s$|%s[A-Z]|[a-z]%s[A-Z])", abbrev, abbrev, abbrev, abbrev)
		if matched, _ := regexp.MatchString(pattern, lowerName); matched {
			// Double-check it's not part of a full word
			if !d.isPartOfFullWord(name, abbrev) {
				return true
			}
		}
	}
	return false
}

// isPartOfFullWord checks if the abbreviation is part of a complete word
func (d *NamingDetector) isPartOfFullWord(name, abbrev string) bool {
	fullWords := map[string][]string{
		"mgr": {"manager"},
		"calc": {"calculate", "calculation"},
		"str": {"string", "structure", "stream"},
		"addr": {"address"},
		"proc": {"process", "processor"},
	}
	
	if fullWordList, exists := fullWords[abbrev]; exists {
		lowerName := strings.ToLower(name)
		for _, fullWord := range fullWordList {
			if strings.Contains(lowerName, fullWord) {
				return true
			}
		}
	}
	
	return false
}

// Go-specific naming convention checks

func (d *NamingDetector) isProperGoFunctionCase(name string, isExported bool) bool {
	if len(name) == 0 {
		return false
	}

	// Go functions should be camelCase or PascalCase
	if isExported {
		// Exported functions should start with uppercase
		return unicode.IsUpper(rune(name[0])) && d.isCamelCase(name)
	} else {
		// Unexported functions should start with lowercase
		return unicode.IsLower(rune(name[0])) && d.isCamelCase(name)
	}
}

func (d *NamingDetector) isProperGoTypeCase(name string, isExported bool) bool {
	if len(name) == 0 {
		return false
	}

	// Go types follow the same rules as functions
	return d.isProperGoFunctionCase(name, isExported)
}

func (d *NamingDetector) isProperGoVariableCase(name string, isExported bool) bool {
	if len(name) == 0 {
		return false
	}

	// Go variables follow the same rules as functions
	return d.isProperGoFunctionCase(name, isExported)
}

func (d *NamingDetector) isProperGoConstantCase(name string, isExported bool) bool {
	if len(name) == 0 {
		return false
	}

	// Go constants can be camelCase/PascalCase or ALL_CAPS for some cases
	if d.isAllCapsWithUnderscores(name) {
		return true // ALL_CAPS acceptable for constants
	}

	// Otherwise follow normal casing rules
	return d.isProperGoFunctionCase(name, isExported)
}

func (d *NamingDetector) isCamelCase(name string) bool {
	// Check if the name follows camelCase or PascalCase pattern
	// Should not have underscores or consecutive uppercase letters (except acronyms)
	if strings.Contains(name, "_") {
		return false
	}

	// Allow for common acronyms like HTTP, URL, JSON, etc.
	acronymPattern := regexp.MustCompile(`[A-Z]{2,}`)
	if acronymPattern.MatchString(name) {
		// Check if it's a known acronym at the end or followed by lowercase
		return d.hasValidAcronym(name)
	}

	return true
}

func (d *NamingDetector) isAllCapsWithUnderscores(name string) bool {
	// Check if name is ALL_CAPS format
	pattern := regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)
	return pattern.MatchString(name)
}

func (d *NamingDetector) hasValidAcronym(name string) bool {
	// Common Go acronyms
	commonAcronyms := []string{
		"HTTP", "HTTPS", "URL", "URI", "XML", "JSON", "API", "SQL", "UUID",
		"CPU", "RAM", "IO", "OS", "DB", "ID", "TCP", "UDP", "IP", "DNS",
	}

	for _, acronym := range commonAcronyms {
		if strings.Contains(name, acronym) {
			return true
		}
	}
	return false
}

func (d *NamingDetector) getGoCasingSuggestion(name string, isExported bool, itemType string) string {
	if isExported {
		return fmt.Sprintf("Exported %s names should start with uppercase and use PascalCase (e.g., MyFunction)", itemType)
	} else {
		return fmt.Sprintf("Unexported %s names should start with lowercase and use camelCase (e.g., myFunction)", itemType)
	}
}

// Code snippet generation helpers

func (d *NamingDetector) generateFunctionNameSnippet(fn *scanner.FunctionInfo) string {
	if fn.IsMethod && fn.ReceiverType != "" {
		return fmt.Sprintf("func (%s) %s", fn.ReceiverType, fn.Name)
	}
	return fmt.Sprintf("func %s", fn.Name)
}

func (d *NamingDetector) generateParameterSnippet(fn *scanner.FunctionInfo, paramName string) string {
	for _, param := range fn.Parameters {
		if param.Name == paramName {
			return fmt.Sprintf("%s %s", param.Name, param.Type)
		}
	}
	return paramName
}