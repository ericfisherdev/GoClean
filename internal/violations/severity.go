package violations

import (
	"github.com/ericfisherdev/goclean/internal/models"
)

// SeverityClassifier provides centralized severity classification logic
type SeverityClassifier struct {
	config *SeverityConfig
}

// SeverityConfig defines thresholds for severity classification
type SeverityConfig struct {
	// Multipliers for different severity levels
	LowThresholdMultiplier      float64 `yaml:"low_threshold_multiplier"`      // Default: 1.0 (at threshold)
	MediumThresholdMultiplier   float64 `yaml:"medium_threshold_multiplier"`   // Default: 1.5
	HighThresholdMultiplier     float64 `yaml:"high_threshold_multiplier"`     // Default: 2.0
	CriticalThresholdMultiplier float64 `yaml:"critical_threshold_multiplier"` // Default: 3.0

	// Context-based severity adjustments
	PublicFunctionSeverityBoost bool `yaml:"public_function_severity_boost"` // Boost severity for public functions
	TestFilesSeverityReduction  bool `yaml:"test_files_severity_reduction"`  // Reduce severity for test files
	LegacyCodeSeverityReduction bool `yaml:"legacy_code_severity_reduction"` // Reduce severity for legacy code

	// Violation type specific severity overrides
	ViolationTypeWeights map[models.ViolationType]float64 `yaml:"violation_type_weights"`
}

// ViolationContext provides context information for severity calculation
type ViolationContext struct {
	IsPublic       bool   `json:"is_public"`
	IsTestFile     bool   `json:"is_test_file"`
	IsLegacyCode   bool   `json:"is_legacy_code"`
	FileExtension  string `json:"file_extension"`
	PackageName    string `json:"package_name"`
	FunctionName   string `json:"function_name,omitempty"`
	TypeName       string `json:"type_name,omitempty"`
	ProjectContext string `json:"project_context,omitempty"`
}

// NewSeverityClassifier creates a new severity classifier
func NewSeverityClassifier(config *SeverityConfig) *SeverityClassifier {
	if config == nil {
		config = DefaultSeverityConfig()
	}
	return &SeverityClassifier{
		config: config,
	}
}

// DefaultSeverityConfig returns the default severity configuration
func DefaultSeverityConfig() *SeverityConfig {
	return &SeverityConfig{
		LowThresholdMultiplier:      1.0,
		MediumThresholdMultiplier:   1.5,
		HighThresholdMultiplier:     2.0,
		CriticalThresholdMultiplier: 3.0,
		PublicFunctionSeverityBoost: true,
		TestFilesSeverityReduction:  true,
		LegacyCodeSeverityReduction: false,
		ViolationTypeWeights: map[models.ViolationType]float64{
			models.ViolationTypeFunctionLength:         1.0,
			models.ViolationTypeCyclomaticComplexity:  1.2, // More critical
			models.ViolationTypeParameterCount:        0.8,
			models.ViolationTypeNestingDepth:         1.1,
			models.ViolationTypeNaming:               0.7,
			models.ViolationTypeClassSize:            1.0,
			models.ViolationTypeMissingDocumentation: 0.6, // Less critical
			models.ViolationTypeMagicNumbers:         0.5, // Least critical
			models.ViolationTypeDuplication:          1.3, // Very critical
		},
	}
}

// ClassifySeverity calculates the severity level based on violation metrics and context
func (s *SeverityClassifier) ClassifySeverity(violationType models.ViolationType, actualValue, threshold int, context *ViolationContext) models.Severity {
	// Calculate base severity based on threshold multipliers
	baseSeverity := s.calculateBaseSeverity(actualValue, threshold)
	
	// Apply violation type weighting
	adjustedSeverity := s.applyViolationTypeWeight(baseSeverity, violationType)
	
	// Apply context-based adjustments
	finalSeverity := s.applyContextAdjustments(adjustedSeverity, context)
	
	return finalSeverity
}

// ClassifySeverityFloat calculates severity for floating-point values
func (s *SeverityClassifier) ClassifySeverityFloat(violationType models.ViolationType, actualValue, threshold float64, context *ViolationContext) models.Severity {
	// Calculate base severity using the same logic but with floats
	baseSeverity := s.calculateBaseSeverityFloat(actualValue, threshold)
	
	// Apply violation type weighting
	adjustedSeverity := s.applyViolationTypeWeight(baseSeverity, violationType)
	
	// Apply context-based adjustments
	finalSeverity := s.applyContextAdjustments(adjustedSeverity, context)
	
	return finalSeverity
}

// calculateBaseSeverity determines base severity level based on threshold multipliers
func (s *SeverityClassifier) calculateBaseSeverity(actualValue, threshold int) models.Severity {
	ratio := float64(actualValue) / float64(threshold)
	
	if ratio >= s.config.CriticalThresholdMultiplier {
		return models.SeverityCritical
	}
	if ratio >= s.config.HighThresholdMultiplier {
		return models.SeverityHigh
	}
	if ratio >= s.config.MediumThresholdMultiplier {
		return models.SeverityMedium
	}
	
	return models.SeverityLow
}

// calculateBaseSeverityFloat determines base severity for float values
func (s *SeverityClassifier) calculateBaseSeverityFloat(actualValue, threshold float64) models.Severity {
	ratio := actualValue / threshold
	
	if ratio >= s.config.CriticalThresholdMultiplier {
		return models.SeverityCritical
	}
	if ratio >= s.config.HighThresholdMultiplier {
		return models.SeverityHigh
	}
	if ratio >= s.config.MediumThresholdMultiplier {
		return models.SeverityMedium
	}
	
	return models.SeverityLow
}

// applyViolationTypeWeight adjusts severity based on violation type importance
func (s *SeverityClassifier) applyViolationTypeWeight(baseSeverity models.Severity, violationType models.ViolationType) models.Severity {
	weight, exists := s.config.ViolationTypeWeights[violationType]
	if !exists {
		weight = 1.0 // Default weight
	}
	
	// Apply weight adjustment based on thresholds
	if weight >= 1.2 && baseSeverity < models.SeverityCritical {
		return s.increaseSeverity(baseSeverity)
	} else if weight <= 0.8 && baseSeverity > models.SeverityLow {
		return s.decreaseSeverity(baseSeverity)
	}
	
	return baseSeverity
}

// applyContextAdjustments applies context-based severity modifications
func (s *SeverityClassifier) applyContextAdjustments(baseSeverity models.Severity, context *ViolationContext) models.Severity {
	if context == nil {
		return baseSeverity
	}
	
	adjustedSeverity := baseSeverity
	
	// Boost severity for public APIs (they have broader impact)
	if s.config.PublicFunctionSeverityBoost && context.IsPublic && adjustedSeverity < models.SeverityCritical {
		adjustedSeverity = s.increaseSeverity(adjustedSeverity)
	}
	
	// Reduce severity for test files (less critical for production)
	if s.config.TestFilesSeverityReduction && context.IsTestFile && adjustedSeverity > models.SeverityLow {
		adjustedSeverity = s.decreaseSeverity(adjustedSeverity)
	}
	
	// Reduce severity for legacy code (harder to refactor)
	if s.config.LegacyCodeSeverityReduction && context.IsLegacyCode && adjustedSeverity > models.SeverityLow {
		adjustedSeverity = s.decreaseSeverity(adjustedSeverity)
	}
	
	return adjustedSeverity
}

// increaseSeverity moves to the next higher severity level
func (s *SeverityClassifier) increaseSeverity(severity models.Severity) models.Severity {
	switch severity {
	case models.SeverityLow:
		return models.SeverityMedium
	case models.SeverityMedium:
		return models.SeverityHigh
	case models.SeverityHigh:
		return models.SeverityCritical
	default:
		return severity // Already at max or unknown
	}
}

// decreaseSeverity moves to the next lower severity level
func (s *SeverityClassifier) decreaseSeverity(severity models.Severity) models.Severity {
	switch severity {
	case models.SeverityCritical:
		return models.SeverityHigh
	case models.SeverityHigh:
		return models.SeverityMedium
	case models.SeverityMedium:
		return models.SeverityLow
	default:
		return severity // Already at min or unknown
	}
}

// GetSeverityDescription returns a human-readable description of the severity level
func (s *SeverityClassifier) GetSeverityDescription(severity models.Severity) string {
	switch severity {
	case models.SeverityLow:
		return "Minor issue that should be addressed when convenient"
	case models.SeverityMedium:
		return "Moderate issue that should be addressed in the next refactoring cycle"
	case models.SeverityHigh:
		return "Significant issue that should be addressed soon as it impacts code quality"
	case models.SeverityCritical:
		return "Critical issue that requires immediate attention as it severely impacts maintainability"
	default:
		return "Unknown severity level"
	}
}

// GetSeverityWeight returns the numeric weight for sorting and prioritization
func (s *SeverityClassifier) GetSeverityWeight(severity models.Severity) int {
	switch severity {
	case models.SeverityLow:
		return 1
	case models.SeverityMedium:
		return 2
	case models.SeverityHigh:
		return 3
	case models.SeverityCritical:
		return 4
	default:
		return 0
	}
}

// IsContextBasedAdjustmentEnabled checks if context-based adjustments are enabled
func (s *SeverityClassifier) IsContextBasedAdjustmentEnabled() bool {
	return s.config.PublicFunctionSeverityBoost || 
		   s.config.TestFilesSeverityReduction || 
		   s.config.LegacyCodeSeverityReduction
}

// GetThresholdMultiplier returns the configured multiplier for a severity level
func (s *SeverityClassifier) GetThresholdMultiplier(severity models.Severity) float64 {
	switch severity {
	case models.SeverityLow:
		return s.config.LowThresholdMultiplier
	case models.SeverityMedium:
		return s.config.MediumThresholdMultiplier
	case models.SeverityHigh:
		return s.config.HighThresholdMultiplier
	case models.SeverityCritical:
		return s.config.CriticalThresholdMultiplier
	default:
		return 1.0
	}
}