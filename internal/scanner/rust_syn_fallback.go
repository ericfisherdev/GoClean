//go:build !cgo || no_rust
// +build !cgo no_rust

// Package scanner provides fallback functionality when CGO or Rust support is disabled.
// This file implements no-op functions for the Rust parsing interface.
package scanner

import (
	"fmt"
	"sync"

	"github.com/ericfisherdev/goclean/internal/types"
)

// Fallback implementation when CGO is disabled
type FallbackSynParser struct {
	initialized bool
	mutex       sync.RWMutex
}

// Global instance for fallback mode
var globalFallbackParser *FallbackSynParser
var fallbackOnce sync.Once

// GetGlobalSynParser returns a fallback parser instance when CGO is disabled
func GetGlobalSynParser() (*FallbackSynParser, error) {
	fallbackOnce.Do(func() {
		globalFallbackParser = &FallbackSynParser{
			initialized: false,
		}
	})
	return globalFallbackParser, nil
}

// Initialize initializes the fallback parser (no-op)
func (p *FallbackSynParser) Initialize() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	p.initialized = true
	return nil
}

// Cleanup cleans up the fallback parser (no-op)
func (p *FallbackSynParser) Cleanup() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	p.initialized = false
	return nil
}

// IsInitialized returns whether the fallback parser is initialized
func (p *FallbackSynParser) IsInitialized() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	
	return p.initialized
}

// ParseFile returns an error indicating Rust support is not available
func (p *FallbackSynParser) ParseFile(content, filePath string) (*types.RustASTInfo, error) {
	return nil, fmt.Errorf("Rust support is not available in this build (CGO disabled or Rust library not found)")
}

// ParseRustFile returns an error indicating Rust support is not available
func (p *FallbackSynParser) ParseRustFile(content []byte, filePath string) (*types.RustASTInfo, error) {
	return nil, fmt.Errorf("Rust support is not available in this build (CGO disabled or Rust library not found)")
}

// ValidateSyntax returns an error indicating Rust support is not available
func (p *FallbackSynParser) ValidateSyntax(content string) error {
	return fmt.Errorf("Rust syntax validation is not available in this build (CGO disabled or Rust library not found)")
}

// ParseExpression returns an error indicating Rust support is not available
func (p *FallbackSynParser) ParseExpression(expression string) (map[string]interface{}, error) {
	return nil, fmt.Errorf("Rust expression parsing is not available in this build (CGO disabled or Rust library not found)")
}

// GetCapabilities returns empty capabilities for fallback mode
func (p *FallbackSynParser) GetCapabilities() (*SynParserCapabilities, error) {
	return &SynParserCapabilities{
		Version:              "fallback-1.0.0",
		SupportedFeatures:    []string{},
		MaxFileSize:          0,
		ThreadSafe:           true,
		BatchProcessing:      false,
		SyntaxValidation:     false,
		ExpressionParsing:    false,
		ErrorRecovery:        false,
		IncrementalParsing:   false,
		PositionTracking:     false,
		CommentPreservation:  false,
		TokenStreaming:       false,
		CustomConfiguration:  false,
	}, nil
}

// GetVersion returns the fallback version
func (p *FallbackSynParser) GetVersion() (string, error) {
	return "fallback-1.0.0", nil
}

// CanParse always returns false for fallback mode
func (p *FallbackSynParser) CanParse(content string) bool {
	return false
}

// BatchParse returns an error indicating batch parsing is not available
func (p *FallbackSynParser) BatchParse(requests []ParseRequest) ([]ParseResult, error) {
	return nil, fmt.Errorf("Rust batch parsing is not available in this build (CGO disabled or Rust library not found)")
}

// GetParseStats returns empty stats for fallback mode
func (p *FallbackSynParser) GetParseStats(content string) (*ParseStats, error) {
	return &ParseStats{
		ParseTimeMs:    0,
		FileSize:       len(content),
		NodeCount:      0,
		ErrorCount:     1, // Indicate parsing failed
		WarningCount:   0,
		ComplexityScore: 0,
		MemoryUsageKB:  0,
		Features:       make(map[string]int),
		Language:       "rust",
		ParserVersion:  "fallback-1.0.0",
		Success:        false,
	}, nil
}

// Ensure fallback parser implements the same interface
var _ interface {
	Initialize() error
	Cleanup() error
	IsInitialized() bool
	ParseFile(string, string) (*types.RustASTInfo, error)
	ValidateSyntax(string) error
	ParseExpression(string) (map[string]interface{}, error)
	GetCapabilities() (*SynParserCapabilities, error)
	GetVersion() (string, error)
	CanParse(string) bool
	BatchParse([]ParseRequest) ([]ParseResult, error)
	GetParseStats(string) (*ParseStats, error)
} = (*FallbackSynParser)(nil)

// Helper functions for graceful degradation

// IsRustSupportAvailable returns false when using fallback implementation
func IsRustSupportAvailable() bool {
	return false
}

// GetRustSupportStatus returns status information about Rust support
func GetRustSupportStatus() map[string]interface{} {
	return map[string]interface{}{
		"available":     false,
		"reason":        "CGO disabled or Rust library not found",
		"parser_type":   "fallback",
		"capabilities":  []string{},
		"build_tags":    []string{"!cgo", "no_rust"},
		"version":       "fallback-1.0.0",
	}
}

// ParseFallbackMessage returns a user-friendly message about Rust support
func ParseFallbackMessage() string {
	return "Rust parsing is not available in this build. To enable Rust support:\n" +
		"1. Install the Rust toolchain (https://rustup.rs/)\n" +
		"2. Build with CGO enabled: CGO_ENABLED=1 go build\n" +
		"3. Ensure the Rust parser library is available\n" +
		"4. Use 'make build' instead of 'make build-go-only'"
}