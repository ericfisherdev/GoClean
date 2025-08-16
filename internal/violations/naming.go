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
func (d *NamingDetector) Detect(fileInfo *models.FileInfo, astInfo interface{}) []*models.Violation {
	var violations []*models.Violation

	if astInfo == nil {
		return violations
	}
	
	// Type assertion to get scanner.GoASTInfo
	goAstInfo, ok := astInfo.(*types.GoASTInfo)
	if !ok || goAstInfo == nil {
		// For non-Go files, we have limited naming analysis
		return violations
	}

	// Check function names
	if goAstInfo.Functions != nil {
		for _, function := range goAstInfo.Functions {
			if function != nil {
				violations = append(violations, d.checkFunctionNaming(function, fileInfo.Path)...)
				// Check parameter names
				violations = append(violations, d.checkParameterNaming(function, fileInfo.Path)...)
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

// checkFunctionNaming analyzes function names for violations
func (d *NamingDetector) checkFunctionNaming(fn *types.FunctionInfo, filePath string) []*models.Violation {
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
func (d *NamingDetector) checkParameterNaming(fn *types.FunctionInfo, filePath string) []*models.Violation {
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

		// Check for non-descriptive parameter names with context consideration
		if d.isNonDescriptiveName(param.Name) && !d.isCommonShortParam(param.Name) && !d.isAcceptableInContext(param.Name, fn) {
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
func (d *NamingDetector) checkTypeNaming(typeInfo *types.TypeInfo, filePath string) []*models.Violation {
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
func (d *NamingDetector) checkVariableNaming(variable *types.VariableInfo, filePath string) []*models.Violation {
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
func (d *NamingDetector) checkConstantNaming(constant *types.ConstantInfo, filePath string) []*models.Violation {
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

// isAcceptableInContext determines if a parameter name is acceptable given the function context
func (d *NamingDetector) isAcceptableInContext(paramName string, fn *types.FunctionInfo) bool {
	lowerParamName := strings.ToLower(paramName)
	lowerFuncName := strings.ToLower(fn.Name)
	
	// Context 1: Mathematical/Comparison functions - single letters are often conventional
	mathFunctionPatterns := []string{
		"^(min|max|abs|sqrt|pow|exp|log|sin|cos|tan|ceil|floor|round)$",
		"^(compare|sort|swap|add|sub|mul|div|mod)$",
		"^(distance|angle|magnitude|normalize)$",
	}
	
	for _, pattern := range mathFunctionPatterns {
		if matched, _ := regexp.MatchString(pattern, lowerFuncName); matched {
			// In math functions, single letters and simple names are conventional
			if len(paramName) == 1 || d.isMathematicalParameter(lowerParamName) {
				return true
			}
		}
	}
	
	// Context 2: Generic data processing functions - 'data' might be acceptable
	dataProcessingPatterns := []string{
		"^(save|load|write|read|export|import|process|transform|convert).*",
		"^(serialize|deserialize|encode|decode|parse|format).*",
		"^(compress|decompress|encrypt|decrypt|hash|validate).*",
	}
	
	for _, pattern := range dataProcessingPatterns {
		if matched, _ := regexp.MatchString(pattern, lowerFuncName); matched {
			if d.isGenericDataParameter(lowerParamName, fn) {
				return true
			}
		}
	}
	
	// Context 3: Utility/Helper functions - context matters for acceptability
	utilityPatterns := []string{
		"^(get|set|is|has|can|should|will).*",
		"^(create|make|new|build|generate).*",
		"^(find|search|filter|map|reduce).*",
	}
	
	for _, pattern := range utilityPatterns {
		if matched, _ := regexp.MatchString(pattern, lowerFuncName); matched {
			if d.isUtilityParameter(lowerParamName, fn) {
				return true
			}
		}
	}
	
	// Context 4: Interface implementation methods - standard parameter names
	if d.isInterfaceMethod(fn) && d.isStandardInterfaceParameter(lowerParamName) {
		return true
	}
	
	return false
}

// isMathematicalParameter checks if a parameter name is acceptable in mathematical contexts
func (d *NamingDetector) isMathematicalParameter(paramName string) bool {
	mathParams := map[string]bool{
		"a": true, "b": true, "c": true,     // General math variables
		"x": true, "y": true, "z": true,     // Coordinates/unknowns  
		"n": true, "m": true, "k": true,     // Counts/indices
		"i": true, "j": true,                // Loop/matrix indices
		"v": true, "u": true, "w": true,     // Vectors
		"p": true, "q": true,                // Points/parameters
		"r": true, "theta": true,            // Polar coordinates
		"min": true, "max": true,            // Range values
		"val": true, "value": true,          // Generic values
	}
	return mathParams[paramName]
}

// isGenericDataParameter checks if 'data' or similar is acceptable based on function signature
func (d *NamingDetector) isGenericDataParameter(paramName string, fn *types.FunctionInfo) bool {
	genericDataNames := map[string]bool{
		"data": true, "content": true, "payload": true,
		"input": true, "output": true, "result": true,
		"bytes": true, "buffer": true, "stream": true,
	}
	
	if !genericDataNames[paramName] {
		return false
	}
	
	// If function has only 1-2 parameters and one is clearly the main data, it's acceptable
	if len(fn.Parameters) <= 2 {
		return true
	}
	
	// If the parameter type suggests it's the main data (slice, interface{}, etc.)
	for _, param := range fn.Parameters {
		if strings.EqualFold(param.Name, paramName) {
			paramType := strings.ToLower(param.Type)
			if strings.Contains(paramType, "[]") ||           // Slice types
			   strings.Contains(paramType, "interface") ||    // Generic interfaces
			   strings.Contains(paramType, "byte") ||         // Byte data
			   strings.Contains(paramType, "string") {        // String data
				return true
			}
		}
	}
	
	return false
}

// isUtilityParameter checks if a parameter name is acceptable in utility functions
func (d *NamingDetector) isUtilityParameter(paramName string, fn *types.FunctionInfo) bool {
	// For very short utility functions (1-3 lines), simple names might be acceptable
	// Note: We don't have line count here, so we use parameter count as a proxy
	if len(fn.Parameters) == 1 && (paramName == "v" || paramName == "val" || paramName == "item") {
		return true
	}
	
	// Standard utility parameter names
	utilityParams := map[string]bool{
		"ctx": true, "context": true,        // Context parameters
		"opts": true, "options": true,       // Options/config
		"cfg": true, "config": true,         // Configuration
		"src": true, "dst": true,            // Source/destination
		"key": true, "val": true,            // Key-value pairs
	}
	
	return utilityParams[paramName]
}

// isInterfaceMethod checks if this function is likely implementing an interface
func (d *NamingDetector) isInterfaceMethod(fn *types.FunctionInfo) bool {
	// Common interface method patterns
	interfaceMethodPatterns := []string{
		"^(String|Error|Read|Write|Close|Seek)$",
		"^(Marshal|Unmarshal|Encode|Decode).*",
		"^(Scan|Value)$", // database/sql interfaces
	}
	
	lowerName := strings.ToLower(fn.Name)
	for _, pattern := range interfaceMethodPatterns {
		if matched, _ := regexp.MatchString(strings.ToLower(pattern), lowerName); matched {
			return true
		}
	}
	
	return false
}

// isStandardInterfaceParameter checks for standard interface method parameters
func (d *NamingDetector) isStandardInterfaceParameter(paramName string) bool {
	standardParams := map[string]bool{
		"p": true,      // Write(p []byte)
		"b": true,      // Read(b []byte) 
		"v": true,      // Scan(v interface{})
		"data": true,   // Marshal/Unmarshal data
		"dst": true,    // encoding destinations
		"src": true,    // encoding sources
	}
	return standardParams[paramName]
}

func (d *NamingDetector) hasInappropriateAbbreviation(name string) bool {
	// Extract individual words from the camelCase name (preserve capitalization for splitting)
	words := d.extractWordsFromName(name)
	
	// Check each word to see if it's a known problematic abbreviation
	for _, word := range words {
		if d.isKnownAbbreviation(word) {
			return true
		}
	}
	
	return false
}

// AbbreviationInfo contains information about an abbreviation
type AbbreviationInfo struct {
	fullWords []string // Complete words this abbreviation represents
	minLength int      // Minimum length to consider as abbreviation
}

// containsCompleteWords checks if the name contains recognizable complete words using programmatic heuristics
func (d *NamingDetector) containsCompleteWords(name string) bool {
	// Split camelCase and extract potential words (preserve original capitalization for splitting)
	words := d.extractWordsFromName(name)
	
	// Require that at least one word is clearly complete AND
	// no words are clearly abbreviations for the name to be considered "complete"
	hasCompleteWord := false
	
	for _, word := range words {
		if d.isKnownAbbreviation(word) {
			return false // If any word is a known abbreviation, reject the whole name
		}
		
		if d.isLikelyCompleteWord(word) {
			hasCompleteWord = true
		}
	}
	
	return hasCompleteWord
}

// extractWordsFromName splits a function/variable name into component words
func (d *NamingDetector) extractWordsFromName(name string) []string {
	var words []string
	var currentWord strings.Builder
	
	for i, char := range name {
		if i > 0 && unicode.IsUpper(char) && unicode.IsLower(rune(name[i-1])) {
			// CamelCase boundary: lowercase to uppercase
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

// isLikelyCompleteWord uses multiple heuristics to determine if a word is complete (not an abbreviation)
func (d *NamingDetector) isLikelyCompleteWord(word string) bool {
	if len(word) <= 2 {
		return false // Very short words are likely abbreviations
	}
	
	// First check: Is this a known abbreviation? If so, reject immediately
	if d.isKnownAbbreviation(word) {
		return false
	}
	
	// Heuristic 1: Length-based analysis
	if len(word) >= 6 {
		return true // Words 6+ characters are usually complete
	}
	
	// For shorter words (3-5 chars), require multiple positive signals
	positiveSignals := 0
	
	// Heuristic 2: Vowel distribution analysis
	if d.hasGoodVowelDistribution(word) {
		positiveSignals++
	}
	
	// Heuristic 3: Common programming word patterns
	if d.matchesProgrammingWordPatterns(word) {
		positiveSignals++
	}
	
	// Heuristic 4: Morphological analysis (common suffixes/prefixes)
	if d.hasProgrammingMorphology(word) {
		positiveSignals++
	}
	
	// Heuristic 5: Known complete words (small essential list) - these override other rules
	if d.isKnownCompleteWord(word) {
		return true // Known complete words are always accepted
	}
	
	// For 3-5 character words, require at least 2 positive signals
	if len(word) <= 5 {
		return positiveSignals >= 2
	}
	
	// For 6+ character words, one positive signal is enough
	return positiveSignals >= 1
}

// isKnownAbbreviation checks if a word is a known problematic abbreviation
func (d *NamingDetector) isKnownAbbreviation(word string) bool {
	knownAbbrevs := map[string]bool{
		// Common abbreviations that should always be flagged
		"req": true, "res": true, "resp": true, "cfg": true, "mgr": true,
		"calc": true, "comp": true, "proc": true, "str": true, "num": true, "addr": true,
		"btn": true, "img": true, "mng": true, "mgmt": true, "strg": true,
		"nbr": true, "conf": true,
	}
	
	return knownAbbrevs[word]
}

// hasGoodVowelDistribution checks if the word has a natural vowel distribution
func (d *NamingDetector) hasGoodVowelDistribution(word string) bool {
	if len(word) < 3 {
		return false
	}
	
	vowels := 0
	consonants := 0
	
	for _, char := range word {
		if d.isVowel(char) {
			vowels++
		} else if unicode.IsLetter(char) {
			consonants++
		}
	}
	
	if vowels == 0 {
		return false // No vowels suggests abbreviation
	}
	
	// Healthy vowel ratio (typically 25-50% in English)
	vowelRatio := float64(vowels) / float64(vowels+consonants)
	return vowelRatio >= 0.20 && vowelRatio <= 0.60
}

// isVowel checks if a character is a vowel
func (d *NamingDetector) isVowel(char rune) bool {
	vowels := "aeiou"
	return strings.ContainsRune(vowels, unicode.ToLower(char))
}

// matchesProgrammingWordPatterns checks for common programming word patterns
func (d *NamingDetector) matchesProgrammingWordPatterns(word string) bool {
	// Common programming word endings that indicate complete words
	completeSuffixes := []string{
		"tion", "sion", "ment", "ness", "able", "ible", "ful", "ing",
		"ed", "er", "est", "ly", "ure", "age", "ous", "ive", "ity",
		"ate", "ize", "ise", "or", "ar", "ant", "ent",
	}
	
	for _, suffix := range completeSuffixes {
		if len(word) > len(suffix) && strings.HasSuffix(word, suffix) {
			return true
		}
	}
	
	// Common programming prefixes
	completePrefixes := []string{
		"pre", "post", "sub", "super", "inter", "trans", "over", "under",
		"out", "up", "down", "re", "un", "dis", "mis", "non",
	}
	
	for _, prefix := range completePrefixes {
		if len(word) > len(prefix)+2 && strings.HasPrefix(word, prefix) {
			return true
		}
	}
	
	return false
}

// hasProgrammingMorphology checks for programming-specific morphological patterns
func (d *NamingDetector) hasProgrammingMorphology(word string) bool {
	// Plural forms
	if len(word) > 3 && strings.HasSuffix(word, "s") {
		singular := word[:len(word)-1]
		if len(singular) >= 4 {
			return true // Plurals of reasonable length are likely complete
		}
	}
	
	// Past tense forms
	if len(word) > 4 && strings.HasSuffix(word, "ed") {
		return true
	}
	
	// Gerund forms
	if len(word) > 4 && strings.HasSuffix(word, "ing") {
		return true
	}
	
	// Comparative/superlative
	if len(word) > 4 && (strings.HasSuffix(word, "er") || strings.HasSuffix(word, "est")) {
		return true
	}
	
	return false
}

// isKnownCompleteWord checks against a minimal list of essential complete words
func (d *NamingDetector) isKnownCompleteWord(word string) bool {
	// Minimal essential list - only words that are commonly misidentified by heuristics
	essentialWords := map[string]bool{
		// Words that might not pass other heuristics but are definitely complete
		"staff": true, "key": true, "age": true, "bar": true, "row": true,
		"data": true, "item": true, "user": true, "file": true, "code": true,
		"line": true, "page": true, "node": true, "edge": true, "core": true,
		"mode": true, "type": true, "size": true, "path": true, "name": true,
		"time": true, "date": true, "week": true, "year": true, "hour": true,
		
		// Programming terms that are short but complete
		"hash": true, "json": true, "yaml": true, "xml": true, "http": true,
		"auth": true, "sync": true, "async": true, "exec": true, "init": true,
	}
	
	return essentialWords[word]
}

// containsOnlyCompleteWords checks if the name consists of only complete words without problematic abbreviations
func (d *NamingDetector) containsOnlyCompleteWords(lowerName string) bool {
	// Complete words that are safe and should never be flagged as abbreviations
	safeCompleteWords := []string{
		// Core complete words
		"generation", "generate", "generator", "complete", "completion",
		"features", "feature", "select", "selection", "selector", 
		"staff", "catering", "processing",
		"configuration", "management", "calculation", "computation",
		"structure", "procedure", "address", "number",
		"string", "stream", "storage", "button", "image",
		"computer", "component", "comparison", "manager", "conference",
		
		// Words containing common abbreviation patterns that should be safe
		"requires", "request", "requests", "required", "requirement",
		"response", "result", "results", "resource", "resources", 
		"responsible", "responsibility",
		"process", "processing", "processor",
	}
	
	// Check if the name consists only of safe complete words
	for _, word := range safeCompleteWords {
		if lowerName == word {
			return true // Exact match to a safe word
		}
		
		// Check if it's a compound of safe words (e.g., "selectFeatures", "requiresStaff")
		if strings.Contains(lowerName, word) {
			// Remove this word and check if the remainder is also safe
			remaining := strings.Replace(lowerName, word, "", 1)
			if remaining == "" {
				return true // Only this word
			}
			// Recursively check if the remaining part is also safe
			if d.containsOnlyCompleteWords(remaining) {
				return true
			}
		}
	}
	
	return false
}

// hasProblematicAbbreviation checks if a specific abbreviation is problematically used
func (d *NamingDetector) hasProblematicAbbreviation(lowerName, abbrev string, info AbbreviationInfo) bool {
	// First check if it's part of any full word from our known list
	for _, fullWord := range info.fullWords {
		if strings.Contains(lowerName, fullWord) {
			return false // It's part of a complete word, not an abbreviation
		}
	}
	
	// Check if the abbreviation appears in a problematic way
	// Pattern 1: Abbreviation at start followed by capital letter or end: reqData, req
	pattern1 := fmt.Sprintf("^%s([a-z]*|$)", abbrev)
	if matched, _ := regexp.MatchString(pattern1, lowerName); matched {
		// Found at start, check if it's part of a complete word
		return !d.isAbbrevPartOfCompleteWord(lowerName, abbrev, info)
	}
	
	// Pattern 2: Abbreviation at end: sendReq  
	pattern2 := fmt.Sprintf("([a-z]+)%s$", abbrev)
	if matched, _ := regexp.MatchString(pattern2, lowerName); matched {
		// Found at end, check if it's part of a complete word
		return !d.isAbbrevPartOfCompleteWord(lowerName, abbrev, info)
	}
	
	// Pattern 3: Abbreviation in middle with camelCase boundaries: handleReqResponse
	pattern3 := fmt.Sprintf("([a-z]+)%s([a-z]+)", abbrev)
	if matched, _ := regexp.MatchString(pattern3, lowerName); matched {
		// Found in middle, check if it's part of a complete word
		return !d.isAbbrevPartOfCompleteWord(lowerName, abbrev, info)
	}
	
	return false
}

// isAbbrevPartOfCompleteWord checks if the abbreviation is part of a complete word we missed
func (d *NamingDetector) isAbbrevPartOfCompleteWord(lowerName, abbrev string, info AbbreviationInfo) bool {
	// Find the position of the abbreviation
	abbrevIndex := strings.Index(lowerName, abbrev)
	if abbrevIndex == -1 {
		return false
	}
	
	// Check if there are additional letters that might make it a complete word
	endIndex := abbrevIndex + len(abbrev)
	
	// If there are more characters after the abbreviation, it might be a complete word
	if endIndex < len(lowerName) {
		// Extract the potential complete word
		remainingChars := lowerName[endIndex:]
		potentialWord := abbrev + remainingChars
		
		// Check if this potential word is in our known complete words list
		for _, fullWord := range info.fullWords {
			if strings.HasPrefix(fullWord, potentialWord) || potentialWord == fullWord {
				return true
			}
		}
	}
	
	// Check if there are characters before that make it a complete word
	if abbrevIndex > 0 {
		precedingChars := lowerName[:abbrevIndex]
		potentialWord := precedingChars + abbrev
		
		for _, fullWord := range info.fullWords {
			if strings.HasSuffix(fullWord, potentialWord) || potentialWord == fullWord {
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
		"AST", "GUI", "CLI", "SSH", "TLS", "SSL", "HTML", "CSS", "JWT",
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

func (d *NamingDetector) generateFunctionNameSnippet(fn *types.FunctionInfo) string {
	if fn.IsMethod && fn.ReceiverType != "" {
		return fmt.Sprintf("func (%s) %s", fn.ReceiverType, fn.Name)
	}
	return fmt.Sprintf("func %s", fn.Name)
}

func (d *NamingDetector) generateParameterSnippet(fn *types.FunctionInfo, paramName string) string {
	for _, param := range fn.Parameters {
		if param.Name == paramName {
			return fmt.Sprintf("%s %s", param.Name, param.Type)
		}
	}
	return paramName
}