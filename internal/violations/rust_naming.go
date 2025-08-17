package violations

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/types"
	"github.com/ericfisherdev/goclean/internal/violations/morphology"
)

// Rust-specific violation rule constants
const (
	// Function naming violations
	RustInvalidFunctionNaming = "rust-invalid-function-naming"
	RustInvalidMethodNaming   = "rust-invalid-method-naming"
	
	// Type naming violations
	RustInvalidStructNaming = "rust-invalid-struct-naming"
	RustInvalidEnumNaming   = "rust-invalid-enum-naming"
	RustInvalidTraitNaming  = "rust-invalid-trait-naming"
	RustInvalidTypeAlias    = "rust-invalid-type-alias"
	
	// Constant/Static naming violations
	RustInvalidConstantNaming = "rust-invalid-constant-naming"
	RustInvalidStaticNaming   = "rust-invalid-static-naming"
	
	// Module naming violations
	RustInvalidModuleNaming = "rust-invalid-module-naming"
	RustInvalidCrateNaming  = "rust-invalid-crate-naming"
	
	// Variable naming violations
	RustInvalidVariableNaming  = "rust-invalid-variable-naming"
	RustInvalidParameterNaming = "rust-invalid-parameter-naming"
	
	// Other naming violations
	RustNonDescriptiveName     = "rust-non-descriptive-name"
	RustAcronymCasing          = "rust-acronym-casing"
	RustUnclearAbbreviation    = "rust-unclear-abbreviation"
	RustInconsistentNaming     = "rust-inconsistent-naming"
)

// RustNamingDetector detects Rust-specific naming convention violations
type RustNamingDetector struct {
	config           *DetectorConfig
	morphEngine      *morphology.MorphologyEngine
	programmingTerms *morphology.ProgrammingTermAnalyzer
	conventionChecker *RustConventionChecker
}

// NewRustNamingDetector creates a new Rust naming convention detector
func NewRustNamingDetector(config *DetectorConfig) *RustNamingDetector {
	if config == nil {
		config = DefaultDetectorConfig()
	}
	
	// Initialize morphology engine
	morphEngine := morphology.NewMorphologyEngine()
	programmingTerms := morphology.NewProgrammingTermAnalyzer(morphEngine)
	
	return &RustNamingDetector{
		config:           config,
		morphEngine:      morphEngine,
		programmingTerms: programmingTerms,
		conventionChecker: NewRustConventionChecker(),
	}
}

// Name returns the name of this detector
func (d *RustNamingDetector) Name() string {
	return "Rust Naming Convention Analysis"
}

// Description returns a description of what this detector checks for
func (d *RustNamingDetector) Description() string {
	return "Detects violations of Rust naming conventions including snake_case for functions/variables, PascalCase for types, SCREAMING_SNAKE_CASE for constants"
}

// Detect analyzes Rust naming conventions and returns violations
func (d *RustNamingDetector) Detect(fileInfo *models.FileInfo, astInfo interface{}) []*models.Violation {
	var violations []*models.Violation

	if astInfo == nil {
		return violations
	}
	
	// Type assertion to get RustASTInfo
	rustAstInfo, ok := astInfo.(*types.RustASTInfo)
	if !ok || rustAstInfo == nil {
		return violations
	}

	// Check function names
	for _, function := range rustAstInfo.Functions {
		if function != nil {
			violations = append(violations, d.checkFunctionNaming(function, fileInfo.Path)...)
			// Check parameter names
			violations = append(violations, d.checkParameterNaming(function, fileInfo.Path)...)
		}
	}

	// Check struct names
	for _, structInfo := range rustAstInfo.Structs {
		if structInfo != nil {
			violations = append(violations, d.checkStructNaming(structInfo, fileInfo.Path)...)
		}
	}

	// Check enum names
	for _, enumInfo := range rustAstInfo.Enums {
		if enumInfo != nil {
			violations = append(violations, d.checkEnumNaming(enumInfo, fileInfo.Path)...)
		}
	}

	// Check trait names
	for _, traitInfo := range rustAstInfo.Traits {
		if traitInfo != nil {
			violations = append(violations, d.checkTraitNaming(traitInfo, fileInfo.Path)...)
		}
	}

	// Check constant names
	for _, constant := range rustAstInfo.Constants {
		if constant != nil {
			violations = append(violations, d.checkConstantNaming(constant, fileInfo.Path)...)
		}
	}

	// Check module names
	for _, module := range rustAstInfo.Modules {
		if module != nil {
			violations = append(violations, d.checkModuleNaming(module, fileInfo.Path)...)
		}
	}

	return violations
}

// checkFunctionNaming analyzes Rust function names for violations
func (d *RustNamingDetector) checkFunctionNaming(fn *types.RustFunctionInfo, filePath string) []*models.Violation {
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
			Rule:        RustNonDescriptiveName,
			Suggestion:  "Choose a more descriptive name that clearly indicates what the function does",
			CodeSnippet: d.generateFunctionSnippet(fn),
		})
	}

	// Check for proper Rust function naming (snake_case)
	if !d.conventionChecker.IsValidFunctionName(fn.Name) {
		violations = append(violations, &models.Violation{
			Type:        models.ViolationTypeNaming,
			Severity:    models.SeverityMedium,
			Message:     fmt.Sprintf("Function '%s' does not follow Rust naming conventions (should be snake_case)", fn.Name),
			File:        filePath,
			Line:        fn.StartLine,
			Column:      fn.StartColumn,
			Rule:        RustInvalidFunctionNaming,
			Suggestion:  fmt.Sprintf("Rename to '%s'", d.conventionChecker.ToSnakeCase(fn.Name)),
			CodeSnippet: d.generateFunctionSnippet(fn),
		})
	}

	// Check for inappropriate abbreviations
	if d.hasInappropriateAbbreviation(fn.Name) {
		suggestion := d.generateAbbreviationSuggestion(fn.Name)
		violations = append(violations, &models.Violation{
			Type:        models.ViolationTypeNaming,
			Severity:    models.SeverityLow,
			Message:     fmt.Sprintf("Function '%s' contains unclear abbreviations", fn.Name),
			File:        filePath,
			Line:        fn.StartLine,
			Column:      fn.StartColumn,
			Rule:        RustUnclearAbbreviation,
			Suggestion:  suggestion,
			CodeSnippet: d.generateFunctionSnippet(fn),
		})
	}

	return violations
}

// checkParameterNaming analyzes Rust parameter names for violations
func (d *RustNamingDetector) checkParameterNaming(fn *types.RustFunctionInfo, filePath string) []*models.Violation {
	var violations []*models.Violation

	for _, param := range fn.Parameters {
		if param.Name == "" || param.Name == "_" || param.Name == "self" {
			continue // Skip unnamed, ignored, or self parameters
		}

		// Check for single letter variables (if not allowed and not in acceptable context)
		if !d.config.AllowSingleLetterVars && d.isSingleLetterVar(param.Name) && !d.isAcceptableInRustContext(param.Name, fn) {
			violations = append(violations, &models.Violation{
				Type:        models.ViolationTypeNaming,
				Severity:    models.SeverityLow,
				Message:     fmt.Sprintf("Parameter '%s' in function '%s' uses single letter naming", param.Name, fn.Name),
				File:        filePath,
				Line:        fn.StartLine,
				Column:      fn.StartColumn,
				Rule:        RustInvalidParameterNaming,
				Suggestion:  fmt.Sprintf("Use a more descriptive name for parameter '%s'", param.Name),
				CodeSnippet: d.generateParameterSnippet(fn, param),
			})
		}

		// Check for proper Rust parameter naming (snake_case)
		if !d.conventionChecker.IsValidVariableName(param.Name) {
			violations = append(violations, &models.Violation{
				Type:        models.ViolationTypeNaming,
				Severity:    models.SeverityLow,
				Message:     fmt.Sprintf("Parameter '%s' in function '%s' does not follow Rust naming conventions (should be snake_case)", param.Name, fn.Name),
				File:        filePath,
				Line:        fn.StartLine,
				Column:      fn.StartColumn,
				Rule:        RustInvalidParameterNaming,
				Suggestion:  fmt.Sprintf("Rename to '%s'", d.conventionChecker.ToSnakeCase(param.Name)),
				CodeSnippet: d.generateParameterSnippet(fn, param),
			})
		}

		// Check for non-descriptive names (with context consideration)
		if d.isNonDescriptiveName(param.Name) && !d.isAcceptableInRustContext(param.Name, fn) {
			violations = append(violations, &models.Violation{
				Type:        models.ViolationTypeNaming,
				Severity:    models.SeverityLow,
				Message:     fmt.Sprintf("Parameter '%s' in function '%s' has a non-descriptive name", param.Name, fn.Name),
				File:        filePath,
				Line:        fn.StartLine,
				Column:      fn.StartColumn,
				Rule:        RustNonDescriptiveName,
				Suggestion:  fmt.Sprintf("Choose a more descriptive name for parameter '%s'", param.Name),
				CodeSnippet: d.generateParameterSnippet(fn, param),
			})
		}
	}

	return violations
}

// checkStructNaming analyzes Rust struct names for violations
func (d *RustNamingDetector) checkStructNaming(structInfo *types.RustStructInfo, filePath string) []*models.Violation {
	var violations []*models.Violation

	// Check for non-descriptive names
	if d.isNonDescriptiveName(structInfo.Name) {
		violations = append(violations, &models.Violation{
			Type:        models.ViolationTypeNaming,
			Severity:    models.SeverityMedium,
			Message:     fmt.Sprintf("Struct '%s' has a non-descriptive name", structInfo.Name),
			File:        filePath,
			Line:        structInfo.StartLine,
			Column:      structInfo.StartColumn,
			Rule:        RustNonDescriptiveName,
			Suggestion:  "Choose a more descriptive type name that clearly indicates its purpose",
			CodeSnippet: fmt.Sprintf("struct %s", structInfo.Name),
		})
	}

	// Check for proper Rust struct naming (PascalCase)
	if !d.conventionChecker.IsValidTypeName(structInfo.Name) {
		violations = append(violations, &models.Violation{
			Type:        models.ViolationTypeNaming,
			Severity:    models.SeverityHigh,
			Message:     fmt.Sprintf("Struct '%s' does not follow Rust naming conventions (should be PascalCase)", structInfo.Name),
			File:        filePath,
			Line:        structInfo.StartLine,
			Column:      structInfo.StartColumn,
			Rule:        RustInvalidStructNaming,
			Suggestion:  fmt.Sprintf("Rename to '%s'", d.conventionChecker.ToPascalCase(structInfo.Name)),
			CodeSnippet: fmt.Sprintf("struct %s", structInfo.Name),
		})
	}

	// Check for acronym handling
	if d.hasImproperAcronymCasing(structInfo.Name) {
		violations = append(violations, &models.Violation{
			Type:        models.ViolationTypeNaming,
			Severity:    models.SeverityLow,
			Message:     fmt.Sprintf("Struct '%s' has improper acronym casing", structInfo.Name),
			File:        filePath,
			Line:        structInfo.StartLine,
			Column:      structInfo.StartColumn,
			Rule:        RustAcronymCasing,
			Suggestion:  d.getAcronymCasingSuggestion(structInfo.Name),
			CodeSnippet: fmt.Sprintf("struct %s", structInfo.Name),
		})
	}

	return violations
}

// checkEnumNaming analyzes Rust enum names for violations
func (d *RustNamingDetector) checkEnumNaming(enumInfo *types.RustEnumInfo, filePath string) []*models.Violation {
	var violations []*models.Violation

	// Check for non-descriptive names
	if d.isNonDescriptiveName(enumInfo.Name) {
		violations = append(violations, &models.Violation{
			Type:        models.ViolationTypeNaming,
			Severity:    models.SeverityMedium,
			Message:     fmt.Sprintf("Enum '%s' has a non-descriptive name", enumInfo.Name),
			File:        filePath,
			Line:        enumInfo.StartLine,
			Column:      enumInfo.StartColumn,
			Rule:        RustNonDescriptiveName,
			Suggestion:  "Choose a more descriptive type name that clearly indicates its purpose",
			CodeSnippet: fmt.Sprintf("enum %s", enumInfo.Name),
		})
	}

	// Check for proper Rust enum naming (PascalCase)
	if !d.conventionChecker.IsValidTypeName(enumInfo.Name) {
		violations = append(violations, &models.Violation{
			Type:        models.ViolationTypeNaming,
			Severity:    models.SeverityHigh,
			Message:     fmt.Sprintf("Enum '%s' does not follow Rust naming conventions (should be PascalCase)", enumInfo.Name),
			File:        filePath,
			Line:        enumInfo.StartLine,
			Column:      enumInfo.StartColumn,
			Rule:        RustInvalidEnumNaming,
			Suggestion:  fmt.Sprintf("Rename to '%s'", d.conventionChecker.ToPascalCase(enumInfo.Name)),
			CodeSnippet: fmt.Sprintf("enum %s", enumInfo.Name),
		})
	}

	return violations
}

// checkTraitNaming analyzes Rust trait names for violations
func (d *RustNamingDetector) checkTraitNaming(traitInfo *types.RustTraitInfo, filePath string) []*models.Violation {
	var violations []*models.Violation

	// Check for non-descriptive names
	if d.isNonDescriptiveName(traitInfo.Name) {
		violations = append(violations, &models.Violation{
			Type:        models.ViolationTypeNaming,
			Severity:    models.SeverityMedium,
			Message:     fmt.Sprintf("Trait '%s' has a non-descriptive name", traitInfo.Name),
			File:        filePath,
			Line:        traitInfo.StartLine,
			Column:      traitInfo.StartColumn,
			Rule:        RustNonDescriptiveName,
			Suggestion:  "Choose a more descriptive trait name that clearly indicates its purpose",
			CodeSnippet: fmt.Sprintf("trait %s", traitInfo.Name),
		})
	}

	// Check for proper Rust trait naming (PascalCase)
	if !d.conventionChecker.IsValidTypeName(traitInfo.Name) {
		violations = append(violations, &models.Violation{
			Type:        models.ViolationTypeNaming,
			Severity:    models.SeverityHigh,
			Message:     fmt.Sprintf("Trait '%s' does not follow Rust naming conventions (should be PascalCase)", traitInfo.Name),
			File:        filePath,
			Line:        traitInfo.StartLine,
			Column:      traitInfo.StartColumn,
			Rule:        RustInvalidTraitNaming,
			Suggestion:  fmt.Sprintf("Rename to '%s'", d.conventionChecker.ToPascalCase(traitInfo.Name)),
			CodeSnippet: fmt.Sprintf("trait %s", traitInfo.Name),
		})
	}

	return violations
}

// checkConstantNaming analyzes Rust constant names for violations
func (d *RustNamingDetector) checkConstantNaming(constant *types.RustConstantInfo, filePath string) []*models.Violation {
	var violations []*models.Violation

	// Check for proper Rust constant naming (SCREAMING_SNAKE_CASE)
	if !d.conventionChecker.IsValidConstantName(constant.Name) {
		violations = append(violations, &models.Violation{
			Type:        models.ViolationTypeNaming,
			Severity:    models.SeverityMedium,
			Message:     fmt.Sprintf("Constant '%s' does not follow Rust naming conventions (should be SCREAMING_SNAKE_CASE)", constant.Name),
			File:        filePath,
			Line:        constant.Line,
			Column:      constant.Column,
			Rule:        RustInvalidConstantNaming,
			Suggestion:  fmt.Sprintf("Rename to '%s'", d.conventionChecker.ToScreamingSnakeCase(constant.Name)),
			CodeSnippet: fmt.Sprintf("const %s", constant.Name),
		})
	}

	return violations
}

// checkModuleNaming analyzes Rust module names for violations
func (d *RustNamingDetector) checkModuleNaming(module *types.RustModuleInfo, filePath string) []*models.Violation {
	var violations []*models.Violation

	// Check for proper Rust module naming (snake_case)
	if !d.conventionChecker.IsValidModuleName(module.Name) {
		violations = append(violations, &models.Violation{
			Type:        models.ViolationTypeNaming,
			Severity:    models.SeverityMedium,
			Message:     fmt.Sprintf("Module '%s' does not follow Rust naming conventions (should be snake_case)", module.Name),
			File:        filePath,
			Line:        module.StartLine,
			Column:      module.StartColumn,
			Rule:        RustInvalidModuleNaming,
			Suggestion:  fmt.Sprintf("Rename to '%s'", d.conventionChecker.ToSnakeCase(module.Name)),
			CodeSnippet: fmt.Sprintf("mod %s", module.Name),
		})
	}

	return violations
}

// Helper methods

func (d *RustNamingDetector) isNonDescriptiveName(name string) bool {
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

func (d *RustNamingDetector) isSingleLetterVar(name string) bool {
	return len(name) == 1 && unicode.IsLetter(rune(name[0]))
}

func (d *RustNamingDetector) isAcceptableInRustContext(paramName string, fn *types.RustFunctionInfo) bool {
	lowerParamName := strings.ToLower(paramName)
	lowerFuncName := strings.ToLower(fn.Name)
	
	// Common acceptable short parameter names in Rust
	acceptableShortParams := map[string]bool{
		"i": true, "j": true, "k": true, // Loop counters
		"x": true, "y": true, "z": true, // Coordinates/math
		"n": true, "m": true,             // Counts/sizes
		"r": true, "w": true,             // Readers/writers (though Rust prefers reader/writer)
		"t": true,                        // Type parameter or time
		"f": true,                        // Function/closure parameter
		"s": true,                        // String/slice
		"v": true,                        // Value/vector
		"ok": true,                       // Result pattern
		"id": true,                       // Identifier
		"tx": true, "rx": true,           // Channel sender/receiver
	}
	
	if acceptableShortParams[lowerParamName] {
		return true
	}
	
	// In iterator/closure contexts, single letters are more acceptable
	if strings.Contains(lowerFuncName, "map") || 
	   strings.Contains(lowerFuncName, "filter") || 
	   strings.Contains(lowerFuncName, "fold") ||
	   strings.Contains(lowerFuncName, "iter") {
		return true
	}
	
	// Mathematical functions allow single letters
	mathPatterns := []string{
		"^(min|max|abs|sqrt|pow|exp|log|sin|cos|tan).*",
		"^(add|sub|mul|div|rem)$",
	}
	
	for _, pattern := range mathPatterns {
		if matched, _ := regexp.MatchString(pattern, lowerFuncName); matched {
			return true
		}
	}
	
	return false
}

func (d *RustNamingDetector) hasInappropriateAbbreviation(name string) bool {
	// Extract words from snake_case or camelCase
	words := d.extractWordsFromRustName(name)
	
	for _, word := range words {
		if d.isKnownAbbreviation(word) {
			return true
		}
	}
	
	return false
}

func (d *RustNamingDetector) extractWordsFromRustName(name string) []string {
	// Handle snake_case
	if strings.Contains(name, "_") {
		return strings.Split(strings.ToLower(name), "_")
	}
	
	// Handle PascalCase/camelCase
	var words []string
	var currentWord strings.Builder
	
	for i, char := range name {
		if i > 0 && unicode.IsUpper(char) && unicode.IsLower(rune(name[i-1])) {
			if currentWord.Len() > 0 {
				words = append(words, strings.ToLower(currentWord.String()))
				currentWord.Reset()
			}
		}
		currentWord.WriteRune(unicode.ToLower(char))
	}
	
	if currentWord.Len() > 0 {
		words = append(words, currentWord.String())
	}
	
	return words
}

func (d *RustNamingDetector) isKnownAbbreviation(word string) bool {
	knownAbbrevs := map[string]bool{
		// Common problematic abbreviations in Rust
		"req": true, "res": true, "resp": true, "cfg": true, "mgr": true,
		"calc": true, "comp": true, "proc": true, "num": true,
		"addr": true, "btn": true, "img": true, "val": true, "var": true,
		// Note: 'str' is NOT included here as it's a common Rust type/pattern
		// impl, mod, pub are Rust keywords and should not be considered abbreviations
	}
	
	// Special case: 'str' is acceptable in Rust context (common type)
	if word == "str" {
		return false
	}
	
	return knownAbbrevs[word]
}

func (d *RustNamingDetector) hasImproperAcronymCasing(name string) bool {
	// Check for improper acronym patterns in PascalCase names
	// In Rust, acronyms should be like "Http" not "HTTP" when not at the end
	
	// Pattern 1: Multiple consecutive uppercase letters not at the end
	// e.g., "HTTPServer" should be "HttpServer"
	for i := 0; i < len(name)-2; i++ {
		if unicode.IsUpper(rune(name[i])) && 
		   unicode.IsUpper(rune(name[i+1])) && 
		   i+2 < len(name) && unicode.IsUpper(rune(name[i+2])) {
			// Three or more consecutive uppercase letters
			return true
		}
		if unicode.IsUpper(rune(name[i])) && 
		   unicode.IsUpper(rune(name[i+1])) && 
		   i+2 < len(name) && unicode.IsLower(rune(name[i+2])) &&
		   i > 0 {
			// Two uppercase letters followed by lowercase (like "HTTPServer")
			// Exception: allow at the beginning (like "IOError")
			return true
		}
	}
	
	return false
}

func (d *RustNamingDetector) getAcronymCasingSuggestion(name string) string {
	// In Rust, acronyms in PascalCase should be like "HttpServer" not "HTTPServer"
	return "In Rust, acronyms should be capitalized as words (e.g., 'HttpServer' not 'HTTPServer', 'IoError' not 'IOError')"
}

func (d *RustNamingDetector) generateAbbreviationSuggestion(name string) string {
	if d.programmingTerms != nil {
		// Analyze the programming term
		analysis := d.programmingTerms.AnalyzeProgrammingTerm(name)
		
		// If we have specific suggestions from morphological analysis, use them
		if len(analysis.SuggestedFixes) > 0 {
			return fmt.Sprintf("Consider these improvements: %s", strings.Join(analysis.SuggestedFixes, "; "))
		}
	}
	
	// Extract word components and suggest expansions
	words := d.extractWordsFromRustName(name)
	var suggestions []string
	
	expansions := map[string]string{
		"req":  "request",
		"res":  "response",
		"resp": "response",
		"cfg":  "config",
		"mgr":  "manager",
		"calc": "calculate",
		"comp": "compute",
		"proc": "process",
		"str":  "string",
		"num":  "number",
		"addr": "address",
		"btn":  "button",
		"img":  "image",
		"val":  "value",
		"var":  "variable",
	}
	
	for _, word := range words {
		if expansion, ok := expansions[word]; ok {
			suggestions = append(suggestions, fmt.Sprintf("'%s' → '%s'", word, expansion))
		}
	}
	
	if len(suggestions) > 0 {
		return fmt.Sprintf("Consider expanding abbreviations: %s", strings.Join(suggestions, ", "))
	}
	
	return "Consider spelling out abbreviations for better readability"
}

// Code snippet generation helpers

func (d *RustNamingDetector) generateFunctionSnippet(fn *types.RustFunctionInfo) string {
	var modifiers []string
	if fn.IsPublic {
		modifiers = append(modifiers, "pub")
	}
	if fn.IsAsync {
		modifiers = append(modifiers, "async")
	}
	if fn.IsUnsafe {
		modifiers = append(modifiers, "unsafe")
	}
	
	prefix := ""
	if len(modifiers) > 0 {
		prefix = strings.Join(modifiers, " ") + " "
	}
	
	return fmt.Sprintf("%sfn %s", prefix, fn.Name)
}

func (d *RustNamingDetector) generateParameterSnippet(fn *types.RustFunctionInfo, param types.RustParameterInfo) string {
	paramStr := param.Name
	if param.IsMutable {
		paramStr = "mut " + paramStr
	}
	if param.IsRef {
		paramStr = "&" + paramStr
	}
	if param.Type != "" {
		paramStr = fmt.Sprintf("%s: %s", paramStr, param.Type)
	}
	return paramStr
}