package violations

import (
	"fmt"
	"strings"

	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/types"
)

// RustTraitDetector analyzes Rust traits for clean code violations
// Focuses on maintainability metrics that complement clippy analysis
type RustTraitDetector struct {
	config    *DetectorConfig
	violations []*models.Violation
}

// NewRustTraitDetector creates a new Rust trait detector
func NewRustTraitDetector(config *DetectorConfig) *RustTraitDetector {
	return &RustTraitDetector{
		config:     config,
		violations: make([]*models.Violation, 0),
	}
}

// Detect analyzes Rust AST for trait-related violations
func (d *RustTraitDetector) Detect(astInfo *types.RustASTInfo) []*models.Violation {
	d.violations = make([]*models.Violation, 0)

	if astInfo == nil {
		return d.violations
	}

	// Analyze traits for various violations
	for _, trait := range astInfo.Traits {
		d.analyzeTraitComplexity(trait, astInfo.FilePath)
		d.analyzeTraitNaming(trait, astInfo.FilePath)
		d.analyzeTraitSize(trait, astInfo.FilePath)
		d.analyzeAssociatedTypes(trait, astInfo.FilePath)
	}

	// Analyze trait bounds in implementations
	for _, impl := range astInfo.Impls {
		d.analyzeTraitBounds(impl, astInfo.FilePath)
	}

	// Analyze functions for trait bound complexity
	for _, function := range astInfo.Functions {
		d.analyzeFunctionTraitBounds(function, astInfo.FilePath)
	}

	return d.violations
}

// analyzeTraitComplexity checks for overly complex traits
func (d *RustTraitDetector) analyzeTraitComplexity(trait *types.RustTraitInfo, filePath string) {
	// Calculate trait complexity score based on multiple factors
	complexityScore := d.calculateTraitComplexity(trait)
	
	maxComplexity := d.getConfigValue("max_trait_complexity", 15)
	if complexityScore > maxComplexity {
		d.addViolation(
			models.ViolationTypeRustOverlyComplexTrait,
			fmt.Sprintf("Trait '%s' is overly complex (complexity: %d, max: %d)",
				trait.Name, complexityScore, maxComplexity),
			filePath,
			trait.StartLine,
			trait.StartColumn,
			models.SeverityMedium,
			fmt.Sprintf("Consider splitting trait '%s' into smaller, focused traits. Complex traits violate the Single Responsibility Principle.", trait.Name),
		)
	}
}

// calculateTraitComplexity computes a complexity score for a trait
func (d *RustTraitDetector) calculateTraitComplexity(trait *types.RustTraitInfo) int {
	score := 0
	
	// Base complexity from method count
	score += trait.MethodCount * 2
	
	// Add complexity for each line of the trait
	lineCount := trait.EndLine - trait.StartLine + 1
	if lineCount > 20 {
		score += (lineCount - 20) / 5 // 1 point per 5 lines over 20
	}
	
	// Additional complexity if the trait name suggests multiple responsibilities
	if d.hasMultipleResponsibilities(trait.Name) {
		score += 5
	}
	
	// More complex if many associated types (estimated from generic method patterns)
	if trait.MethodCount > 0 {
		// Estimate associated types complexity based on trait size
		estimatedAssociatedTypes := (trait.MethodCount + 2) / 3
		score += estimatedAssociatedTypes * 2
	}
	
	return score
}

// hasMultipleResponsibilities checks if trait name suggests multiple responsibilities
func (d *RustTraitDetector) hasMultipleResponsibilities(traitName string) bool {
	// Look for conjunction words or multiple concerns in trait names
	multiResponsibilityIndicators := []string{
		"And", "Plus", "With", "Manager", "Handler", "Processor", "Controller",
		"Utility", "Helper", "Mixed", "Combined", "Multi",
	}
	
	lowerName := strings.ToLower(traitName)
	for _, indicator := range multiResponsibilityIndicators {
		if strings.Contains(lowerName, strings.ToLower(indicator)) {
			return true
		}
	}
	
	// Check for CamelCase patterns that might indicate multiple words
	upperCount := 0
	for _, char := range traitName {
		if char >= 'A' && char <= 'Z' {
			upperCount++
		}
	}
	
	// If more than 3 capital letters, might indicate multiple concepts
	return upperCount > 3
}

// analyzeTraitNaming checks trait naming conventions
func (d *RustTraitDetector) analyzeTraitNaming(trait *types.RustTraitInfo, filePath string) {
	if !d.isValidRustTraitName(trait.Name) {
		d.addViolation(
			models.ViolationTypeRustInvalidTraitNaming,
			fmt.Sprintf("Trait '%s' does not follow Rust PascalCase naming convention", trait.Name),
			filePath,
			trait.StartLine,
			trait.StartColumn,
			models.SeverityLow,
			fmt.Sprintf("Rename trait to use PascalCase convention (e.g., %s)", d.suggestTraitName(trait.Name)),
		)
	}
	
	// Check for descriptive naming
	if d.isNonDescriptiveName(trait.Name) {
		d.addViolation(
			models.ViolationTypeRustInvalidTraitNaming,
			fmt.Sprintf("Trait name '%s' is not descriptive enough", trait.Name),
			filePath,
			trait.StartLine,
			trait.StartColumn,
			models.SeverityLow,
			"Use more descriptive trait names that clearly indicate the behavior or capability they provide",
		)
	}
}

// isValidRustTraitName checks if trait name follows Rust conventions
func (d *RustTraitDetector) isValidRustTraitName(name string) bool {
	if len(name) == 0 {
		return false
	}
	
	// Must start with uppercase letter
	if name[0] < 'A' || name[0] > 'Z' {
		return false
	}
	
	// Check for PascalCase pattern
	for i, char := range name {
		if i == 0 {
			continue // Already checked first character
		}
		
		// Allow alphanumeric characters
		if (char >= 'a' && char <= 'z') || 
		   (char >= 'A' && char <= 'Z') || 
		   (char >= '0' && char <= '9') {
			continue
		}
		
		// Disallow underscores, hyphens, etc.
		return false
	}
	
	return true
}

// suggestTraitName suggests a corrected trait name
func (d *RustTraitDetector) suggestTraitName(name string) string {
	// Convert to PascalCase
	parts := strings.FieldsFunc(name, func(c rune) bool {
		return c == '_' || c == '-' || c == ' '
	})
	
	var result strings.Builder
	for _, part := range parts {
		if len(part) > 0 {
			result.WriteString(strings.ToUpper(string(part[0])))
			if len(part) > 1 {
				result.WriteString(strings.ToLower(part[1:]))
			}
		}
	}
	
	suggestion := result.String()
	if suggestion == "" {
		return "MyTrait" // Fallback
	}
	return suggestion
}

// isNonDescriptiveName checks if trait name is too generic
func (d *RustTraitDetector) isNonDescriptiveName(name string) bool {
	nonDescriptiveNames := []string{
		"T", "Trait", "Interface", "Base", "Common", "Util", "Helper",
		"Manager", "Handler", "Processor", "Thing", "Item", "Object",
		"Data", "Info", "Type", "Value",
	}
	
	lowerName := strings.ToLower(name)
	for _, badName := range nonDescriptiveNames {
		if lowerName == strings.ToLower(badName) {
			return true
		}
	}
	
	// Names that are too short (except for well-known single letters)
	if len(name) <= 2 && !d.isAcceptableShortName(name) {
		return true
	}
	
	return false
}

// isAcceptableShortName checks if short trait name is acceptable
func (d *RustTraitDetector) isAcceptableShortName(name string) bool {
	acceptableShort := []string{
		"Eq", "Ord", "IO", "UI", "DB", "OS", "CPU", "GPU", "API", "URL", "URI",
	}
	
	for _, acceptable := range acceptableShort {
		if name == acceptable {
			return true
		}
	}
	
	return false
}

// analyzeTraitSize checks for traits that are too large
func (d *RustTraitDetector) analyzeTraitSize(trait *types.RustTraitInfo, filePath string) {
	lineCount := trait.EndLine - trait.StartLine + 1
	maxLines := d.getConfigValue("max_trait_lines", 50)
	
	if lineCount > maxLines {
		d.addViolation(
			models.ViolationTypeRustOverlyComplexTrait,
			fmt.Sprintf("Trait '%s' is too large (%d lines, max: %d)",
				trait.Name, lineCount, maxLines),
			filePath,
			trait.StartLine,
			trait.StartColumn,
			models.SeverityMedium,
			fmt.Sprintf("Consider splitting large trait '%s' into smaller, focused traits", trait.Name),
		)
	}
	
	// Check method count
	maxMethods := d.getConfigValue("max_trait_methods", 8)
	if trait.MethodCount > maxMethods {
		d.addViolation(
			models.ViolationTypeRustOverlyComplexTrait,
			fmt.Sprintf("Trait '%s' has too many methods (%d methods, max: %d)",
				trait.Name, trait.MethodCount, maxMethods),
			filePath,
			trait.StartLine,
			trait.StartColumn,
			models.SeverityMedium,
			fmt.Sprintf("Consider splitting trait '%s' into smaller traits with fewer methods", trait.Name),
		)
	}
}

// analyzeAssociatedTypes checks for excessive associated types (estimated)
func (d *RustTraitDetector) analyzeAssociatedTypes(trait *types.RustTraitInfo, filePath string) {
	// Since we don't have direct access to associated types in the current AST structure,
	// we'll estimate based on trait complexity and method patterns
	estimatedAssociatedTypes := d.estimateAssociatedTypes(trait)
	
	maxAssociatedTypes := d.getConfigValue("max_associated_types", 4)
	if estimatedAssociatedTypes > maxAssociatedTypes {
		d.addViolation(
			models.ViolationTypeRustOverlyComplexTrait,
			fmt.Sprintf("Trait '%s' likely has too many associated types (estimated: %d, max: %d)",
				trait.Name, estimatedAssociatedTypes, maxAssociatedTypes),
			filePath,
			trait.StartLine,
			trait.StartColumn,
			models.SeverityMedium,
			fmt.Sprintf("Consider reducing associated types in trait '%s' or splitting into multiple traits", trait.Name),
		)
	}
}

// estimateAssociatedTypes estimates the number of associated types in a trait
func (d *RustTraitDetector) estimateAssociatedTypes(trait *types.RustTraitInfo) int {
	// Heuristic: Complex traits with many methods likely have more associated types
	// This is an approximation since we don't have actual associated type info
	if trait.MethodCount <= 2 {
		return 0
	} else if trait.MethodCount <= 5 {
		return 1
	} else if trait.MethodCount <= 8 {
		return 2
	} else {
		return (trait.MethodCount / 3) // Rough estimate
	}
}

// analyzeTraitBounds checks for complex trait bounds in implementations
func (d *RustTraitDetector) analyzeTraitBounds(impl *types.RustImplInfo, filePath string) {
	if impl.TraitName == "" {
		return // Skip inherent impls
	}
	
	// Check if implementation looks complex based on method count
	if impl.MethodCount > 15 {
		d.addViolation(
			models.ViolationTypeRustOverlyComplexTrait,
			fmt.Sprintf("Implementation of trait '%s' for '%s' is very large (%d methods)",
				impl.TraitName, impl.TargetType, impl.MethodCount),
			filePath,
			impl.StartLine,
			impl.StartColumn,
			models.SeverityMedium,
			"Consider splitting large trait implementations into smaller, focused traits",
		)
	}
	
	// Check for potentially complex trait bounds (heuristic based on naming)
	if d.hasComplexTraitBounds(impl.TraitName, impl.TargetType) {
		d.addViolation(
			models.ViolationTypeRustTraitBoundComplexity,
			fmt.Sprintf("Implementation may have complex trait bounds: %s for %s",
				impl.TraitName, impl.TargetType),
			filePath,
			impl.StartLine,
			impl.StartColumn,
			models.SeverityMedium,
			"Simplify trait bounds using where clauses for better readability",
		)
	}
}

// hasComplexTraitBounds uses heuristics to detect potentially complex trait bounds
func (d *RustTraitDetector) hasComplexTraitBounds(traitName, targetType string) bool {
	// Look for patterns that suggest complex generics
	complexPatterns := []string{
		"<", ">", "where", "Clone", "Send", "Sync", "Display", "Debug",
		"Iterator", "IntoIterator", "Future", "Stream",
	}
	
	combined := traitName + " " + targetType
	lowerCombined := strings.ToLower(combined)
	
	patternCount := 0
	for _, pattern := range complexPatterns {
		if strings.Contains(lowerCombined, strings.ToLower(pattern)) {
			patternCount++
		}
	}
	
	// If we find multiple complexity indicators, flag it
	return patternCount >= 3
}

// analyzeFunctionTraitBounds checks function signatures for complex trait bounds
func (d *RustTraitDetector) analyzeFunctionTraitBounds(function *types.RustFunctionInfo, filePath string) {
	// Check return type for complexity
	if d.hasComplexTraitBoundsInType(function.ReturnType) {
		d.addViolation(
			models.ViolationTypeRustTraitBoundComplexity,
			fmt.Sprintf("Function '%s' has complex trait bounds in return type",
				function.Name),
			filePath,
			function.StartLine,
			function.StartColumn,
			models.SeverityMedium,
			"Consider using type aliases or where clauses to simplify complex trait bounds",
		)
	}
	
	// Check parameters for complex trait bounds
	complexParamCount := 0
	for _, param := range function.Parameters {
		if d.hasComplexTraitBoundsInType(param.Type) {
			complexParamCount++
		}
	}
	
	maxComplexParams := d.getConfigValue("max_complex_trait_params", 2)
	if complexParamCount > maxComplexParams {
		d.addViolation(
			models.ViolationTypeRustTraitBoundComplexity,
			fmt.Sprintf("Function '%s' has too many parameters with complex trait bounds (%d)",
				function.Name, complexParamCount),
			filePath,
			function.StartLine,
			function.StartColumn,
			models.SeverityMedium,
			"Simplify function signature by using type aliases or moving complex bounds to where clauses",
		)
	}
}

// hasComplexTraitBoundsInType checks if a type string contains complex trait bounds
func (d *RustTraitDetector) hasComplexTraitBoundsInType(typeStr string) bool {
	if typeStr == "" {
		return false
	}
	
	// Count complexity indicators
	indicators := []string{"+", "where", "impl", "<", ">", "dyn"}
	count := 0
	
	lowerType := strings.ToLower(typeStr)
	for _, indicator := range indicators {
		count += strings.Count(lowerType, indicator)
	}
	
	// Also check for long type names which often indicate complexity
	if len(typeStr) > 50 {
		count += 2
	}
	
	return count >= 3
}

// addViolation adds a violation to the detector's list
func (d *RustTraitDetector) addViolation(
	violationType models.ViolationType,
	message string,
	filePath string,
	line int,
	column int,
	severity models.Severity,
	suggestion string,
) {
	violation := &models.Violation{
		Type:        violationType,
		Severity:    severity,
		Message:     message,
		File:        filePath,
		Line:        line,
		Column:      column,
		Suggestion:  suggestion,
		Description: models.GetRustViolationDescription(violationType),
	}
	
	d.violations = append(d.violations, violation)
}

// getConfigValue gets a configuration value with fallback to default
func (d *RustTraitDetector) getConfigValue(key string, defaultValue int) int {
	if d.config == nil || d.config.RustConfig == nil {
		return defaultValue
	}
	
	// Map configuration keys to RustConfig fields
	switch key {
	case "max_trait_complexity":
		if d.config.RustConfig.MaxTraitComplexity > 0 {
			return d.config.RustConfig.MaxTraitComplexity
		}
	case "max_trait_lines":
		if d.config.RustConfig.MaxTraitLines > 0 {
			return d.config.RustConfig.MaxTraitLines
		}
	case "max_trait_methods":
		if d.config.RustConfig.MaxTraitMethods > 0 {
			return d.config.RustConfig.MaxTraitMethods
		}
	case "max_associated_types":
		if d.config.RustConfig.MaxAssociatedTypes > 0 {
			return d.config.RustConfig.MaxAssociatedTypes
		}
	case "max_complex_trait_params":
		if d.config.RustConfig.MaxComplexTraitParams > 0 {
			return d.config.RustConfig.MaxComplexTraitParams
		}
	}
	
	return defaultValue
}

// Name returns the detector name
func (d *RustTraitDetector) Name() string {
	return "RustTraitDetector"
}

// Description returns the detector description
func (d *RustTraitDetector) Description() string {
	return "Analyzes Rust traits for maintainability issues including complexity, size, naming conventions, and trait bound usage that complement clippy analysis"
}