package models

import (
	"strings"
	"testing"
)

func TestGetRustViolationCategory(t *testing.T) {
	tests := []struct {
		name         string
		violationType ViolationType
		expected     RustViolationCategory
	}{
		// Naming violations
		{
			name:         "Function naming violation",
			violationType: ViolationTypeRustInvalidFunctionNaming,
			expected:     RustCategoryNaming,
		},
		{
			name:         "Struct naming violation",
			violationType: ViolationTypeRustInvalidStructNaming,
			expected:     RustCategoryNaming,
		},
		
		// Safety violations
		{
			name:         "Unnecessary unsafe",
			violationType: ViolationTypeRustUnnecessaryUnsafe,
			expected:     RustCategorySafety,
		},
		{
			name:         "Transmute abuse",
			violationType: ViolationTypeRustTransmuteAbuse,
			expected:     RustCategorySafety,
		},
		
		// Ownership violations
		{
			name:         "Unnecessary clone",
			violationType: ViolationTypeRustUnnecessaryClone,
			expected:     RustCategoryOwnership,
		},
		{
			name:         "Complex lifetime",
			violationType: ViolationTypeRustComplexLifetime,
			expected:     RustCategoryOwnership,
		},
		
		// Performance violations
		{
			name:         "Inefficient string concat",
			violationType: ViolationTypeRustInefficientStringConcat,
			expected:     RustCategoryPerformance,
		},
		{
			name:         "Blocking in async",
			violationType: ViolationTypeRustBlockingInAsync,
			expected:     RustCategoryPerformance,
		},
		
		// Error handling violations
		{
			name:         "Overuse unwrap",
			violationType: ViolationTypeRustOveruseUnwrap,
			expected:     RustCategoryErrorHandling,
		},
		{
			name:         "Missing error propagation",
			violationType: ViolationTypeRustMissingErrorPropagation,
			expected:     RustCategoryErrorHandling,
		},
		
		// Pattern matching violations
		{
			name:         "Non-exhaustive match",
			violationType: ViolationTypeRustNonExhaustiveMatch,
			expected:     RustCategoryPatternMatching,
		},
		{
			name:         "Nested pattern matching",
			violationType: ViolationTypeRustNestedPatternMatching,
			expected:     RustCategoryPatternMatching,
		},
		
		// Trait violations
		{
			name:         "Overly complex trait",
			violationType: ViolationTypeRustOverlyComplexTrait,
			expected:     RustCategoryTraits,
		},
		{
			name:         "Trait bound complexity",
			violationType: ViolationTypeRustTraitBoundComplexity,
			expected:     RustCategoryTraits,
		},
		
		// Macro violations
		{
			name:         "Macro complexity",
			violationType: ViolationTypeRustMacroComplexity,
			expected:     RustCategoryMacros,
		},
		{
			name:         "Macro hygiene",
			violationType: ViolationTypeRustMacroHygiene,
			expected:     RustCategoryMacros,
		},
		
		// Async violations
		{
			name:         "Async fn in trait",
			violationType: ViolationTypeRustAsyncFnInTrait,
			expected:     RustCategoryAsync,
		},
		{
			name:         "Deadlock prone",
			violationType: ViolationTypeRustDeadlockProne,
			expected:     RustCategoryAsync,
		},
		
		// Module violations
		{
			name:         "Improper visibility",
			violationType: ViolationTypeRustImproperVisibility,
			expected:     RustCategoryModules,
		},
		{
			name:         "Circular dependency",
			violationType: ViolationTypeRustCircularDependency,
			expected:     RustCategoryModules,
		},
		
		// Non-Rust violation
		{
			name:         "General function length",
			violationType: ViolationTypeFunctionLength,
			expected:     "",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetRustViolationCategory(tt.violationType)
			if result != tt.expected {
				t.Errorf("GetRustViolationCategory(%v) = %v, want %v", tt.violationType, result, tt.expected)
			}
		})
	}
}

func TestIsRustSpecificViolation(t *testing.T) {
	tests := []struct {
		name         string
		violationType ViolationType
		expected     bool
	}{
		{
			name:         "Rust function naming violation",
			violationType: ViolationTypeRustInvalidFunctionNaming,
			expected:     true,
		},
		{
			name:         "Rust unsafe violation",
			violationType: ViolationTypeRustUnnecessaryUnsafe,
			expected:     true,
		},
		{
			name:         "General function length violation",
			violationType: ViolationTypeFunctionLength,
			expected:     false,
		},
		{
			name:         "General naming violation",
			violationType: ViolationTypeNaming,
			expected:     false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRustSpecificViolation(tt.violationType)
			if result != tt.expected {
				t.Errorf("IsRustSpecificViolation(%v) = %v, want %v", tt.violationType, result, tt.expected)
			}
		})
	}
}

func TestGetRustViolationDescription(t *testing.T) {
	tests := []struct {
		name         string
		violationType ViolationType
		expectedContains string
	}{
		{
			name:         "Function naming",
			violationType: ViolationTypeRustInvalidFunctionNaming,
			expectedContains: "snake_case",
		},
		{
			name:         "Struct naming",
			violationType: ViolationTypeRustInvalidStructNaming,
			expectedContains: "PascalCase",
		},
		{
			name:         "Constant naming",
			violationType: ViolationTypeRustInvalidConstantNaming,
			expectedContains: "SCREAMING_SNAKE_CASE",
		},
		{
			name:         "Unnecessary unsafe",
			violationType: ViolationTypeRustUnnecessaryUnsafe,
			expectedContains: "unsafe",
		},
		{
			name:         "Unnecessary clone",
			violationType: ViolationTypeRustUnnecessaryClone,
			expectedContains: "clone",
		},
		{
			name:         "Overuse unwrap",
			violationType: ViolationTypeRustOveruseUnwrap,
			expectedContains: "unwrap",
		},
		{
			name:         "Unknown violation",
			violationType: ViolationType("unknown_rust_violation"),
			expectedContains: "Unknown Rust violation",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetRustViolationDescription(tt.violationType)
			if result == "" {
				t.Errorf("GetRustViolationDescription(%v) returned empty string", tt.violationType)
			}
			if len(result) < 10 {
				t.Errorf("GetRustViolationDescription(%v) returned suspiciously short description: %q", tt.violationType, result)
			}
			if !strings.Contains(result, tt.expectedContains) {
				t.Errorf("GetRustViolationDescription(%v) returned %q, expected it to contain %q", tt.violationType, result, tt.expectedContains)
			}
		})
	}
}

func TestGetRustViolationSuggestion(t *testing.T) {
	tests := []struct {
		name         string
		violationType ViolationType
		expectedContains string
	}{
		{
			name:         "Function naming suggestion",
			violationType: ViolationTypeRustInvalidFunctionNaming,
			expectedContains: "snake_case",
		},
		{
			name:         "Struct naming suggestion",
			violationType: ViolationTypeRustInvalidStructNaming,
			expectedContains: "PascalCase",
		},
		{
			name:         "Unnecessary clone suggestion",
			violationType: ViolationTypeRustUnnecessaryClone,
			expectedContains: "borrowing",
		},
		{
			name:         "Overuse unwrap suggestion",
			violationType: ViolationTypeRustOveruseUnwrap,
			expectedContains: "error handling",
		},
		{
			name:         "Blocking in async suggestion",
			violationType: ViolationTypeRustBlockingInAsync,
			expectedContains: "async",
		},
		{
			name:         "Unknown violation suggestion",
			violationType: ViolationType("unknown_rust_violation"),
			expectedContains: "Rust documentation",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetRustViolationSuggestion(tt.violationType)
			if result == "" {
				t.Errorf("GetRustViolationSuggestion(%v) returned empty string", tt.violationType)
			}
			if len(result) < 10 {
				t.Errorf("GetRustViolationSuggestion(%v) returned suspiciously short suggestion: %q", tt.violationType, result)
			}
			if !strings.Contains(result, tt.expectedContains) {
				t.Errorf("GetRustViolationSuggestion(%v) returned %q, expected it to contain %q", tt.violationType, result, tt.expectedContains)
			}
		})
	}
}

func TestGetDefaultRustViolationSeverity(t *testing.T) {
	tests := []struct {
		name         string
		violationType ViolationType
		expected     Severity
	}{
		// High severity violations
		{
			name:         "Transmute abuse",
			violationType: ViolationTypeRustTransmuteAbuse,
			expected:     SeverityHigh,
		},
		{
			name:         "Panic prone code",
			violationType: ViolationTypeRustPanicProneCode,
			expected:     SeverityHigh,
		},
		{
			name:         "Race condition",
			violationType: ViolationTypeRustRaceCondition,
			expected:     SeverityHigh,
		},
		
		// Medium severity violations
		{
			name:         "Unnecessary unsafe",
			violationType: ViolationTypeRustUnnecessaryUnsafe,
			expected:     SeverityMedium,
		},
		{
			name:         "Complex lifetime",
			violationType: ViolationTypeRustComplexLifetime,
			expected:     SeverityMedium,
		},
		{
			name:         "Overuse unwrap",
			violationType: ViolationTypeRustOveruseUnwrap,
			expected:     SeverityMedium,
		},
		
		// Low severity violations
		{
			name:         "Function naming",
			violationType: ViolationTypeRustInvalidFunctionNaming,
			expected:     SeverityLow,
		},
		{
			name:         "Unnecessary clone",
			violationType: ViolationTypeRustUnnecessaryClone,
			expected:     SeverityLow,
		},
		{
			name:         "Unused import",
			violationType: ViolationTypeRustUnusedImport,
			expected:     SeverityLow,
		},
		
		// Unknown violation (defaults to medium)
		{
			name:         "Unknown violation",
			violationType: ViolationType("unknown_rust_violation"),
			expected:     SeverityMedium,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetDefaultRustViolationSeverity(tt.violationType)
			if result != tt.expected {
				t.Errorf("GetDefaultRustViolationSeverity(%v) = %v, want %v", tt.violationType, result, tt.expected)
			}
		})
	}
}

func TestAllRustViolationTypesHaveCategory(t *testing.T) {
	// Test that all Rust-specific violation types have a valid category
	rustViolationTypes := []ViolationType{
		// Naming violations
		ViolationTypeRustInvalidFunctionNaming,
		ViolationTypeRustInvalidStructNaming,
		ViolationTypeRustInvalidEnumNaming,
		ViolationTypeRustInvalidTraitNaming,
		ViolationTypeRustInvalidConstantNaming,
		ViolationTypeRustInvalidModuleNaming,
		ViolationTypeRustInvalidVariableNaming,
		
		// Safety violations
		ViolationTypeRustUnnecessaryUnsafe,
		ViolationTypeRustUnsafeWithoutComment,
		ViolationTypeRustTransmuteAbuse,
		ViolationTypeRustRawPointerAbuse,
		
		// Ownership violations
		ViolationTypeRustUnnecessaryClone,
		ViolationTypeRustInefficientBorrowing,
		ViolationTypeRustComplexLifetime,
		ViolationTypeRustMoveSemanticsViolation,
		ViolationTypeRustBorrowCheckerBypass,
		
		// Performance violations
		ViolationTypeRustInefficientStringConcat,
		ViolationTypeRustUnnecessaryAllocation,
		ViolationTypeRustBlockingInAsync,
		ViolationTypeRustInefficientIteration,
		ViolationTypeRustUnnecessaryCollection,
		
		// Error handling violations
		ViolationTypeRustOveruseUnwrap,
		ViolationTypeRustMissingErrorPropagation,
		ViolationTypeRustInconsistentErrorType,
		ViolationTypeRustPanicProneCode,
		ViolationTypeRustUnhandledResult,
		ViolationTypeRustImproperExpect,
		
		// Pattern matching violations
		ViolationTypeRustNonExhaustiveMatch,
		ViolationTypeRustNestedPatternMatching,
		ViolationTypeRustInefficientDestructuring,
		ViolationTypeRustUnreachablePattern,
		ViolationTypeRustMissingMatchArm,
		
		// Trait violations
		ViolationTypeRustOverlyComplexTrait,
		ViolationTypeRustMissingTraitImpl,
		ViolationTypeRustTraitBoundComplexity,
		ViolationTypeRustOrphanRule,
		
		// Macro violations
		ViolationTypeRustMacroComplexity,
		ViolationTypeRustMacroHygiene,
		ViolationTypeRustProceduralMacroMisuse,
		
		// Async violations
		ViolationTypeRustAsyncFnInTrait,
		ViolationTypeRustSendSyncViolation,
		ViolationTypeRustDeadlockProne,
		ViolationTypeRustRaceCondition,
		
		// Module violations
		ViolationTypeRustImproperVisibility,
		ViolationTypeRustCircularDependency,
		ViolationTypeRustModuleOrganization,
		ViolationTypeRustUnusedImport,
	}
	
	for _, violationType := range rustViolationTypes {
		t.Run(string(violationType), func(t *testing.T) {
			category := GetRustViolationCategory(violationType)
			if category == "" {
				t.Errorf("Rust violation type %v has no category assigned", violationType)
			}
			
			// Also test that it's recognized as Rust-specific
			if !IsRustSpecificViolation(violationType) {
				t.Errorf("Rust violation type %v is not recognized as Rust-specific", violationType)
			}
			
			// Test that it has a description
			description := GetRustViolationDescription(violationType)
			if description == "" || description == "Unknown Rust violation" {
				t.Errorf("Rust violation type %v has no proper description", violationType)
			}
			
			// Test that it has a suggestion
			suggestion := GetRustViolationSuggestion(violationType)
			if suggestion == "" || suggestion == "Refer to Rust documentation and best practices" {
				t.Errorf("Rust violation type %v has no proper suggestion", violationType)
			}
			
			// Test that it has a severity
			severity := GetDefaultRustViolationSeverity(violationType)
			if severity < SeverityInfo || severity > SeverityCritical {
				t.Errorf("Rust violation type %v has invalid severity: %v", violationType, severity)
			}
		})
	}
}