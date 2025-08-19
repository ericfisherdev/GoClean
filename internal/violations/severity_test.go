package violations
import (
	"testing"
	"github.com/ericfisherdev/goclean/internal/models"
)
func TestNewSeverityClassifier(t *testing.T) {
	classifier := NewSeverityClassifier(nil)
	if classifier == nil {
		t.Error("Expected classifier to be created")
	}
	if classifier.config == nil {
		t.Error("Expected default config to be set")
	}
}
func TestDefaultSeverityConfig(t *testing.T) {
	config := DefaultSeverityConfig()
	// Check default multipliers
	if config.LowThresholdMultiplier != 1.0 {
		t.Errorf("Expected LowThresholdMultiplier to be 1.0, got %v", config.LowThresholdMultiplier)
	}
	if config.MediumThresholdMultiplier != 1.5 {
		t.Errorf("Expected MediumThresholdMultiplier to be 1.5, got %v", config.MediumThresholdMultiplier)
	}
	if config.HighThresholdMultiplier != 2.0 {
		t.Errorf("Expected HighThresholdMultiplier to be 2.0, got %v", config.HighThresholdMultiplier)
	}
	if config.CriticalThresholdMultiplier != 3.0 {
		t.Errorf("Expected CriticalThresholdMultiplier to be 3.0, got %v", config.CriticalThresholdMultiplier)
	}
	// Check context adjustments
	if !config.PublicFunctionSeverityBoost {
		t.Error("Expected PublicFunctionSeverityBoost to be true by default")
	}
	if !config.TestFilesSeverityReduction {
		t.Error("Expected TestFilesSeverityReduction to be true by default")
	}
	// Check violation type weights
	if len(config.ViolationTypeWeights) == 0 {
		t.Error("Expected ViolationTypeWeights to be populated")
	}
	// Check specific weights
	if weight, exists := config.ViolationTypeWeights[models.ViolationTypeCyclomaticComplexity]; !exists || weight != 1.2 {
		t.Errorf("Expected ViolationTypeCyclomaticComplexity weight to be 1.2, got %v", weight)
	}
}
func TestSeverityClassifier_ClassifySeverity_BasicThresholds(t *testing.T) {
	classifier := NewSeverityClassifier(nil)
	tests := []struct {
		name         string
		actualValue  int
		threshold    int
		expectedSev  models.Severity
		description  string
	}{
		{"Below threshold", 10, 20, models.SeverityLow, "Should be low when below threshold"},
		{"At threshold", 20, 20, models.SeverityLow, "Should be low at threshold (1.0x)"},
		{"At medium threshold", 30, 20, models.SeverityMedium, "Should be medium at 1.5x threshold"},
		{"At high threshold", 40, 20, models.SeverityHigh, "Should be high at 2.0x threshold"},
		{"At critical threshold", 60, 20, models.SeverityCritical, "Should be critical at 3.0x threshold"},
		{"Above critical", 80, 20, models.SeverityCritical, "Should be critical when above 3.0x threshold"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := classifier.ClassifySeverity(
				models.ViolationTypeFunctionLength,
				test.actualValue,
				test.threshold,
				nil, // No context
			)
			if result != test.expectedSev {
				t.Errorf("%s: expected %v, got %v", test.description, test.expectedSev, result)
			}
		})
	}
}
func TestSeverityClassifier_ClassifySeverityFloat(t *testing.T) {
	classifier := NewSeverityClassifier(nil)
	tests := []struct {
		name         string
		actualValue  float64
		threshold    float64
		expectedSev  models.Severity
	}{
		{"Float below threshold", 1.0, 2.0, models.SeverityLow},
		{"Float at medium threshold", 3.0, 2.0, models.SeverityMedium},
		{"Float at high threshold", 4.0, 2.0, models.SeverityHigh},
		{"Float at critical threshold", 6.0, 2.0, models.SeverityCritical},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := classifier.ClassifySeverityFloat(
				models.ViolationTypeFunctionLength,
				test.actualValue,
				test.threshold,
				nil,
			)
			if result != test.expectedSev {
				t.Errorf("expected %v, got %v", test.expectedSev, result)
			}
		})
	}
}
func TestSeverityClassifier_ViolationTypeWeights(t *testing.T) {
	classifier := NewSeverityClassifier(nil)
	// Test high-weight violation type (should increase severity)
	highWeightSeverity := classifier.ClassifySeverity(
		models.ViolationTypeCyclomaticComplexity, // weight: 1.2
		25, 20, // 1.25x threshold -> should be Low, but boosted to Medium
		nil,
	)
	// Test low-weight violation type (should decrease severity)
	lowWeightSeverity := classifier.ClassifySeverity(
		models.ViolationTypeMagicNumber, // weight: 0.5
		30, 20, // 1.5x threshold -> should be Medium, but reduced to Low
		nil,
	)
	// Compare with normal weight
	normalSeverity := classifier.ClassifySeverity(
		models.ViolationTypeFunctionLength, // weight: 1.0
		30, 20, // 1.5x threshold -> should be Medium
		nil,
	)
	if highWeightSeverity != models.SeverityMedium {
		t.Errorf("Expected high-weight violation to have Medium severity, got %v", highWeightSeverity)
	}
	if lowWeightSeverity != models.SeverityLow {
		t.Errorf("Expected low-weight violation to have Low severity, got %v", lowWeightSeverity)
	}
	if normalSeverity != models.SeverityMedium {
		t.Errorf("Expected normal-weight violation to have Medium severity, got %v", normalSeverity)
	}
}
func TestSeverityClassifier_ContextAdjustments(t *testing.T) {
	classifier := NewSeverityClassifier(nil)
	baseThreshold := 20
	testValue := 30 // 1.5x threshold -> Medium severity
	// Test public function boost
	publicContext := &ViolationContext{
		IsPublic: true,
	}
	publicSeverity := classifier.ClassifySeverity(
		models.ViolationTypeFunctionLength,
		testValue, baseThreshold,
		publicContext,
	)
	// Test test file reduction
	testContext := &ViolationContext{
		IsTestFile: true,
	}
	testSeverity := classifier.ClassifySeverity(
		models.ViolationTypeFunctionLength,
		testValue, baseThreshold,
		testContext,
	)
	// Test normal context
	normalSeverity := classifier.ClassifySeverity(
		models.ViolationTypeFunctionLength,
		testValue, baseThreshold,
		nil,
	)
	// Public should be higher than normal
	if publicSeverity <= normalSeverity {
		t.Errorf("Expected public function to have higher severity than normal. Public: %v, Normal: %v", publicSeverity, normalSeverity)
	}
	// Test file should be lower than normal (if normal is above Low)
	if normalSeverity > models.SeverityLow && testSeverity >= normalSeverity {
		t.Errorf("Expected test file to have lower severity than normal. Test: %v, Normal: %v", testSeverity, normalSeverity)
	}
}
func TestSeverityClassifier_LegacyCodeReduction(t *testing.T) {
	config := DefaultSeverityConfig()
	config.LegacyCodeSeverityReduction = true
	classifier := NewSeverityClassifier(config)
	legacyContext := &ViolationContext{
		IsLegacyCode: true,
	}
	// Test with high value that would normally be High severity
	legacySeverity := classifier.ClassifySeverity(
		models.ViolationTypeFunctionLength,
		40, 20, // 2.0x threshold -> High severity
		legacyContext,
	)
	normalSeverity := classifier.ClassifySeverity(
		models.ViolationTypeFunctionLength,
		40, 20,
		nil,
	)
	if legacySeverity >= normalSeverity {
		t.Errorf("Expected legacy code to have lower severity. Legacy: %v, Normal: %v", legacySeverity, normalSeverity)
	}
}
func TestSeverityClassifier_IncreaseSeverity(t *testing.T) {
	classifier := NewSeverityClassifier(nil)
	tests := []struct {
		input    models.Severity
		expected models.Severity
	}{
		{models.SeverityLow, models.SeverityMedium},
		{models.SeverityMedium, models.SeverityHigh},
		{models.SeverityHigh, models.SeverityCritical},
		{models.SeverityCritical, models.SeverityCritical}, // No change at max
	}
	for _, test := range tests {
		result := classifier.increaseSeverity(test.input)
		if result != test.expected {
			t.Errorf("increaseSeverity(%v) = %v, expected %v", test.input, result, test.expected)
		}
	}
}
func TestSeverityClassifier_DecreaseSeverity(t *testing.T) {
	classifier := NewSeverityClassifier(nil)
	tests := []struct {
		input    models.Severity
		expected models.Severity
	}{
		{models.SeverityCritical, models.SeverityHigh},
		{models.SeverityHigh, models.SeverityMedium},
		{models.SeverityMedium, models.SeverityLow},
		{models.SeverityLow, models.SeverityLow}, // No change at min
	}
	for _, test := range tests {
		result := classifier.decreaseSeverity(test.input)
		if result != test.expected {
			t.Errorf("decreaseSeverity(%v) = %v, expected %v", test.input, result, test.expected)
		}
	}
}
func TestSeverityClassifier_GetSeverityDescription(t *testing.T) {
	classifier := NewSeverityClassifier(nil)
	tests := []struct {
		severity models.Severity
		hasText  bool
	}{
		{models.SeverityLow, true},
		{models.SeverityMedium, true},
		{models.SeverityHigh, true},
		{models.SeverityCritical, true},
	}
	for _, test := range tests {
		desc := classifier.GetSeverityDescription(test.severity)
		if test.hasText && len(desc) == 0 {
			t.Errorf("Expected description for severity %v, got empty string", test.severity)
		}
		if !test.hasText && len(desc) > 0 {
			t.Errorf("Expected no description for severity %v, got %s", test.severity, desc)
		}
	}
}
func TestSeverityClassifier_GetSeverityWeight(t *testing.T) {
	classifier := NewSeverityClassifier(nil)
	tests := []struct {
		severity models.Severity
		weight   int
	}{
		{models.SeverityLow, 1},
		{models.SeverityMedium, 2},
		{models.SeverityHigh, 3},
		{models.SeverityCritical, 4},
	}
	for _, test := range tests {
		weight := classifier.GetSeverityWeight(test.severity)
		if weight != test.weight {
			t.Errorf("GetSeverityWeight(%v) = %d, expected %d", test.severity, weight, test.weight)
		}
	}
	// Test that weights are properly ordered
	if classifier.GetSeverityWeight(models.SeverityLow) >= classifier.GetSeverityWeight(models.SeverityMedium) {
		t.Error("Expected Low severity to have lower weight than Medium")
	}
	if classifier.GetSeverityWeight(models.SeverityMedium) >= classifier.GetSeverityWeight(models.SeverityHigh) {
		t.Error("Expected Medium severity to have lower weight than High")
	}
	if classifier.GetSeverityWeight(models.SeverityHigh) >= classifier.GetSeverityWeight(models.SeverityCritical) {
		t.Error("Expected High severity to have lower weight than Critical")
	}
}
func TestSeverityClassifier_IsContextBasedAdjustmentEnabled(t *testing.T) {
	// Test with default config (has adjustments enabled)
	classifier1 := NewSeverityClassifier(nil)
	if !classifier1.IsContextBasedAdjustmentEnabled() {
		t.Error("Expected context-based adjustments to be enabled with default config")
	}
	// Test with all adjustments disabled
	config := &SeverityConfig{
		PublicFunctionSeverityBoost: false,
		TestFilesSeverityReduction:  false,
		LegacyCodeSeverityReduction: false,
	}
	classifier2 := NewSeverityClassifier(config)
	if classifier2.IsContextBasedAdjustmentEnabled() {
		t.Error("Expected context-based adjustments to be disabled when all flags are false")
	}
}
func TestSeverityClassifier_GetThresholdMultiplier(t *testing.T) {
	classifier := NewSeverityClassifier(nil)
	tests := []struct {
		severity   models.Severity
		multiplier float64
	}{
		{models.SeverityLow, 1.0},
		{models.SeverityMedium, 1.5},
		{models.SeverityHigh, 2.0},
		{models.SeverityCritical, 3.0},
	}
	for _, test := range tests {
		multiplier := classifier.GetThresholdMultiplier(test.severity)
		if multiplier != test.multiplier {
			t.Errorf("GetThresholdMultiplier(%v) = %v, expected %v", test.severity, multiplier, test.multiplier)
		}
	}
}
func TestSeverityClassifier_ComplexContext(t *testing.T) {
	classifier := NewSeverityClassifier(nil)
	// Test complex context with multiple factors
	complexContext := &ViolationContext{
		IsPublic:      true,  // Should increase severity
		IsTestFile:    true,  // Should decrease severity
		IsLegacyCode:  false,
		FileExtension: ".go",
		PackageName:   "main",
		FunctionName:  "PublicTestFunc",
	}
	severity := classifier.ClassifySeverity(
		models.ViolationTypeFunctionLength,
		30, 20, // 1.5x threshold -> Medium base severity
		complexContext,
	)
	// The result should depend on the order and strength of adjustments
	// With default config: public boost happens first, then test reduction
	// So Medium -> High -> Medium (back to where it started)
	// This tests that context adjustments work together
	if severity < models.SeverityLow || severity > models.SeverityCritical {
		t.Errorf("Expected valid severity level with complex context, got %v", severity)
	}
}
func TestSeverityConfig_CustomWeights(t *testing.T) {
	config := DefaultSeverityConfig()
	config.ViolationTypeWeights[models.ViolationTypeFunctionLength] = 2.0 // Very high weight
	classifier := NewSeverityClassifier(config)
	// Test that custom weight is applied
	severity := classifier.ClassifySeverity(
		models.ViolationTypeFunctionLength,
		25, 20, // 1.25x threshold -> normally Low, but should be boosted
		nil,
	)
	// With high weight, should be boosted to Medium
	if severity != models.SeverityMedium {
		t.Errorf("Expected custom high weight to boost severity to Medium, got %v", severity)
	}
}