// Package violations provides detectors for various clean code violations in Go source code.
package violations

import (
	"github.com/ericfisherdev/goclean/internal/models"
)

// Detector defines the interface for violation detection
type Detector interface {
	// Detect analyzes the provided file information and returns violations
	Detect(fileInfo *models.FileInfo, astInfo interface{}) []*models.Violation
	
	// Name returns the name of this detector
	Name() string
	
	// Description returns a description of what this detector checks for
	Description() string
}

// DetectorRegistry manages all available detectors
type DetectorRegistry struct {
	detectors []Detector
}

// NewDetectorRegistry creates a new detector registry
func NewDetectorRegistry() *DetectorRegistry {
	return &DetectorRegistry{
		detectors: make([]Detector, 0),
	}
}

// RegisterDetector adds a detector to the registry
func (r *DetectorRegistry) RegisterDetector(detector Detector) {
	r.detectors = append(r.detectors, detector)
}

// GetDetectors returns all registered detectors
func (r *DetectorRegistry) GetDetectors() []Detector {
	return r.detectors
}

// DetectAll runs all detectors against the provided file information
func (r *DetectorRegistry) DetectAll(fileInfo *models.FileInfo, astInfo interface{}) []*models.Violation {
	var allViolations []*models.Violation
	
	for _, detector := range r.detectors {
		violations := detector.Detect(fileInfo, astInfo)
		allViolations = append(allViolations, violations...)
	}
	
	return allViolations
}

// DetectorConfig provides configuration for violation detection
type DetectorConfig struct {
	// Function thresholds
	MaxFunctionLines      int
	MaxCyclomaticComplexity int
	MaxParameters        int
	MaxNestingDepth      int
	
	// Code structure thresholds
	MaxClassLines        int
	MaxMethods          int
	
	// Naming convention rules
	AllowSingleLetterVars bool
	RequireCamelCase     bool
	RequireCommentsForPublic bool
	
	// Test file handling
	AggressiveMode       bool
	SkipTestFiles        bool
	Verbose              bool
	
	// Severity classification config
	SeverityConfig *SeverityConfig
	
	// Rust-specific configuration
	RustConfig *RustDetectorConfig
}

// RustDetectorConfig provides Rust-specific detector configuration
type RustDetectorConfig struct {
	// Ownership and borrowing
	EnableOwnershipAnalysis bool
	MaxLifetimeParams       int
	DetectUnnecessaryClones bool
	
	// Error handling
	EnableErrorHandlingCheck bool
	AllowUnwrap             bool
	AllowExpect             bool
	EnforceResultPropagation bool
	
	// Pattern matching
	EnablePatternMatchCheck  bool
	RequireExhaustiveMatch   bool
	MaxNestedMatchDepth      int
	
	// Trait and impl
	MaxTraitBounds          int
	MaxImplMethods          int
	DetectOrphanInstances   bool
	
	// Unsafe code
	AllowUnsafe             bool
	RequireUnsafeComments   bool
	DetectTransmuteUsage    bool
	
	// Performance
	DetectIneffcientString  bool
	DetectBoxedPrimitives   bool
	DetectBlockingInAsync   bool
	
	// Macro analysis
	MaxMacroComplexity      int
	AllowRecursiveMacros    bool
	
	// Module structure
	MaxModuleDepth          int
	MaxFileLines            int
	
	// Naming conventions
	EnforceSnakeCase        bool
	EnforcePascalCase       bool
	EnforceScreamingSnake   bool
}

// DefaultDetectorConfig returns the default configuration
func DefaultDetectorConfig() *DetectorConfig {
	return &DetectorConfig{
		MaxFunctionLines:      25,
		MaxCyclomaticComplexity: 8,
		MaxParameters:        4,
		MaxNestingDepth:      3,
		MaxClassLines:        150,
		MaxMethods:          20,
		AllowSingleLetterVars: true,
		RequireCamelCase:     true,
		RequireCommentsForPublic: true,
		AggressiveMode:       false,
		SkipTestFiles:        true,
		Verbose:              false,
		SeverityConfig:       DefaultSeverityConfig(),
		RustConfig:           DefaultRustDetectorConfig(),
	}
}

// DefaultRustDetectorConfig returns the default Rust detector configuration
func DefaultRustDetectorConfig() *RustDetectorConfig {
	return &RustDetectorConfig{
		// Ownership and borrowing
		EnableOwnershipAnalysis: true,
		MaxLifetimeParams:       3,
		DetectUnnecessaryClones: true,
		
		// Error handling
		EnableErrorHandlingCheck: true,
		AllowUnwrap:             false,
		AllowExpect:             false,
		EnforceResultPropagation: true,
		
		// Pattern matching
		EnablePatternMatchCheck:  true,
		RequireExhaustiveMatch:   true,
		MaxNestedMatchDepth:      3,
		
		// Trait and impl
		MaxTraitBounds:          5,
		MaxImplMethods:          20,
		DetectOrphanInstances:   true,
		
		// Unsafe code
		AllowUnsafe:             true,  // Allow but track
		RequireUnsafeComments:   true,
		DetectTransmuteUsage:    true,
		
		// Performance
		DetectIneffcientString:  true,
		DetectBoxedPrimitives:   true,
		DetectBlockingInAsync:   true,
		
		// Macro analysis
		MaxMacroComplexity:      10,
		AllowRecursiveMacros:    false,
		
		// Module structure
		MaxModuleDepth:          5,
		MaxFileLines:            500,
		
		// Naming conventions
		EnforceSnakeCase:        true,
		EnforcePascalCase:       true,
		EnforceScreamingSnake:   true,
	}
}

// GetSeverityClassifier returns a severity classifier instance for the detector config
func (c *DetectorConfig) GetSeverityClassifier() *SeverityClassifier {
	if c.SeverityConfig == nil {
		return NewSeverityClassifier(DefaultSeverityConfig())
	}
	return NewSeverityClassifier(c.SeverityConfig)
}

// ClassifyViolationSeverity provides a convenience method for consistent severity classification
func (c *DetectorConfig) ClassifyViolationSeverity(violationType models.ViolationType, actualValue, threshold int, context *ViolationContext) models.Severity {
	classifier := c.GetSeverityClassifier()
	return classifier.ClassifySeverity(violationType, actualValue, threshold, context)
}

// ClassifyViolationSeverityFloat provides severity classification for floating-point values
func (c *DetectorConfig) ClassifyViolationSeverityFloat(violationType models.ViolationType, actualValue, threshold float64, context *ViolationContext) models.Severity {
	classifier := c.GetSeverityClassifier()
	return classifier.ClassifySeverityFloat(violationType, actualValue, threshold, context)
}