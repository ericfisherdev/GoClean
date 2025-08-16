package scanner

import (
	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/types"
	"github.com/ericfisherdev/goclean/internal/violations"
)

// ViolationDetector manages violation detection during scanning
type ViolationDetector struct {
	registry            *violations.DetectorRegistry
	duplicationDetector *violations.DuplicationDetector
	config              *violations.DetectorConfig
}

// NewViolationDetector creates a new violation detector
func NewViolationDetector(config *violations.DetectorConfig) *ViolationDetector {
	if config == nil {
		config = violations.DefaultDetectorConfig()
	}
	
	registry := violations.NewDetectorRegistry()
	
	// Register standard detectors
	registry.RegisterDetector(violations.NewFunctionDetector(config))
	registry.RegisterDetector(violations.NewNamingDetector(config))
	registry.RegisterDetector(violations.NewStructureDetector(config))
	
	// Register Go-specific detectors
	registry.RegisterDetector(violations.NewGoStandardNamingDetector(config))
	
	// Register advanced detectors
	registry.RegisterDetector(violations.NewMagicNumberDetector(config))
	registry.RegisterDetector(violations.NewCommentedCodeDetector(config))
	registry.RegisterDetector(violations.NewTodoTrackerDetector(config))
	registry.RegisterDetector(violations.NewDocumentationDetector(config))
	
	// Create duplication detector separately (needs special handling)
	duplicationDetector := violations.NewDuplicationDetector(config)
	
	return &ViolationDetector{
		registry:            registry,
		duplicationDetector: duplicationDetector,
		config:              config,
	}
}

// DetectViolations detects all violations in a file
func (vd *ViolationDetector) DetectViolations(result *models.ScanResult) {
	if result == nil || result.File == nil {
		return
	}
	
	// Get AST info from the scan result
	var astInfo *types.GoASTInfo
	if result.ASTInfo != nil {
		if goAstInfo, ok := result.ASTInfo.(*types.GoASTInfo); ok {
			astInfo = goAstInfo
		}
	}
	
	// Run standard detectors
	violations := vd.registry.DetectAll(result.File, astInfo)
	
	// Run duplication detector (needs special handling as it compares across files)
	if astInfo != nil {
		dupViolations := vd.duplicationDetector.Detect(result.File, astInfo)
		violations = append(violations, dupViolations...)
	}
	
	// Add violations to the result
	result.Violations = violations
}

// ResetDuplicationCache resets the duplication detector's cache
// This should be called at the start of each new scan
func (vd *ViolationDetector) ResetDuplicationCache() {
	vd.duplicationDetector.Reset()
}

// GetConfig returns the detector configuration
func (vd *ViolationDetector) GetConfig() *violations.DetectorConfig {
	return vd.config
}