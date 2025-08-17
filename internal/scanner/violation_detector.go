package scanner

import (
	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/types"
	"github.com/ericfisherdev/goclean/internal/violations"
)

// ViolationDetector manages violation detection during scanning
type ViolationDetector struct {
	registry                 *violations.DetectorRegistry
	duplicationDetector      *violations.DuplicationDetector
	rustDuplicationDetector  *violations.RustDuplicationDetector
	config                   *violations.DetectorConfig
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
	
	// Register Rust-specific detectors
	registry.RegisterDetector(violations.NewRustFunctionDetector(config))
	registry.RegisterDetector(violations.NewRustNamingDetector(config))
	registry.RegisterDetector(violations.NewRustDocumentationDetector(config))
	registry.RegisterDetector(violations.NewRustMagicNumberDetector(config))
	registry.RegisterDetector(violations.NewRustStructureDetector(config))
	registry.RegisterDetector(violations.NewRustOwnershipDetector(config))
	registry.RegisterDetector(violations.NewRustErrorHandlingDetector(config))
	
	// Register advanced detectors
	registry.RegisterDetector(violations.NewMagicNumberDetector(config))
	registry.RegisterDetector(violations.NewCommentedCodeDetector(config))
	registry.RegisterDetector(violations.NewTodoTrackerDetector(config))
	registry.RegisterDetector(violations.NewDocumentationDetector(config))
	
	// Create duplication detectors separately (needs special handling)
	duplicationDetector := violations.NewDuplicationDetector(config)
	rustDuplicationDetector := violations.NewRustDuplicationDetector(config)
	
	return &ViolationDetector{
		registry:                registry,
		duplicationDetector:     duplicationDetector,
		rustDuplicationDetector: rustDuplicationDetector,
		config:                  config,
	}
}

// DetectViolations detects all violations in a file
func (vd *ViolationDetector) DetectViolations(result *models.ScanResult) {
	if result == nil || result.File == nil {
		return
	}
	
	// Run standard detectors on the AST info (Go or Rust)
	violations := vd.registry.DetectAll(result.File, result.ASTInfo)
	
	// Run duplication detectors (needs special handling as they compare across files)
	if result.ASTInfo != nil {
		if goAstInfo, ok := result.ASTInfo.(*types.GoASTInfo); ok {
			dupViolations := vd.duplicationDetector.Detect(result.File, goAstInfo)
			violations = append(violations, dupViolations...)
		} else if rustAstInfo, ok := result.ASTInfo.(*types.RustASTInfo); ok {
			rustDupViolations := vd.rustDuplicationDetector.Detect(result.File, rustAstInfo)
			violations = append(violations, rustDupViolations...)
		}
	}
	
	// Add violations to the result
	result.Violations = violations
}

// ResetDuplicationCache resets the duplication detectors' caches
// This should be called at the start of each new scan
func (vd *ViolationDetector) ResetDuplicationCache() {
	vd.duplicationDetector.Reset()
	vd.rustDuplicationDetector.Reset()
}

// GetConfig returns the detector configuration
func (vd *ViolationDetector) GetConfig() *violations.DetectorConfig {
	return vd.config
}