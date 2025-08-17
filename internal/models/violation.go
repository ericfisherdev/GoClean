package models

// Severity represents the severity level of a violation
type Severity int

const (
	SeverityInfo Severity = iota
	SeverityLow
	SeverityMedium
	SeverityHigh
	SeverityCritical
)

// String returns the string representation of the Severity level.
func (s Severity) String() string {
	switch s {
	case SeverityInfo:
		return "Info"
	case SeverityLow:
		return "Low"
	case SeverityMedium:
		return "Medium"
	case SeverityHigh:
		return "High"
	case SeverityCritical:
		return "Critical"
	default:
		return "Unknown"
	}
}

// ViolationType represents the category of violation
type ViolationType string

const (
	// General violation types (language-agnostic)
	ViolationTypeFunctionLength         ViolationType = "function_length"
	ViolationTypeCyclomaticComplexity  ViolationType = "cyclomatic_complexity"
	ViolationTypeParameterCount        ViolationType = "parameter_count"
	ViolationTypeNestingDepth         ViolationType = "nesting_depth"
	ViolationTypeNaming               ViolationType = "naming_convention"
	ViolationTypeClassSize            ViolationType = "class_size"
	ViolationTypeMissingDocumentation ViolationType = "missing_documentation"
	ViolationTypeMagicNumbers         ViolationType = "magic_numbers"
	ViolationTypeDuplication          ViolationType = "code_duplication"
	ViolationTypeMagicNumber          ViolationType = "magic_number"
	ViolationTypeCommentedCode        ViolationType = "commented_code"
	ViolationTypeTodo                 ViolationType = "todo_marker"
	ViolationTypeDocumentation        ViolationType = "documentation_quality"
	ViolationTypeStructure            ViolationType = "code_structure"
	
	// Rust-specific naming violations
	ViolationTypeRustInvalidFunctionNaming ViolationType = "rust_invalid_function_naming"
	ViolationTypeRustInvalidStructNaming   ViolationType = "rust_invalid_struct_naming"
	ViolationTypeRustInvalidEnumNaming     ViolationType = "rust_invalid_enum_naming"
	ViolationTypeRustInvalidTraitNaming    ViolationType = "rust_invalid_trait_naming"
	ViolationTypeRustInvalidConstantNaming ViolationType = "rust_invalid_constant_naming"
	ViolationTypeRustInvalidModuleNaming   ViolationType = "rust_invalid_module_naming"
	ViolationTypeRustInvalidVariableNaming ViolationType = "rust_invalid_variable_naming"
	
	// Rust-specific safety violations
	ViolationTypeRustUnnecessaryUnsafe    ViolationType = "rust_unnecessary_unsafe"
	ViolationTypeRustUnsafeWithoutComment ViolationType = "rust_unsafe_without_comment"
	ViolationTypeRustTransmuteAbuse       ViolationType = "rust_transmute_abuse"
	ViolationTypeRustRawPointerAbuse      ViolationType = "rust_raw_pointer_abuse"
	
	// Rust-specific ownership violations
	ViolationTypeRustUnnecessaryClone        ViolationType = "rust_unnecessary_clone"
	ViolationTypeRustInefficientBorrowing    ViolationType = "rust_inefficient_borrowing"
	ViolationTypeRustComplexLifetime         ViolationType = "rust_complex_lifetime"
	ViolationTypeRustMoveSemanticsViolation  ViolationType = "rust_move_semantics_violation"
	ViolationTypeRustBorrowCheckerBypass     ViolationType = "rust_borrow_checker_bypass"
	
	// Rust-specific performance violations
	ViolationTypeRustInefficientStringConcat ViolationType = "rust_inefficient_string_concat"
	ViolationTypeRustUnnecessaryAllocation   ViolationType = "rust_unnecessary_allocation"
	ViolationTypeRustBlockingInAsync         ViolationType = "rust_blocking_in_async"
	ViolationTypeRustInefficientIteration    ViolationType = "rust_inefficient_iteration"
	ViolationTypeRustUnnecessaryCollection   ViolationType = "rust_unnecessary_collection"
	
	// Rust-specific error handling violations
	ViolationTypeRustOveruseUnwrap             ViolationType = "rust_overuse_unwrap"
	ViolationTypeRustMissingErrorPropagation   ViolationType = "rust_missing_error_propagation"
	ViolationTypeRustInconsistentErrorType     ViolationType = "rust_inconsistent_error_type"
	ViolationTypeRustPanicProneCode           ViolationType = "rust_panic_prone_code"
	ViolationTypeRustUnhandledResult          ViolationType = "rust_unhandled_result"
	ViolationTypeRustImproperExpect           ViolationType = "rust_improper_expect"
	
	// Rust-specific pattern matching violations
	ViolationTypeRustNonExhaustiveMatch      ViolationType = "rust_non_exhaustive_match"
	ViolationTypeRustNestedPatternMatching   ViolationType = "rust_nested_pattern_matching"
	ViolationTypeRustInefficientDestructuring ViolationType = "rust_inefficient_destructuring"
	ViolationTypeRustUnreachablePattern      ViolationType = "rust_unreachable_pattern"
	ViolationTypeRustMissingMatchArm         ViolationType = "rust_missing_match_arm"
	
	// Rust-specific trait and implementation violations
	ViolationTypeRustOverlyComplexTrait      ViolationType = "rust_overly_complex_trait"
	ViolationTypeRustMissingTraitImpl        ViolationType = "rust_missing_trait_impl"
	ViolationTypeRustTraitBoundComplexity    ViolationType = "rust_trait_bound_complexity"
	ViolationTypeRustOrphanRule              ViolationType = "rust_orphan_rule"
	
	// Rust-specific macro violations
	ViolationTypeRustMacroComplexity         ViolationType = "rust_macro_complexity"
	ViolationTypeRustMacroHygiene            ViolationType = "rust_macro_hygiene"
	ViolationTypeRustProceduralMacroMisuse   ViolationType = "rust_procedural_macro_misuse"
	
	// Rust-specific async/concurrency violations
	ViolationTypeRustAsyncFnInTrait          ViolationType = "rust_async_fn_in_trait"
	ViolationTypeRustSendSyncViolation       ViolationType = "rust_send_sync_violation"
	ViolationTypeRustDeadlockProne           ViolationType = "rust_deadlock_prone"
	ViolationTypeRustRaceCondition           ViolationType = "rust_race_condition"
	
	// Rust-specific module and visibility violations
	ViolationTypeRustImproperVisibility      ViolationType = "rust_improper_visibility"
	ViolationTypeRustCircularDependency      ViolationType = "rust_circular_dependency"
	ViolationTypeRustModuleOrganization      ViolationType = "rust_module_organization"
	ViolationTypeRustUnusedImport            ViolationType = "rust_unused_import"
)

// Violation represents a clean code violation found during scanning
type Violation struct {
	ID          string        `json:"id"`
	Type        ViolationType `json:"type"`
	Severity    Severity      `json:"severity"`
	Message     string        `json:"message"`
	Description string        `json:"description"`
	File        string        `json:"file"`
	Line        int           `json:"line"`
	Column      int           `json:"column"`
	EndLine     int           `json:"end_line,omitempty"`
	EndColumn   int           `json:"end_column,omitempty"`
	Context     string        `json:"context,omitempty"`
	Rule        string        `json:"rule"`
	Suggestion  string        `json:"suggestion,omitempty"`
	CodeSnippet string        `json:"code_snippet,omitempty"`
}

// Location represents a position in source code
type Location struct {
	File      string `json:"file"`
	Line      int    `json:"line"`
	Column    int    `json:"column"`
	EndLine   int    `json:"end_line,omitempty"`
	EndColumn int    `json:"end_column,omitempty"`
}