package violations

import (
	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/scanner"
)

// Detector defines the interface for violation detection
type Detector interface {
	// Detect analyzes the provided file information and returns violations
	Detect(fileInfo *models.FileInfo, astInfo *scanner.GoASTInfo) []*models.Violation
	
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
func (r *DetectorRegistry) DetectAll(fileInfo *models.FileInfo, astInfo *scanner.GoASTInfo) []*models.Violation {
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
		AllowSingleLetterVars: false,
		RequireCamelCase:     true,
		RequireCommentsForPublic: true,
	}
}