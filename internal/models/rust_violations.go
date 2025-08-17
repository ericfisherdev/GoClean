package models

// RustViolationCategory represents categories of Rust-specific violations
type RustViolationCategory string

const (
	RustCategoryNaming         RustViolationCategory = "naming"
	RustCategorySafety         RustViolationCategory = "safety"
	RustCategoryOwnership      RustViolationCategory = "ownership"
	RustCategoryPerformance    RustViolationCategory = "performance"
	RustCategoryErrorHandling  RustViolationCategory = "error_handling"
	RustCategoryPatternMatching RustViolationCategory = "pattern_matching"
	RustCategoryTraits         RustViolationCategory = "traits"
	RustCategoryMacros         RustViolationCategory = "macros"
	RustCategoryAsync          RustViolationCategory = "async"
	RustCategoryModules        RustViolationCategory = "modules"
)

// GetRustViolationCategory returns the category for a given Rust violation type
func GetRustViolationCategory(violationType ViolationType) RustViolationCategory {
	switch violationType {
	// Naming violations
	case ViolationTypeRustInvalidFunctionNaming,
		 ViolationTypeRustInvalidStructNaming,
		 ViolationTypeRustInvalidEnumNaming,
		 ViolationTypeRustInvalidTraitNaming,
		 ViolationTypeRustInvalidConstantNaming,
		 ViolationTypeRustInvalidModuleNaming,
		 ViolationTypeRustInvalidVariableNaming:
		return RustCategoryNaming
		
	// Safety violations
	case ViolationTypeRustUnnecessaryUnsafe,
		 ViolationTypeRustUnsafeWithoutComment,
		 ViolationTypeRustTransmuteAbuse,
		 ViolationTypeRustRawPointerAbuse:
		return RustCategorySafety
		
	// Ownership violations
	case ViolationTypeRustUnnecessaryClone,
		 ViolationTypeRustInefficientBorrowing,
		 ViolationTypeRustComplexLifetime,
		 ViolationTypeRustMoveSemanticsViolation,
		 ViolationTypeRustBorrowCheckerBypass:
		return RustCategoryOwnership
		
	// Performance violations
	case ViolationTypeRustInefficientStringConcat,
		 ViolationTypeRustUnnecessaryAllocation,
		 ViolationTypeRustBlockingInAsync,
		 ViolationTypeRustInefficientIteration,
		 ViolationTypeRustUnnecessaryCollection:
		return RustCategoryPerformance
		
	// Error handling violations
	case ViolationTypeRustOveruseUnwrap,
		 ViolationTypeRustMissingErrorPropagation,
		 ViolationTypeRustInconsistentErrorType,
		 ViolationTypeRustPanicProneCode,
		 ViolationTypeRustUnhandledResult,
		 ViolationTypeRustImproperExpect:
		return RustCategoryErrorHandling
		
	// Pattern matching violations
	case ViolationTypeRustNonExhaustiveMatch,
		 ViolationTypeRustNestedPatternMatching,
		 ViolationTypeRustInefficientDestructuring,
		 ViolationTypeRustUnreachablePattern,
		 ViolationTypeRustMissingMatchArm:
		return RustCategoryPatternMatching
		
	// Trait violations
	case ViolationTypeRustOverlyComplexTrait,
		 ViolationTypeRustMissingTraitImpl,
		 ViolationTypeRustTraitBoundComplexity,
		 ViolationTypeRustOrphanRule:
		return RustCategoryTraits
		
	// Macro violations
	case ViolationTypeRustMacroComplexity,
		 ViolationTypeRustMacroHygiene,
		 ViolationTypeRustProceduralMacroMisuse:
		return RustCategoryMacros
		
	// Async violations
	case ViolationTypeRustAsyncFnInTrait,
		 ViolationTypeRustSendSyncViolation,
		 ViolationTypeRustDeadlockProne,
		 ViolationTypeRustRaceCondition:
		return RustCategoryAsync
		
	// Module violations
	case ViolationTypeRustImproperVisibility,
		 ViolationTypeRustCircularDependency,
		 ViolationTypeRustModuleOrganization,
		 ViolationTypeRustUnusedImport:
		return RustCategoryModules
		
	default:
		return ""
	}
}

// IsRustSpecificViolation checks if a violation type is Rust-specific
func IsRustSpecificViolation(violationType ViolationType) bool {
	return GetRustViolationCategory(violationType) != ""
}

// GetRustViolationDescription returns a human-readable description for Rust violation types
func GetRustViolationDescription(violationType ViolationType) string {
	switch violationType {
	// Naming violations
	case ViolationTypeRustInvalidFunctionNaming:
		return "Function name does not follow Rust snake_case convention"
	case ViolationTypeRustInvalidStructNaming:
		return "Struct name does not follow Rust PascalCase convention"
	case ViolationTypeRustInvalidEnumNaming:
		return "Enum name does not follow Rust PascalCase convention"
	case ViolationTypeRustInvalidTraitNaming:
		return "Trait name does not follow Rust PascalCase convention"
	case ViolationTypeRustInvalidConstantNaming:
		return "Constant name does not follow Rust SCREAMING_SNAKE_CASE convention"
	case ViolationTypeRustInvalidModuleNaming:
		return "Module name does not follow Rust snake_case convention"
	case ViolationTypeRustInvalidVariableNaming:
		return "Variable name does not follow Rust snake_case convention"
		
	// Safety violations
	case ViolationTypeRustUnnecessaryUnsafe:
		return "Unnecessary use of unsafe block - code can be safe"
	case ViolationTypeRustUnsafeWithoutComment:
		return "Unsafe block lacks explanation comment"
	case ViolationTypeRustTransmuteAbuse:
		return "Dangerous or unnecessary use of mem::transmute"
	case ViolationTypeRustRawPointerAbuse:
		return "Improper raw pointer usage without proper safety guarantees"
		
	// Ownership violations
	case ViolationTypeRustUnnecessaryClone:
		return "Unnecessary clone() call - consider borrowing instead"
	case ViolationTypeRustInefficientBorrowing:
		return "Inefficient borrowing pattern detected"
	case ViolationTypeRustComplexLifetime:
		return "Overly complex lifetime annotations that could be simplified"
	case ViolationTypeRustMoveSemanticsViolation:
		return "Improper use of move semantics"
	case ViolationTypeRustBorrowCheckerBypass:
		return "Attempt to bypass borrow checker with unsafe code"
		
	// Performance violations
	case ViolationTypeRustInefficientStringConcat:
		return "Inefficient string concatenation - consider using format! or StringBuilder"
	case ViolationTypeRustUnnecessaryAllocation:
		return "Unnecessary heap allocation detected"
	case ViolationTypeRustBlockingInAsync:
		return "Blocking operation in async context"
	case ViolationTypeRustInefficientIteration:
		return "Inefficient iteration pattern - consider using iterators"
	case ViolationTypeRustUnnecessaryCollection:
		return "Unnecessary collection allocation for simple operations"
		
	// Error handling violations
	case ViolationTypeRustOveruseUnwrap:
		return "Overuse of unwrap() - consider proper error handling"
	case ViolationTypeRustMissingErrorPropagation:
		return "Missing error propagation with ? operator"
	case ViolationTypeRustInconsistentErrorType:
		return "Inconsistent error types across function boundaries"
	case ViolationTypeRustPanicProneCode:
		return "Code pattern prone to panics"
	case ViolationTypeRustUnhandledResult:
		return "Result type not properly handled"
	case ViolationTypeRustImproperExpect:
		return "Improper use of expect() without descriptive message"
		
	// Pattern matching violations
	case ViolationTypeRustNonExhaustiveMatch:
		return "Non-exhaustive pattern matching"
	case ViolationTypeRustNestedPatternMatching:
		return "Overly nested pattern matching - consider refactoring"
	case ViolationTypeRustInefficientDestructuring:
		return "Inefficient pattern destructuring"
	case ViolationTypeRustUnreachablePattern:
		return "Unreachable pattern in match expression"
	case ViolationTypeRustMissingMatchArm:
		return "Missing match arm for important cases"
		
	// Trait violations
	case ViolationTypeRustOverlyComplexTrait:
		return "Trait is overly complex - consider splitting"
	case ViolationTypeRustMissingTraitImpl:
		return "Missing important trait implementation"
	case ViolationTypeRustTraitBoundComplexity:
		return "Overly complex trait bounds"
	case ViolationTypeRustOrphanRule:
		return "Violation of orphan rule for trait implementation"
		
	// Macro violations
	case ViolationTypeRustMacroComplexity:
		return "Macro is overly complex"
	case ViolationTypeRustMacroHygiene:
		return "Macro hygiene violation"
	case ViolationTypeRustProceduralMacroMisuse:
		return "Improper use of procedural macros"
		
	// Async violations
	case ViolationTypeRustAsyncFnInTrait:
		return "Async function in trait without proper bounds"
	case ViolationTypeRustSendSyncViolation:
		return "Send/Sync trait violation in concurrent code"
	case ViolationTypeRustDeadlockProne:
		return "Code pattern prone to deadlocks"
	case ViolationTypeRustRaceCondition:
		return "Potential race condition detected"
		
	// Module violations
	case ViolationTypeRustImproperVisibility:
		return "Improper visibility modifier usage"
	case ViolationTypeRustCircularDependency:
		return "Circular dependency between modules"
	case ViolationTypeRustModuleOrganization:
		return "Poor module organization structure"
	case ViolationTypeRustUnusedImport:
		return "Unused import statement"
		
	default:
		return "Unknown Rust violation"
	}
}

// GetRustViolationSuggestion returns a suggestion for fixing Rust violations
func GetRustViolationSuggestion(violationType ViolationType) string {
	switch violationType {
	// Naming violations
	case ViolationTypeRustInvalidFunctionNaming:
		return "Rename function to use snake_case (e.g., my_function)"
	case ViolationTypeRustInvalidStructNaming:
		return "Rename struct to use PascalCase (e.g., MyStruct)"
	case ViolationTypeRustInvalidEnumNaming:
		return "Rename enum to use PascalCase (e.g., MyEnum)"
	case ViolationTypeRustInvalidTraitNaming:
		return "Rename trait to use PascalCase (e.g., MyTrait)"
	case ViolationTypeRustInvalidConstantNaming:
		return "Rename constant to use SCREAMING_SNAKE_CASE (e.g., MY_CONSTANT)"
	case ViolationTypeRustInvalidModuleNaming:
		return "Rename module to use snake_case (e.g., my_module)"
	case ViolationTypeRustInvalidVariableNaming:
		return "Rename variable to use snake_case (e.g., my_variable)"
		
	// Safety violations
	case ViolationTypeRustUnnecessaryUnsafe:
		return "Remove unsafe block if the contained code is actually safe"
	case ViolationTypeRustUnsafeWithoutComment:
		return "Add comment explaining why unsafe is needed and what invariants are maintained"
	case ViolationTypeRustTransmuteAbuse:
		return "Replace transmute with safe alternatives like From/Into traits or proper casting"
	case ViolationTypeRustRawPointerAbuse:
		return "Use safe abstractions or ensure proper null checks and lifetime management"
		
	// Ownership violations
	case ViolationTypeRustUnnecessaryClone:
		return "Use borrowing (&) instead of cloning when possible"
	case ViolationTypeRustInefficientBorrowing:
		return "Optimize borrowing patterns to reduce unnecessary references"
	case ViolationTypeRustComplexLifetime:
		return "Simplify lifetime annotations or restructure code to reduce complexity"
	case ViolationTypeRustMoveSemanticsViolation:
		return "Understand and properly use Rust's move semantics"
	case ViolationTypeRustBorrowCheckerBypass:
		return "Work with the borrow checker instead of bypassing it with unsafe code"
		
	// Performance violations
	case ViolationTypeRustInefficientStringConcat:
		return "Use format!() macro or String::with_capacity() for efficient string building"
	case ViolationTypeRustUnnecessaryAllocation:
		return "Use stack allocation or borrowing instead of heap allocation"
	case ViolationTypeRustBlockingInAsync:
		return "Use async alternatives or tokio::task::spawn_blocking for CPU-bound work"
	case ViolationTypeRustInefficientIteration:
		return "Use iterator methods (map, filter, collect) instead of manual loops"
	case ViolationTypeRustUnnecessaryCollection:
		return "Use iterator chains or direct operations without intermediate collections"
		
	// Error handling violations
	case ViolationTypeRustOveruseUnwrap:
		return "Use pattern matching, if let, or ? operator for proper error handling"
	case ViolationTypeRustMissingErrorPropagation:
		return "Use ? operator to propagate errors up the call stack"
	case ViolationTypeRustInconsistentErrorType:
		return "Use consistent error types, consider using a crate like anyhow or thiserror"
	case ViolationTypeRustPanicProneCode:
		return "Add bounds checking or use safe alternatives that return Result/Option"
	case ViolationTypeRustUnhandledResult:
		return "Handle the Result with match, if let, or ? operator"
	case ViolationTypeRustImproperExpect:
		return "Provide descriptive message explaining why unwrapping is safe"
		
	// Pattern matching violations
	case ViolationTypeRustNonExhaustiveMatch:
		return "Add missing match arms or use _ => {} for catch-all"
	case ViolationTypeRustNestedPatternMatching:
		return "Extract nested matches into separate functions or use if let chains"
	case ViolationTypeRustInefficientDestructuring:
		return "Use more efficient destructuring patterns or partial destructuring"
	case ViolationTypeRustUnreachablePattern:
		return "Remove unreachable pattern or reorder match arms"
	case ViolationTypeRustMissingMatchArm:
		return "Add match arms for important cases instead of using catch-all"
		
	// Trait violations
	case ViolationTypeRustOverlyComplexTrait:
		return "Split large trait into smaller, focused traits"
	case ViolationTypeRustMissingTraitImpl:
		return "Implement required traits like Debug, Clone, or PartialEq where appropriate"
	case ViolationTypeRustTraitBoundComplexity:
		return "Simplify trait bounds or use where clauses for readability"
	case ViolationTypeRustOrphanRule:
		return "Implement trait for local type or create newtype wrapper"
		
	// Macro violations
	case ViolationTypeRustMacroComplexity:
		return "Break complex macro into smaller pieces or use functions instead"
	case ViolationTypeRustMacroHygiene:
		return "Use proper variable scoping in macros to avoid name conflicts"
	case ViolationTypeRustProceduralMacroMisuse:
		return "Consider if a procedural macro is necessary or if a regular function would suffice"
		
	// Async violations
	case ViolationTypeRustAsyncFnInTrait:
		return "Use async-trait crate or consider alternative design patterns"
	case ViolationTypeRustSendSyncViolation:
		return "Ensure types implement Send/Sync or use proper synchronization primitives"
	case ViolationTypeRustDeadlockProne:
		return "Acquire locks in consistent order or use timeout-based locking"
	case ViolationTypeRustRaceCondition:
		return "Use atomic operations or proper synchronization mechanisms"
		
	// Module violations
	case ViolationTypeRustImproperVisibility:
		return "Use appropriate visibility modifiers (pub, pub(crate), pub(super))"
	case ViolationTypeRustCircularDependency:
		return "Restructure modules to eliminate circular dependencies"
	case ViolationTypeRustModuleOrganization:
		return "Organize related functionality into logical module hierarchies"
	case ViolationTypeRustUnusedImport:
		return "Remove unused import or use #[allow(unused_imports)] if needed for conditional compilation"
		
	default:
		return "Refer to Rust documentation and best practices"
	}
}

// GetDefaultRustViolationSeverity returns the default severity for Rust violation types
func GetDefaultRustViolationSeverity(violationType ViolationType) Severity {
	switch violationType {
	// High severity (safety and correctness issues)
	case ViolationTypeRustTransmuteAbuse,
		 ViolationTypeRustRawPointerAbuse,
		 ViolationTypeRustBorrowCheckerBypass,
		 ViolationTypeRustPanicProneCode,
		 ViolationTypeRustDeadlockProne,
		 ViolationTypeRustRaceCondition,
		 ViolationTypeRustNonExhaustiveMatch:
		return SeverityHigh
		
	// Medium severity (performance and maintainability issues)
	case ViolationTypeRustUnnecessaryUnsafe,
		 ViolationTypeRustUnsafeWithoutComment,
		 ViolationTypeRustComplexLifetime,
		 ViolationTypeRustInefficientStringConcat,
		 ViolationTypeRustUnnecessaryAllocation,
		 ViolationTypeRustBlockingInAsync,
		 ViolationTypeRustOveruseUnwrap,
		 ViolationTypeRustMissingErrorPropagation,
		 ViolationTypeRustInconsistentErrorType,
		 ViolationTypeRustNestedPatternMatching,
		 ViolationTypeRustOverlyComplexTrait,
		 ViolationTypeRustTraitBoundComplexity,
		 ViolationTypeRustMacroComplexity,
		 ViolationTypeRustAsyncFnInTrait,
		 ViolationTypeRustSendSyncViolation,
		 ViolationTypeRustCircularDependency:
		return SeverityMedium
		
	// Low severity (style and best practice issues)
	case ViolationTypeRustInvalidFunctionNaming,
		 ViolationTypeRustInvalidStructNaming,
		 ViolationTypeRustInvalidEnumNaming,
		 ViolationTypeRustInvalidTraitNaming,
		 ViolationTypeRustInvalidConstantNaming,
		 ViolationTypeRustInvalidModuleNaming,
		 ViolationTypeRustInvalidVariableNaming,
		 ViolationTypeRustUnnecessaryClone,
		 ViolationTypeRustInefficientBorrowing,
		 ViolationTypeRustMoveSemanticsViolation,
		 ViolationTypeRustInefficientIteration,
		 ViolationTypeRustUnnecessaryCollection,
		 ViolationTypeRustUnhandledResult,
		 ViolationTypeRustImproperExpect,
		 ViolationTypeRustInefficientDestructuring,
		 ViolationTypeRustUnreachablePattern,
		 ViolationTypeRustMissingMatchArm,
		 ViolationTypeRustMissingTraitImpl,
		 ViolationTypeRustOrphanRule,
		 ViolationTypeRustMacroHygiene,
		 ViolationTypeRustProceduralMacroMisuse,
		 ViolationTypeRustImproperVisibility,
		 ViolationTypeRustModuleOrganization,
		 ViolationTypeRustUnusedImport:
		return SeverityLow
		
	default:
		return SeverityMedium
	}
}