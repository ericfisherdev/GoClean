//go:build cgo && !no_rust
// +build cgo,!no_rust

// Package scanner provides CGO bindings for the Rust syn crate parsing library.
// This file implements the C/Go interface for the goclean-rust-parser library.
package scanner

/*
#cgo LDFLAGS: -L../../lib -lgoclean_rust_parser -ldl -lm
#include <stdlib.h>
#include <string.h>

// C function declarations for the Rust library
extern int goclean_rust_init();
extern void goclean_rust_cleanup();
extern int goclean_rust_parse_file(const char* source, size_t source_len, const char* file_path, char** output, size_t* output_len);
extern int goclean_rust_validate_syntax(const char* source, size_t source_len);
extern int goclean_rust_parse_expression(const char* expr, size_t expr_len, char** output, size_t* output_len);
extern int goclean_rust_get_capabilities(char** capabilities, size_t* capabilities_len);
extern int goclean_rust_version(char** version, size_t* version_len);
extern void goclean_rust_free_string(char* ptr);

// Batch parsing and advanced features
extern int goclean_rust_batch_parse(const char** sources, const size_t* source_lens, const char** file_paths, size_t file_count, char** results, size_t* results_len);
extern int goclean_rust_get_parse_stats(const char* source, size_t source_len, char** stats_json, size_t* stats_len);
extern int goclean_rust_can_parse(const char* source, size_t source_len);

// FFI result structure for complex operations
struct FFIResult {
    int success;
    int error_code;
    char* error_message;
    char* data;
    size_t data_len;
};

// FFI configuration structure
struct FFIParseConfig {
    int include_docs;
    int include_positions;
    int parse_macros;
    int include_private;
    unsigned int max_complexity_calc;
    int include_generics;
};

extern int goclean_rust_parse_with_config(const char* source, size_t source_len, const char* file_path, const struct FFIParseConfig* config, struct FFIResult* result);
extern void goclean_rust_free_result(struct FFIResult* result);
*/
import "C"

import (
	"encoding/json"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/ericfisherdev/goclean/internal/types"
)

// RustSynParser provides CGO-based Rust parsing using the syn crate
type RustSynParser struct {
	initialized      bool
	mutex           sync.RWMutex
	verbose         bool
	
	// Memory safety tracking
	allocatedCStrings int64  // Track allocated C strings
	activeParseCalls  int64  // Track active parsing operations
	totalParseCalls   int64  // Total parsing operations
	lastCleanup       time.Time
	
	// Error tracking and recovery
	consecutiveErrors int64
	lastError         error
	errorMutex        sync.RWMutex
	
	// Performance monitoring
	totalParseTime    time.Duration
	avgParseTime      time.Duration
	perfMutex         sync.RWMutex
}


// NewRustSynParser creates a new Rust parser using the syn crate via CGO
func NewRustSynParser(verbose bool) (*RustSynParser, error) {
	parser := &RustSynParser{
		verbose: verbose,
	}

	if err := parser.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize Rust syn parser: %w", err)
	}

	return parser, nil
}

// Initialize initializes the Rust parsing library
func (p *RustSynParser) Initialize() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.initialized {
		return nil
	}

	// Reset safety counters
	atomic.StoreInt64(&p.allocatedCStrings, 0)
	atomic.StoreInt64(&p.activeParseCalls, 0)
	atomic.StoreInt64(&p.totalParseCalls, 0)
	atomic.StoreInt64(&p.consecutiveErrors, 0)
	p.lastCleanup = time.Now()

	result := C.goclean_rust_init()
	if result != 0 {
		p.recordError(fmt.Errorf("failed to initialize Rust library, error code: %d", result))
		return p.lastError
	}

	p.initialized = true
	if p.verbose {
		fmt.Println("Rust syn parser initialized successfully")
	}

	// Set up automatic cleanup routine
	go p.memoryCleanupRoutine()

	return nil
}

// Cleanup cleans up the Rust parsing library resources
func (p *RustSynParser) Cleanup() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.initialized {
		C.goclean_rust_cleanup()
		p.initialized = false
		if p.verbose {
			fmt.Println("Rust syn parser cleaned up")
		}
	}
}

// IsInitialized returns whether the parser has been initialized
func (p *RustSynParser) IsInitialized() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.initialized
}

// ParseRustFile parses a Rust source file and returns comprehensive AST information
func (p *RustSynParser) ParseRustFile(content []byte, filePath string) (*types.RustASTInfo, error) {
	return p.ParseRustFileWithConfig(content, filePath, DefaultRustSynConfig())
}

// ParseRustFileWithConfig parses a Rust source file with custom configuration
func (p *RustSynParser) ParseRustFileWithConfig(content []byte, filePath string, config *RustSynConfig) (*types.RustASTInfo, error) {
	startTime := time.Now()
	
	// Track active parsing operation
	atomic.AddInt64(&p.activeParseCalls, 1)
	atomic.AddInt64(&p.totalParseCalls, 1)
	defer atomic.AddInt64(&p.activeParseCalls, -1)
	
	// Check for too many consecutive errors
	if atomic.LoadInt64(&p.consecutiveErrors) > 10 {
		return nil, fmt.Errorf("too many consecutive errors, parser may be unstable")
	}

	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if !p.initialized {
		return nil, fmt.Errorf("Rust parser not initialized")
	}

	if p.verbose {
		fmt.Printf("Parsing Rust file with syn crate: %s (%d bytes)\n", filePath, len(content))
	}

	// Validate input size to prevent memory issues
	const maxFileSize = 50 * 1024 * 1024 // 50MB limit
	if len(content) > maxFileSize {
		return nil, fmt.Errorf("file too large: %d bytes (max %d bytes)", len(content), maxFileSize)
	}

	// Convert Go strings to C strings with tracking
	sourceC := p.allocateCString(string(content))
	defer p.freeCString(sourceC)

	filePathC := p.allocateCString(filePath)
	defer p.freeCString(filePathC)

	// Prepare config
	configC := C.struct_FFIParseConfig{
		include_docs:       boolToCInt(config.IncludeDocs),
		include_positions:  boolToCInt(config.IncludePositions),
		parse_macros:       boolToCInt(config.ParseMacros),
		include_private:    boolToCInt(config.IncludePrivate),
		max_complexity_calc: C.uint(config.MaxComplexityCalc),
		include_generics:   boolToCInt(config.IncludeGenerics),
	}

	// Prepare result structure
	var resultC C.struct_FFIResult

	// Call the Rust function
	result := C.goclean_rust_parse_with_config(
		sourceC,
		C.size_t(len(content)),
		filePathC,
		&configC,
		&resultC,
	)

	// Handle result with better error tracking
	if result != 0 {
		err := fmt.Errorf("parsing failed with error code: %d", result)
		p.recordError(err)
		return nil, err
	}

	defer func() {
		// Ensure cleanup happens even if panic occurs
		if r := recover(); r != nil {
			C.goclean_rust_free_result(&resultC)
			panic(r)
		} else {
			C.goclean_rust_free_result(&resultC)
		}
	}()

	if resultC.success == 0 {
		errorMsg := "unknown error"
		if resultC.error_message != nil {
			errorMsg = C.GoString(resultC.error_message)
		}
		err := fmt.Errorf("parsing failed: %s (error code: %d)", errorMsg, resultC.error_code)
		p.recordError(err)
		return nil, err
	}

	// Convert JSON result to Go struct
	if resultC.data == nil {
		err := fmt.Errorf("no data returned from parser")
		p.recordError(err)
		return nil, err
	}

	jsonData := C.GoStringN(resultC.data, C.int(resultC.data_len))
	
	// Parse the JSON into our Rust AST structure
	var rustAST RustASTFromJSON
	if err := json.Unmarshal([]byte(jsonData), &rustAST); err != nil {
		parseErr := fmt.Errorf("failed to parse JSON result: %w", err)
		p.recordError(parseErr)
		return nil, parseErr
	}

	// Convert to types.RustASTInfo
	astInfo := convertFromJSONAST(&rustAST)

	// Record successful parse
	p.recordSuccessfulParse(time.Since(startTime))

	if p.verbose {
		fmt.Printf("Successfully parsed %s: %d functions, %d structs, %d enums\n",
			filePath, len(astInfo.Functions), len(astInfo.Structs), len(astInfo.Enums))
	}

	return astInfo, nil
}

// ValidateSyntax checks if the Rust code has valid syntax without full parsing
func (p *RustSynParser) ValidateSyntax(content []byte) (bool, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if !p.initialized {
		return false, fmt.Errorf("Rust parser not initialized")
	}

	sourceC := p.allocateCString(string(content))
	defer p.freeCString(sourceC)

	result := C.goclean_rust_validate_syntax(sourceC, C.size_t(len(content)))
	
	switch result {
	case 1:
		return true, nil
	case 0:
		return false, nil
	default:
		err := fmt.Errorf("validation failed with error code: %d", result)
		p.recordError(err)
		return false, err
	}
}

// ParseExpression parses a single Rust expression and returns basic information
func (p *RustSynParser) ParseExpression(expression string) (map[string]interface{}, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if !p.initialized {
		return nil, fmt.Errorf("Rust parser not initialized")
	}

	exprC := p.allocateCString(expression)
	defer p.freeCString(exprC)

	var outputPtr *C.char
	var outputLen C.size_t

	result := C.goclean_rust_parse_expression(
		exprC,
		C.size_t(len(expression)),
		&outputPtr,
		&outputLen,
	)

	if result != 0 {
		err := fmt.Errorf("expression parsing failed with error code: %d", result)
		p.recordError(err)
		return nil, err
	}

	defer C.goclean_rust_free_string(outputPtr)

	jsonData := C.GoStringN(outputPtr, C.int(outputLen))

	var result_map map[string]interface{}
	if err := json.Unmarshal([]byte(jsonData), &result_map); err != nil {
		parseErr := fmt.Errorf("failed to parse expression result: %w", err)
		p.recordError(parseErr)
		return nil, parseErr
	}

	return result_map, nil
}

// GetCapabilities returns information about the library's capabilities
func (p *RustSynParser) GetCapabilities() (*LibraryCapabilities, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if !p.initialized {
		return nil, fmt.Errorf("Rust parser not initialized")
	}

	var capabilitiesPtr *C.char
	var capabilitiesLen C.size_t

	result := C.goclean_rust_get_capabilities(&capabilitiesPtr, &capabilitiesLen)
	if result != 0 {
		return nil, fmt.Errorf("failed to get capabilities, error code: %d", result)
	}

	defer C.goclean_rust_free_string(capabilitiesPtr)

	jsonData := C.GoStringN(capabilitiesPtr, C.int(capabilitiesLen))

	var capabilities LibraryCapabilities
	if err := json.Unmarshal([]byte(jsonData), &capabilities); err != nil {
		return nil, fmt.Errorf("failed to parse capabilities: %w", err)
	}

	return &capabilities, nil
}

// GetVersion returns version information about the library
func (p *RustSynParser) GetVersion() (map[string]interface{}, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if !p.initialized {
		return nil, fmt.Errorf("Rust parser not initialized")
	}

	var versionPtr *C.char
	var versionLen C.size_t

	result := C.goclean_rust_version(&versionPtr, &versionLen)
	if result != 0 {
		return nil, fmt.Errorf("failed to get version, error code: %d", result)
	}

	defer C.goclean_rust_free_string(versionPtr)

	jsonData := C.GoStringN(versionPtr, C.int(versionLen))

	var version map[string]interface{}
	if err := json.Unmarshal([]byte(jsonData), &version); err != nil {
		return nil, fmt.Errorf("failed to parse version: %w", err)
	}

	return version, nil
}

// GetParseStats returns parsing statistics for the given source code
func (p *RustSynParser) GetParseStats(content []byte) (*ParsingStats, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if !p.initialized {
		return nil, fmt.Errorf("Rust parser not initialized")
	}

        sourceC := p.allocateCString(string(content))
        defer p.freeCString(sourceC)

	var statsPtr *C.char
	var statsLen C.size_t

	result := C.goclean_rust_get_parse_stats(
		sourceC,
		C.size_t(len(content)),
		&statsPtr,
		&statsLen,
	)

	if result != 0 {
		return nil, fmt.Errorf("failed to get parse stats, error code: %d", result)
	}

	defer C.goclean_rust_free_string(statsPtr)

	jsonData := C.GoStringN(statsPtr, C.int(statsLen))

	var stats ParsingStats
	if err := json.Unmarshal([]byte(jsonData), &stats); err != nil {
		return nil, fmt.Errorf("failed to parse stats: %w", err)
	}

	return &stats, nil
}

// CanParse checks if the content looks like Rust code that can be parsed
func (p *RustSynParser) CanParse(content []byte) (int, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if !p.initialized {
		return -1, fmt.Errorf("Rust parser not initialized")
	}

	sourceC := p.allocateCString(string(content))
	defer p.freeCString(sourceC)

	result := C.goclean_rust_can_parse(sourceC, C.size_t(len(content)))
	return int(result), nil
}

// Helper function to convert Go bool to C int
func boolToCInt(b bool) C.int {
	if b {
		return 1
	}
	return 0
}

// Global parser instance (initialized once)
var (
	globalSynParser *RustSynParser
	globalSynMutex  sync.Once
)

// GetGlobalSynParser returns a singleton instance of the Rust syn parser
func GetGlobalSynParser() (*RustSynParser, error) {
	var err error
	globalSynMutex.Do(func() {
		globalSynParser, err = NewRustSynParser(false)
	})
	return globalSynParser, err
}

// CleanupGlobalSynParser cleans up the global parser instance
func CleanupGlobalSynParser() {
	if globalSynParser != nil {
		globalSynParser.Cleanup()
		globalSynParser = nil
	}
}

// RustASTFromJSON represents the JSON structure returned by the Rust library
type RustASTFromJSON struct {
	FilePath  string                  `json:"file_path"`
	Module    RustModuleFromJSON      `json:"module"`
	Functions []RustFunctionFromJSON  `json:"functions"`
	Structs   []RustStructFromJSON    `json:"structs"`
	Enums     []RustEnumFromJSON      `json:"enums"`
	Traits    []RustTraitFromJSON     `json:"traits"`
	Impls     []RustImplFromJSON      `json:"impls"`
	Constants []RustConstantFromJSON  `json:"constants"`
	Statics   []RustStaticFromJSON    `json:"statics"`
	Uses      []RustUseFromJSON       `json:"uses"`
	Mods      []RustModFromJSON       `json:"mods"`
	Macros    []RustMacroFromJSON     `json:"macros"`
}

// RustModuleFromJSON represents module information from JSON
type RustModuleFromJSON struct {
	Name       string `json:"name"`
	IsPublic   bool   `json:"is_public"`
	Visibility string `json:"visibility"`
}

// RustFunctionFromJSON represents function information from JSON
type RustFunctionFromJSON struct {
	Name           string                      `json:"name"`
	StartLine      int                         `json:"start_line"`
	EndLine        int                         `json:"end_line"`
	StartColumn    int                         `json:"start_column"`
	EndColumn      int                         `json:"end_column"`
	Parameters     []RustParameterFromJSON     `json:"parameters"`
	ReturnType     string                      `json:"return_type"`
	IsPublic       bool                        `json:"is_public"`
	IsAsync        bool                        `json:"is_async"`
	IsUnsafe       bool                        `json:"is_unsafe"`
	IsConst        bool                        `json:"is_const"`
	Complexity     int                         `json:"complexity"`
	LineCount      int                         `json:"line_count"`
	HasDocComments bool                        `json:"has_doc_comments"`
	Visibility     string                      `json:"visibility"`
	Generics       []RustGenericFromJSON       `json:"generics"`
}

// RustParameterFromJSON represents parameter information from JSON
type RustParameterFromJSON struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	IsMutable bool   `json:"is_mutable"`
	IsRef     bool   `json:"is_ref"`
}

// RustGenericFromJSON represents generic parameter information from JSON
type RustGenericFromJSON struct {
	Name        string   `json:"name"`
	Bounds      []string `json:"bounds"`
	DefaultType string   `json:"default_type"`
}

// RustStructFromJSON represents struct information from JSON
type RustStructFromJSON struct {
	Name           string                   `json:"name"`
	StartLine      int                      `json:"start_line"`
	EndLine        int                      `json:"end_line"`
	StartColumn    int                      `json:"start_column"`
	EndColumn      int                      `json:"end_column"`
	IsPublic       bool                     `json:"is_public"`
	FieldCount     int                      `json:"field_count"`
	Visibility     string                   `json:"visibility"`
	HasDocComments bool                     `json:"has_doc_comments"`
	Fields         []RustFieldFromJSON      `json:"fields"`
	Generics       []RustGenericFromJSON    `json:"generics"`
}

// RustFieldFromJSON represents struct field information from JSON
type RustFieldFromJSON struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	IsPublic   bool   `json:"is_public"`
	Visibility string `json:"visibility"`
}

// RustEnumFromJSON represents enum information from JSON
type RustEnumFromJSON struct {
	Name           string                   `json:"name"`
	StartLine      int                      `json:"start_line"`
	EndLine        int                      `json:"end_line"`
	StartColumn    int                      `json:"start_column"`
	EndColumn      int                      `json:"end_column"`
	IsPublic       bool                     `json:"is_public"`
	VariantCount   int                      `json:"variant_count"`
	Visibility     string                   `json:"visibility"`
	HasDocComments bool                     `json:"has_doc_comments"`
	Variants       []RustVariantFromJSON    `json:"variants"`
	Generics       []RustGenericFromJSON    `json:"generics"`
}

// RustVariantFromJSON represents enum variant information from JSON
type RustVariantFromJSON struct {
	Name      string                    `json:"name"`
	Fields    []RustFieldFromJSON       `json:"fields"`
	HasFields bool                      `json:"has_fields"`
}

// RustTraitFromJSON represents trait information from JSON
type RustTraitFromJSON struct {
	Name           string                   `json:"name"`
	StartLine      int                      `json:"start_line"`
	EndLine        int                      `json:"end_line"`
	StartColumn    int                      `json:"start_column"`
	EndColumn      int                      `json:"end_column"`
	IsPublic       bool                     `json:"is_public"`
	MethodCount    int                      `json:"method_count"`
	Visibility     string                   `json:"visibility"`
	HasDocComments bool                     `json:"has_doc_comments"`
	Methods        []RustFunctionFromJSON   `json:"methods"`
	Generics       []RustGenericFromJSON    `json:"generics"`
}

// RustImplFromJSON represents impl block information from JSON
type RustImplFromJSON struct {
	StartLine   int                      `json:"start_line"`
	EndLine     int                      `json:"end_line"`
	StartColumn int                      `json:"start_column"`
	EndColumn   int                      `json:"end_column"`
	TargetType  string                   `json:"target_type"`
	TraitName   string                   `json:"trait_name"`
	MethodCount int                      `json:"method_count"`
	Methods     []RustFunctionFromJSON   `json:"methods"`
	Generics    []RustGenericFromJSON    `json:"generics"`
}

// RustConstantFromJSON represents constant information from JSON
type RustConstantFromJSON struct {
	Name           string `json:"name"`
	Type           string `json:"type"`
	Value          string `json:"value"`
	StartLine      int    `json:"start_line"`
	EndLine        int    `json:"end_line"`
	StartColumn    int    `json:"start_column"`
	EndColumn      int    `json:"end_column"`
	IsPublic       bool   `json:"is_public"`
	Visibility     string `json:"visibility"`
	HasDocComments bool   `json:"has_doc_comments"`
}

// RustStaticFromJSON represents static variable information from JSON
type RustStaticFromJSON struct {
	Name           string `json:"name"`
	Type           string `json:"type"`
	IsMutable      bool   `json:"is_mutable"`
	StartLine      int    `json:"start_line"`
	EndLine        int    `json:"end_line"`
	StartColumn    int    `json:"start_column"`
	EndColumn      int    `json:"end_column"`
	IsPublic       bool   `json:"is_public"`
	Visibility     string `json:"visibility"`
	HasDocComments bool   `json:"has_doc_comments"`
}

// RustUseFromJSON represents use statement information from JSON
type RustUseFromJSON struct {
	Path           string `json:"path"`
	IsPublic       bool   `json:"is_public"`
	Visibility     string `json:"visibility"`
	StartLine      int    `json:"start_line"`
	EndLine        int    `json:"end_line"`
	StartColumn    int    `json:"start_column"`
	EndColumn      int    `json:"end_column"`
}

// RustModFromJSON represents module declaration information from JSON
type RustModFromJSON struct {
	Name           string `json:"name"`
	IsPublic       bool   `json:"is_public"`
	Visibility     string `json:"visibility"`
	StartLine      int    `json:"start_line"`
	EndLine        int    `json:"end_line"`
	StartColumn    int    `json:"start_column"`
	EndColumn      int    `json:"end_column"`
	HasDocComments bool   `json:"has_doc_comments"`
}

// RustMacroFromJSON represents macro information from JSON
type RustMacroFromJSON struct {
	Name           string `json:"name"`
	StartLine      int    `json:"start_line"`
	EndLine        int    `json:"end_line"`
	StartColumn    int    `json:"start_column"`
	EndColumn      int    `json:"end_column"`
	IsPublic       bool   `json:"is_public"`
	Visibility     string `json:"visibility"`
	HasDocComments bool   `json:"has_doc_comments"`
}

// convertFromJSONAST converts the JSON AST structure to types.RustASTInfo
func convertFromJSONAST(jsonAST *RustASTFromJSON) *types.RustASTInfo {
	result := &types.RustASTInfo{
		FilePath:  jsonAST.FilePath,
		CrateName: jsonAST.Module.Name,
		Functions: make([]*types.RustFunctionInfo, 0, len(jsonAST.Functions)),
		Structs:   make([]*types.RustStructInfo, 0, len(jsonAST.Structs)),
		Enums:     make([]*types.RustEnumInfo, 0, len(jsonAST.Enums)),
		Traits:    make([]*types.RustTraitInfo, 0, len(jsonAST.Traits)),
		Impls:     make([]*types.RustImplInfo, 0, len(jsonAST.Impls)),
		Modules:   make([]*types.RustModuleInfo, 0, len(jsonAST.Mods)),
		Constants: make([]*types.RustConstantInfo, 0, len(jsonAST.Constants)),
		Uses:      make([]*types.RustUseInfo, 0, len(jsonAST.Uses)),
		Macros:    make([]*types.RustMacroInfo, 0, len(jsonAST.Macros)),
	}

	// Convert functions
	for _, fn := range jsonAST.Functions {
		result.Functions = append(result.Functions, convertFunctionFromJSON(&fn))
	}

	// Convert structs
	for _, st := range jsonAST.Structs {
		result.Structs = append(result.Structs, convertStructFromJSON(&st))
	}

	// Convert enums
	for _, en := range jsonAST.Enums {
		result.Enums = append(result.Enums, convertEnumFromJSON(&en))
	}

	// Convert traits
	for _, tr := range jsonAST.Traits {
		result.Traits = append(result.Traits, convertTraitFromJSON(&tr))
	}

	// Convert impls
	for _, impl := range jsonAST.Impls {
		result.Impls = append(result.Impls, convertImplFromJSON(&impl))
	}

	// Convert modules
	for _, m := range jsonAST.Mods {
		result.Modules = append(result.Modules, convertModDeclFromJSON(&m))
	}

	// Convert constants
	for _, c := range jsonAST.Constants {
		result.Constants = append(result.Constants, convertConstantFromJSON(&c))
	}

	// Convert uses
	for _, u := range jsonAST.Uses {
		result.Uses = append(result.Uses, convertUseFromJSON(&u))
	}

	// Convert macros
	for _, mac := range jsonAST.Macros {
		result.Macros = append(result.Macros, convertMacroFromJSON(&mac))
	}

	return result
}

// Helper conversion functions
// Note: Module information is extracted directly as CrateName in main conversion

func convertFunctionFromJSON(f *RustFunctionFromJSON) *types.RustFunctionInfo {
	result := &types.RustFunctionInfo{
		Name:           f.Name,
		StartLine:      f.StartLine,
		EndLine:        f.EndLine,
		StartColumn:    f.StartColumn,
		EndColumn:      f.EndColumn,
		ReturnType:     f.ReturnType,
		IsPublic:       f.IsPublic,
		IsAsync:        f.IsAsync,
		IsUnsafe:       f.IsUnsafe,
		IsConst:        f.IsConst,
		Complexity:     f.Complexity,
		LineCount:      f.LineCount,
		HasDocComments: f.HasDocComments,
		Visibility:     f.Visibility,
		Parameters:     make([]types.RustParameterInfo, 0, len(f.Parameters)),
	}

	// Convert parameters
	for _, p := range f.Parameters {
		result.Parameters = append(result.Parameters, types.RustParameterInfo{
			Name:      p.Name,
			Type:      p.Type,
			IsMutable: p.IsMutable,
			IsRef:     p.IsRef,
		})
	}

	return result
}

func convertStructFromJSON(s *RustStructFromJSON) *types.RustStructInfo {
	return &types.RustStructInfo{
		Name:           s.Name,
		StartLine:      s.StartLine,
		EndLine:        s.EndLine,
		StartColumn:    s.StartColumn,
		EndColumn:      s.EndColumn,
		IsPublic:       s.IsPublic,
		FieldCount:     s.FieldCount,
		Visibility:     s.Visibility,
		HasDocComments: s.HasDocComments,
	}
}

func convertEnumFromJSON(e *RustEnumFromJSON) *types.RustEnumInfo {
	return &types.RustEnumInfo{
		Name:           e.Name,
		StartLine:      e.StartLine,
		EndLine:        e.EndLine,
		StartColumn:    e.StartColumn,
		EndColumn:      e.EndColumn,
		IsPublic:       e.IsPublic,
		VariantCount:   e.VariantCount,
		Visibility:     e.Visibility,
		HasDocComments: e.HasDocComments,
	}
}

func convertTraitFromJSON(t *RustTraitFromJSON) *types.RustTraitInfo {
	return &types.RustTraitInfo{
		Name:           t.Name,
		StartLine:      t.StartLine,
		EndLine:        t.EndLine,
		StartColumn:    t.StartColumn,
		EndColumn:      t.EndColumn,
		IsPublic:       t.IsPublic,
		MethodCount:    t.MethodCount,
		Visibility:     t.Visibility,
		HasDocComments: t.HasDocComments,
	}
}

func convertImplFromJSON(i *RustImplFromJSON) *types.RustImplInfo {
	return &types.RustImplInfo{
		StartLine:   i.StartLine,
		EndLine:     i.EndLine,
		StartColumn: i.StartColumn,
		EndColumn:   i.EndColumn,
		TargetType:  i.TargetType,
		TraitName:   i.TraitName,
		MethodCount: i.MethodCount,
	}
}

func convertConstantFromJSON(c *RustConstantFromJSON) *types.RustConstantInfo {
	return &types.RustConstantInfo{
		Name:           c.Name,
		Type:           c.Type,
		Line:           c.StartLine,
		Column:         c.StartColumn,
		IsPublic:       c.IsPublic,
		Visibility:     c.Visibility,
		HasDocComments: c.HasDocComments,
	}
}

// Note: RustStaticInfo doesn't exist in types package
// Statics are handled as constants or skipped for now

func convertUseFromJSON(u *RustUseFromJSON) *types.RustUseInfo {
	return &types.RustUseInfo{
		Path:       u.Path,
		Alias:      "", // Not provided in JSON, could be extracted if needed
		Line:       u.StartLine,
		Column:     u.StartColumn,
		Visibility: u.Visibility,
	}
}

func convertModDeclFromJSON(m *RustModFromJSON) *types.RustModuleInfo {
	return &types.RustModuleInfo{
		Name:           m.Name,
		StartLine:      m.StartLine,
		EndLine:        m.EndLine,
		StartColumn:    m.StartColumn,
		EndColumn:      m.EndColumn,
		IsPublic:       m.IsPublic,
		Visibility:     m.Visibility,
		HasDocComments: m.HasDocComments,
	}
}

func convertMacroFromJSON(m *RustMacroFromJSON) *types.RustMacroInfo {
	return &types.RustMacroInfo{
		Name:           m.Name,
		StartLine:      m.StartLine,
		EndLine:        m.EndLine,
		StartColumn:    m.StartColumn,
		EndColumn:      m.EndColumn,
		IsPublic:       m.IsPublic,
		MacroType:      "macro_rules!", // Default type, could be enhanced
		HasDocComments: m.HasDocComments,
	}
}

// Memory safety helper methods

// allocateCString allocates a C string and tracks it for safety
func (p *RustSynParser) allocateCString(s string) *C.char {
	atomic.AddInt64(&p.allocatedCStrings, 1)
	return C.CString(s)
}

// freeCString frees a C string and updates tracking
func (p *RustSynParser) freeCString(cstr *C.char) {
	if cstr != nil {
		C.free(unsafe.Pointer(cstr))
		atomic.AddInt64(&p.allocatedCStrings, -1)
	}
}

// recordError records a parsing error for tracking
func (p *RustSynParser) recordError(err error) {
	p.errorMutex.Lock()
	defer p.errorMutex.Unlock()
	
	p.lastError = err
	atomic.AddInt64(&p.consecutiveErrors, 1)
}

// recordSuccessfulParse records a successful parsing operation
func (p *RustSynParser) recordSuccessfulParse(duration time.Duration) {
	// Reset consecutive errors on success
	atomic.StoreInt64(&p.consecutiveErrors, 0)
	
	// Update performance metrics
	p.perfMutex.Lock()
	defer p.perfMutex.Unlock()
	
	p.totalParseTime += duration
	totalCalls := atomic.LoadInt64(&p.totalParseCalls)
	if totalCalls > 0 {
		p.avgParseTime = p.totalParseTime / time.Duration(totalCalls)
	}
}

// memoryCleanupRoutine performs periodic memory safety checks
func (p *RustSynParser) memoryCleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		p.mutex.RLock()
		initialized := p.initialized
		p.mutex.RUnlock()
		
		if !initialized {
			return
		}
		
		// Check for memory leaks
		allocatedStrings := atomic.LoadInt64(&p.allocatedCStrings)
		activeCalls := atomic.LoadInt64(&p.activeParseCalls)
		
		if p.verbose && (allocatedStrings > 0 || activeCalls > 0) {
			fmt.Printf("Memory safety check: %d allocated C strings, %d active parse calls\n", 
				allocatedStrings, activeCalls)
		}
		
		// Force garbage collection if memory usage seems high
		if allocatedStrings > 100 {
			runtime.GC()
		}
		
		p.lastCleanup = time.Now()
	}
}

// GetMemorySafetyMetrics returns current memory safety metrics
func (p *RustSynParser) GetMemorySafetyMetrics() map[string]interface{} {
	p.perfMutex.RLock()
	defer p.perfMutex.RUnlock()
	
	p.errorMutex.RLock()
	defer p.errorMutex.RUnlock()
	
	var lastErrorMsg string
	if p.lastError != nil {
		lastErrorMsg = p.lastError.Error()
	}
	
	return map[string]interface{}{
		"allocated_c_strings":    atomic.LoadInt64(&p.allocatedCStrings),
		"active_parse_calls":     atomic.LoadInt64(&p.activeParseCalls),
		"total_parse_calls":      atomic.LoadInt64(&p.totalParseCalls),
		"consecutive_errors":     atomic.LoadInt64(&p.consecutiveErrors),
		"last_error":             lastErrorMsg,
		"avg_parse_time_ms":      p.avgParseTime.Milliseconds(),
		"total_parse_time_ms":    p.totalParseTime.Milliseconds(),
		"last_cleanup":           p.lastCleanup.Format(time.RFC3339),
		"initialized":            p.initialized,
	}
}

// ForceCleanup forces immediate cleanup of resources
func (p *RustSynParser) ForceCleanup() {
	runtime.GC()
	p.lastCleanup = time.Now()
	
	if p.verbose {
		metrics := p.GetMemorySafetyMetrics()
		fmt.Printf("Forced cleanup completed. Metrics: %+v\n", metrics)
	}
}

// ValidateMemoryState checks for potential memory issues
func (p *RustSynParser) ValidateMemoryState() error {
	allocatedStrings := atomic.LoadInt64(&p.allocatedCStrings)
	activeCalls := atomic.LoadInt64(&p.activeParseCalls)
	consecutiveErrors := atomic.LoadInt64(&p.consecutiveErrors)
	
	// Check for obvious memory leaks
	if allocatedStrings > activeCalls*2 {
		return fmt.Errorf("potential memory leak: %d allocated C strings with only %d active calls",
			allocatedStrings, activeCalls)
	}
	
	// Check for excessive errors
	if consecutiveErrors > 20 {
		return fmt.Errorf("too many consecutive errors (%d), parser may be unstable",
			consecutiveErrors)
	}
	
	// Check for extremely long-running operations
	if activeCalls > 0 && time.Since(p.lastCleanup) > 30*time.Minute {
		return fmt.Errorf("parsing operations running for over 30 minutes, possible deadlock")
	}
	
	return nil
}