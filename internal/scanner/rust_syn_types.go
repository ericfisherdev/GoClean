// Package scanner provides common types for Rust syn parser implementations.
// This file defines types used by both CGO and fallback implementations.
package scanner

import "github.com/ericfisherdev/goclean/internal/types"

// ParseRequest represents a request to parse a Rust file
type ParseRequest struct {
	FilePath string `json:"file_path"`
	Content  string `json:"content"`
	Config   *RustSynConfig `json:"config,omitempty"`
}

// ParseResult represents the result of a parsing operation
type ParseResult struct {
	Success   bool               `json:"success"`
	AST       *types.RustASTInfo `json:"ast_info,omitempty"`
	Error     string             `json:"error,omitempty"`
	ErrorCode int                `json:"error_code,omitempty"`
	FilePath  string             `json:"file_path"`
}

// ParseStats represents statistics about parsing performance
type ParseStats struct {
	ParseTimeMs     int               `json:"parse_time_ms"`
	FileSize        int               `json:"file_size"`
	NodeCount       int               `json:"node_count"`
	ErrorCount      int               `json:"error_count"`
	WarningCount    int               `json:"warning_count"`
	ComplexityScore int               `json:"complexity_score"`
	MemoryUsageKB   int               `json:"memory_usage_kb"`
	Features        map[string]int    `json:"features"`
	Language        string            `json:"language"`
	ParserVersion   string            `json:"parser_version"`
	Success         bool              `json:"success"`
}

// SynParserCapabilities represents the capabilities of the Rust parsing library
type SynParserCapabilities struct {
	Version              string   `json:"version"`
	SupportedFeatures    []string `json:"supported_features"`
	MaxFileSize          int      `json:"max_file_size"`
	ThreadSafe           bool     `json:"thread_safe"`
	BatchProcessing      bool     `json:"batch_processing"`
	SyntaxValidation     bool     `json:"syntax_validation"`
	ExpressionParsing    bool     `json:"expression_parsing"`
	ErrorRecovery        bool     `json:"error_recovery"`
	IncrementalParsing   bool     `json:"incremental_parsing"`
	PositionTracking     bool     `json:"position_tracking"`
	CommentPreservation  bool     `json:"comment_preservation"`
	TokenStreaming       bool     `json:"token_streaming"`
	CustomConfiguration  bool     `json:"custom_configuration"`
}

// RustSynConfig represents configuration options for the Rust syn parser
type RustSynConfig struct {
	IncludeDocs       bool `json:"include_docs"`
	IncludePositions  bool `json:"include_positions"`
	ParseMacros       bool `json:"parse_macros"`
	IncludePrivate    bool `json:"include_private"`
	MaxComplexityCalc uint `json:"max_complexity_calc"`
	IncludeGenerics   bool `json:"include_generics"`
}

// DefaultRustSynConfig returns the default configuration for Rust parsing
func DefaultRustSynConfig() *RustSynConfig {
	return &RustSynConfig{
		IncludeDocs:       true,
		IncludePositions:  true,
		ParseMacros:       true,
		IncludePrivate:    true,
		MaxComplexityCalc: 100,
		IncludeGenerics:   true,
	}
}

// BatchParseResult represents the result of batch parsing multiple files
type BatchParseResult struct {
	TotalFiles       int                   `json:"total_files"`
	SuccessfulParses int                   `json:"successful_parses"`
	Results          []SingleParseResult   `json:"results"`
}

// SingleParseResult represents a single file parsing result in batch operation
type SingleParseResult struct {
	FilePath string             `json:"file_path"`
	Success  bool               `json:"success"`
	AST      *types.RustASTInfo `json:"ast_info,omitempty"`
	Error    string             `json:"error,omitempty"`
}

// ParsingStats represents statistics about parsing performance (legacy)
type ParsingStats struct {
	LineCount             int    `json:"line_count"`
	CharCount             int    `json:"char_count"`
	ByteCount             int    `json:"byte_count"`
	EstimatedFunctions    int    `json:"estimated_functions"`
	EstimatedStructs      int    `json:"estimated_structs"`
	EstimatedEnums        int    `json:"estimated_enums"`
	EstimatedTraits       int    `json:"estimated_traits"`
	EstimatedImpls        int    `json:"estimated_impls"`
	ParseComplexity       string `json:"parse_complexity"`
}

// LibraryCapabilities represents the capabilities of the Rust parsing library (legacy)
type LibraryCapabilities struct {
	Version              string                 `json:"version"`
	SynVersion           string                 `json:"syn_version"`
	Features             map[string]bool        `json:"features"`
	SupportedConstructs  []string               `json:"supported_constructs"`
	ParsingCapabilities  map[string]bool        `json:"parsing_capabilities"`
	Performance          map[string]bool        `json:"performance"`
}