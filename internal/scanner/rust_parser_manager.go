package scanner

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ericfisherdev/goclean/internal/types"
)

// RustParserType represents the type of Rust parser being used
type RustParserType int

const (
	ParserTypeUnknown RustParserType = iota
	ParserTypeSyn                    // Syn crate via CGO
	ParserTypeRegex                  // Regex-based fallback
	ParserTypeFallback               // No-op fallback (when neither works)
)

func (pt RustParserType) String() string {
	switch pt {
	case ParserTypeSyn:
		return "syn-crate"
	case ParserTypeRegex:
		return "regex-fallback"
	case ParserTypeFallback:
		return "no-op-fallback"
	default:
		return "unknown"
	}
}

// RustParserManager manages different Rust parsing strategies with automatic fallback
type RustParserManager struct {
	currentParserType   RustParserType
	synParser          *RustSynParser
	regexAnalyzer      *RustASTAnalyzer
	fallbackNotified   int64  // atomic flag for notification
	verbose            bool
	mutex              sync.RWMutex
	initializationTime time.Time
	fallbackReason     string
	attemptCount       int64
	successCount       int64
	errorHistory       []error
}

// NewRustParserManager creates a new parser manager with automatic fallback detection
func NewRustParserManager(verbose bool) *RustParserManager {
	manager := &RustParserManager{
		currentParserType:  ParserTypeUnknown,
		verbose:           verbose,
		initializationTime: time.Now(),
		errorHistory:      make([]error, 0, 10), // Keep last 10 errors
	}

	// Initialize the best available parser
	manager.initializeBestParser()
	
	return manager
}

// initializeBestParser attempts to initialize parsers in order of preference
func (m *RustParserManager) initializeBestParser() {
	// Try syn crate first (best accuracy and performance)
	if m.tryInitializeSynParser() {
		return
	}

	// Fall back to regex parser (good compatibility, basic accuracy)
	if m.tryInitializeRegexParser() {
		return
	}

	// Last resort - fallback mode (no Rust parsing)
	m.initializeFallbackMode("Both syn crate and regex parsing failed to initialize")
}

// tryInitializeSynParser attempts to initialize the syn crate parser
func (m *RustParserManager) tryInitializeSynParser() bool {
	// Build-time check: only attempt if CGO is enabled
	if !isCGOEnabled() {
		m.recordError(fmt.Errorf("CGO is disabled, syn crate unavailable"))
		return false
	}

	// Runtime check: attempt to initialize syn parser
	synParser, err := NewRustSynParser(m.verbose)
	if err != nil {
		m.recordError(fmt.Errorf("syn parser initialization failed: %w", err))
		return false
	}

	// Verify syn parser is functional with a simple test
	if !m.verifySynParser(synParser) {
		synParser.Cleanup()
		return false
	}

	m.mutex.Lock()
	m.synParser = synParser
	m.currentParserType = ParserTypeSyn
	m.fallbackReason = ""
	m.mutex.Unlock()

	if m.verbose {
		fmt.Println("✅ Rust parsing: Using syn crate (optimal performance and accuracy)")
	}

	return true
}

// tryInitializeRegexParser attempts to initialize the regex-based parser
func (m *RustParserManager) tryInitializeRegexParser() bool {
	regexAnalyzer := NewRustASTAnalyzer(m.verbose)
	
	// Test regex parser with simple Rust code
	testCode := `fn test() { println!("hello"); }`
	_, err := regexAnalyzer.parseWithRegexFallback("test.rs", []byte(testCode))
	if err != nil {
		m.recordError(fmt.Errorf("regex parser test failed: %w", err))
		return false
	}

	// Acquire write lock to prevent races with state mutations
	m.mutex.Lock()
	m.regexAnalyzer = regexAnalyzer
	m.currentParserType = ParserTypeRegex
	m.fallbackReason = "Syn crate unavailable, using regex parsing"
	m.mutex.Unlock()

	if m.verbose {
		fmt.Println("⚠️  Rust parsing: Using regex fallback (limited accuracy)")
	}

	// Notify user about reduced functionality
	m.notifyFallback("Rust syn crate is not available. Using regex-based parsing with limited accuracy.")

	return true
}

// initializeFallbackMode sets up no-op fallback when no parsing is available
func (m *RustParserManager) initializeFallbackMode(reason string) {
	// Acquire write lock to prevent races with state mutations
	m.mutex.Lock()
	m.currentParserType = ParserTypeFallback
	m.fallbackReason = reason
	verbose := m.verbose // Read verbose value while locked
	m.mutex.Unlock()

	// Perform I/O operations after releasing the lock
	if verbose {
		fmt.Printf("❌ Rust parsing: No parser available (%s)\n", reason)
	}

	// Notify user about complete lack of Rust support
	m.notifyFallback(fmt.Sprintf("Rust parsing is completely unavailable: %s", reason))
}

// verifySynParser tests the syn parser with a simple Rust code sample
func (m *RustParserManager) verifySynParser(parser *RustSynParser) bool {
	testCode := `
fn test_function() -> i32 {
    42
}

struct TestStruct {
    field: i32,
}
`

	_, err := parser.ParseRustFile([]byte(testCode), "verification_test.rs")
	if err != nil {
		m.recordError(fmt.Errorf("syn parser verification failed: %w", err))
		return false
	}

	return true
}

// ParseRustFile parses a Rust file using the best available parser
func (m *RustParserManager) ParseRustFile(content []byte, filePath string) (*types.RustASTInfo, error) {
	atomic.AddInt64(&m.attemptCount, 1)

	m.mutex.RLock()
	parserType := m.currentParserType
	m.mutex.RUnlock()

	var result *types.RustASTInfo
	var err error

	switch parserType {
	case ParserTypeSyn:
		result, err = m.parseWithSyn(content, filePath)
		if err != nil {
			// Syn parser failed - attempt graceful degradation to regex
			if m.verbose {
				fmt.Printf("⚠️  Syn parser failed for %s, attempting regex fallback: %v\n", filePath, err)
			}
			result, err = m.parseWithRegexFallback(content, filePath)
		}

	case ParserTypeRegex:
		result, err = m.parseWithRegexFallback(content, filePath)

	case ParserTypeFallback:
		return nil, fmt.Errorf("Rust parsing is not available: %s", m.fallbackReason)

	default:
		return nil, fmt.Errorf("unknown parser type: %v", parserType)
	}

	if err == nil {
		atomic.AddInt64(&m.successCount, 1)
	} else {
		m.recordError(err)
	}

	return result, err
}

// parseWithSyn uses the syn crate parser
func (m *RustParserManager) parseWithSyn(content []byte, filePath string) (*types.RustASTInfo, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	if m.synParser == nil {
		return nil, fmt.Errorf("syn parser not initialized")
	}

	return m.synParser.ParseRustFile(content, filePath)
}

// parseWithRegexFallback uses the regex-based parser
func (m *RustParserManager) parseWithRegexFallback(content []byte, filePath string) (*types.RustASTInfo, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	if m.regexAnalyzer == nil {
		return nil, fmt.Errorf("regex analyzer not initialized")
	}

	return m.regexAnalyzer.parseWithRegexFallback(filePath, content)
}

// ValidateSyntax validates Rust syntax using the best available method
func (m *RustParserManager) ValidateSyntax(content []byte) (bool, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	parserType := m.currentParserType

	switch parserType {
	case ParserTypeSyn:
		if m.synParser == nil {
			return false, fmt.Errorf("syn parser not initialized")
		}
		return m.synParser.ValidateSyntax(content)

	case ParserTypeRegex:
		// Regex parser doesn't have syntax validation, so try basic parsing
		_, err := m.parseWithRegexFallback(content, "syntax_validation.rs")
		return err == nil, nil

	case ParserTypeFallback:
		return false, fmt.Errorf("syntax validation not available: %s", m.fallbackReason)

	default:
		return false, fmt.Errorf("unknown parser type: %v", parserType)
	}
}

// GetCapabilities returns information about the current parser's capabilities
func (m *RustParserManager) GetCapabilities() *RustParserCapabilities {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	capabilities := &RustParserCapabilities{
		ParserType:        m.currentParserType.String(),
		FallbackReason:    m.fallbackReason,
		InitializedAt:     m.initializationTime,
		AttemptCount:      atomic.LoadInt64(&m.attemptCount),
		SuccessCount:      atomic.LoadInt64(&m.successCount),
	}

	switch m.currentParserType {
	case ParserTypeSyn:
		if m.synParser != nil {
			if synCaps, err := m.synParser.GetCapabilities(); err == nil {
				capabilities.SynCapabilities = synCaps
			}
		}
		capabilities.HasSyntaxValidation = true
		capabilities.HasExpressionParsing = true
		capabilities.AccuracyLevel = "high"
		capabilities.PerformanceLevel = "optimal"

	case ParserTypeRegex:
		capabilities.HasSyntaxValidation = false
		capabilities.HasExpressionParsing = false
		capabilities.AccuracyLevel = "basic"
		capabilities.PerformanceLevel = "good"

	case ParserTypeFallback:
		capabilities.HasSyntaxValidation = false
		capabilities.HasExpressionParsing = false
		capabilities.AccuracyLevel = "none"
		capabilities.PerformanceLevel = "n/a"
	}

	return capabilities
}

// RustParserCapabilities contains information about parser capabilities
type RustParserCapabilities struct {
	ParserType           string                   `json:"parser_type"`
	FallbackReason       string                   `json:"fallback_reason,omitempty"`
	InitializedAt        time.Time                `json:"initialized_at"`
	AttemptCount         int64                    `json:"attempt_count"`
	SuccessCount         int64                    `json:"success_count"`
	HasSyntaxValidation  bool                     `json:"has_syntax_validation"`
	HasExpressionParsing bool                     `json:"has_expression_parsing"`
	AccuracyLevel        string                   `json:"accuracy_level"`
	PerformanceLevel     string                   `json:"performance_level"`
	SynCapabilities      *LibraryCapabilities     `json:"syn_capabilities,omitempty"`
	ErrorHistory         []string                 `json:"recent_errors,omitempty"`
}

// GetStatus returns detailed status information about the parser manager
func (m *RustParserManager) GetStatus() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	attemptCount := atomic.LoadInt64(&m.attemptCount)
	successCount := atomic.LoadInt64(&m.successCount)
	successRate := float64(0)
	if attemptCount > 0 {
		successRate = float64(successCount) / float64(attemptCount) * 100
	}

	recentErrors := make([]string, len(m.errorHistory))
	for i, err := range m.errorHistory {
		recentErrors[i] = err.Error()
	}

	return map[string]interface{}{
		"parser_type":        m.currentParserType.String(),
		"fallback_reason":    m.fallbackReason,
		"initialized_at":     m.initializationTime,
		"uptime_seconds":     time.Since(m.initializationTime).Seconds(),
		"attempt_count":      attemptCount,
		"success_count":      successCount,
		"success_rate":       successRate,
		"has_syn_parser":     m.synParser != nil,
		"has_regex_parser":   m.regexAnalyzer != nil,
		"recent_errors":      recentErrors,
		"cgo_enabled":        isCGOEnabled(),
		"rust_available":     m.currentParserType != ParserTypeFallback,
	}
}

// Cleanup cleans up resources used by the parser manager
func (m *RustParserManager) Cleanup() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.synParser != nil {
		m.synParser.Cleanup()
		m.synParser = nil
	}

	// Regex analyzer doesn't need explicit cleanup
	m.regexAnalyzer = nil

	m.currentParserType = ParserTypeFallback
	m.fallbackReason = "Manager has been cleaned up"
}

// recordError records an error in the error history
func (m *RustParserManager) recordError(err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Keep only the last 10 errors
	if len(m.errorHistory) >= 10 {
		m.errorHistory = m.errorHistory[1:]
	}
	m.errorHistory = append(m.errorHistory, err)
}

// notifyFallback notifies the user about fallback mode (only once)
func (m *RustParserManager) notifyFallback(message string) {
	if atomic.CompareAndSwapInt64(&m.fallbackNotified, 0, 1) {
		if m.verbose {
			fmt.Printf("ℹ️  %s\n", message)
			fmt.Println("   To enable full Rust support:")
			fmt.Println("   1. Install Rust toolchain: https://rustup.rs/")
			fmt.Println("   2. Build with: CGO_ENABLED=1 go build")
			fmt.Println("   3. Ensure goclean-rust-parser library is available")
		}
	}
}

// GetFallbackMessage returns a user-friendly message about the current state
func (m *RustParserManager) GetFallbackMessage() string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	switch m.currentParserType {
	case ParserTypeSyn:
		return "✅ Rust parsing: Full support with syn crate (optimal accuracy and performance)"

	case ParserTypeRegex:
		return "⚠️  Rust parsing: Limited support with regex fallback\n" +
			"   • Basic parsing capabilities available\n" +
			"   • Reduced accuracy compared to syn crate\n" +
			"   • Some advanced features unavailable"

	case ParserTypeFallback:
		return "❌ Rust parsing: Not available\n" +
			fmt.Sprintf("   Reason: %s\n", m.fallbackReason) +
			"   To enable Rust support:\n" +
			"   1. Install Rust toolchain (https://rustup.rs/)\n" +
			"   2. Build with CGO enabled: CGO_ENABLED=1 go build\n" +
			"   3. Ensure goclean-rust-parser library is built and available"

	default:
		return "❓ Rust parsing: Status unknown"
	}
}

// Helper functions for build-time detection

// isCGOEnabled checks if CGO is enabled at build time
func isCGOEnabled() bool {
	// This will be true if built with CGO enabled and the CGO build tag is present
	return CGOEnabled
}

// Build-time flags are defined in cgo_enabled.go and cgo_disabled.go

// Global parser manager instance
var (
	globalParserManager *RustParserManager
	globalManagerMutex  sync.RWMutex
)

// GetGlobalParserManager returns the global parser manager instance
func GetGlobalParserManager(verbose bool) *RustParserManager {
	// Fast path: check with read lock first
	globalManagerMutex.RLock()
	if globalParserManager != nil {
		manager := globalParserManager
		globalManagerMutex.RUnlock()
		return manager
	}
	globalManagerMutex.RUnlock()

	// Slow path: initialize with write lock
	globalManagerMutex.Lock()
	defer globalManagerMutex.Unlock()
	
	// Double-check in case another goroutine initialized it
	if globalParserManager == nil {
		globalParserManager = NewRustParserManager(verbose)
	}
	return globalParserManager
}

// CleanupGlobalParserManager cleans up the global parser manager
func CleanupGlobalParserManager() {
	globalManagerMutex.Lock()
	defer globalManagerMutex.Unlock()
	
	if globalParserManager != nil {
		globalParserManager.Cleanup()
		globalParserManager = nil
	}
}