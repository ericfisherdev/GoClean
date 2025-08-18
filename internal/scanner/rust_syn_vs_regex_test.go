package scanner

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ericfisherdev/goclean/internal/types"
)

// TestRustParsingAccuracy_SynVsRegex compares accuracy between syn and regex parsing
func TestRustParsingAccuracy_SynVsRegex(t *testing.T) {
	// Create both parsers
	regexAnalyzer := &RustASTAnalyzer{verbose: false, optimizer: nil}
	
	var synParser *RustSynParser
	var err error
	
	// Try to create syn parser, skip if not available
	synParser, err = NewRustSynParser(false)
	if err != nil {
		t.Skipf("Skipping accuracy comparison: Rust syn parser not available: %v", err)
	}
	defer func() {
		if synParser != nil {
			synParser.Cleanup()
		}
	}()

	testCases := []struct {
		name     string
		code     string
		expected AccuracyExpectations
	}{
		{
			name: "SimpleFunctionParsing",
			code: `
fn hello_world() -> String {
    "Hello, World!".to_string()
}

fn add_numbers(a: i32, b: i32) -> i32 {
    a + b
}
`,
			expected: AccuracyExpectations{
				Functions: 2,
				HasFunction: map[string]bool{
					"hello_world": true,
					"add_numbers": true,
				},
			},
		},
		{
			name: "StructParsing",
			code: `
pub struct User {
    pub name: String,
    pub age: u32,
    email: String,
}

struct Point {
    x: f64,
    y: f64,
}
`,
			expected: AccuracyExpectations{
				Structs: 2,
				HasStruct: map[string]bool{
					"User":  true,
					"Point": true,
				},
			},
		},
		{
			name: "ComplexStructWithImpl",
			code: `
pub struct Rectangle {
    pub width: f64,
    pub height: f64,
}

impl Rectangle {
    pub fn new(width: f64, height: f64) -> Self {
        Rectangle { width, height }
    }
    
    pub fn area(&self) -> f64 {
        self.width * self.height
    }
    
    fn perimeter(&self) -> f64 {
        2.0 * (self.width + self.height)
    }
}
`,
			expected: AccuracyExpectations{
				Structs:   1,
				Functions: 3, // new, area, perimeter
				Impls:     1,
				HasStruct: map[string]bool{"Rectangle": true},
				HasFunction: map[string]bool{
					"new":       true,
					"area":      true,
					"perimeter": true,
				},
			},
		},
		{
			name: "EnumParsing",
			code: `
pub enum Status {
    Active,
    Inactive,
    Pending,
}

enum Result<T, E> {
    Ok(T),
    Err(E),
}
`,
			expected: AccuracyExpectations{
				Enums: 2,
				HasEnum: map[string]bool{
					"Status": true,
					"Result": true,
				},
			},
		},
		{
			name: "TraitParsing",
			code: `
pub trait Display {
    fn fmt(&self) -> String;
}

trait Iterator {
    type Item;
    fn next(&mut self) -> Option<Self::Item>;
}
`,
			expected: AccuracyExpectations{
				Traits:    2,
				Functions: 2, // fmt, next
				HasTrait: map[string]bool{
					"Display":  true,
					"Iterator": true,
				},
			},
		},
		{
			name: "ComplexGenericCode",
			code: `
pub struct Container<T> 
where 
    T: Clone + Send + Sync,
{
    data: Vec<T>,
}

impl<T> Container<T> 
where 
    T: Clone + Send + Sync,
{
    pub fn new() -> Self {
        Self { data: Vec::new() }
    }
    
    pub fn add(&mut self, item: T) {
        self.data.push(item);
    }
}

pub trait Processable<T, U> {
    fn process(&self, input: T) -> U;
}
`,
			expected: AccuracyExpectations{
				Structs:   1,
				Traits:    1,
				Functions: 3, // new, add, process
				Impls:     1,
				HasStruct: map[string]bool{"Container": true},
				HasTrait:  map[string]bool{"Processable": true},
			},
		},
		{
			name: "AsyncAndUnsafeCode",
			code: `
pub async fn fetch_data(url: &str) -> Result<String, Box<dyn std::error::Error>> {
    let response = reqwest::get(url).await?;
    let content = response.text().await?;
    Ok(content)
}

pub unsafe fn raw_memory_access(ptr: *mut u8, len: usize) -> Vec<u8> {
    std::slice::from_raw_parts(ptr, len).to_vec()
}

const fn calculate_size(items: usize) -> usize {
    items * std::mem::size_of::<u64>()
}
`,
			expected: AccuracyExpectations{
				Functions: 3,
				HasFunction: map[string]bool{
					"fetch_data":         true,
					"raw_memory_access":  true,
					"calculate_size":     true,
				},
			},
		},
		{
			name: "MacroAndUseStatements",
			code: `
use std::collections::HashMap;
use std::sync::{Arc, Mutex};

macro_rules! create_getter {
    ($field:ident, $type:ty) => {
        pub fn $field(&self) -> &$type {
            &self.$field
        }
    };
}

pub struct Config {
    name: String,
    value: i32,
}

impl Config {
    create_getter!(name, String);
    create_getter!(value, i32);
}
`,
			expected: AccuracyExpectations{
				Structs:  1,
				Impls:    1,
				Uses:     2,
				Macros:   1,
				HasStruct: map[string]bool{"Config": true},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse with regex analyzer
			regexAST, regexErr := regexAnalyzer.AnalyzeRustFile(tc.name+".rs", []byte(tc.code))
			
			// Parse with syn parser
			synAST, synErr := synParser.ParseRustFile([]byte(tc.code), tc.name+".rs")

			// Compare results
			compareAccuracy(t, tc.name, tc.expected, regexAST, regexErr, synAST, synErr)
		})
	}
}

// AccuracyExpectations defines what we expect from parsing
type AccuracyExpectations struct {
	Functions   int
	Structs     int
	Enums       int
	Traits      int
	Impls       int
	Uses        int
	Macros      int
	Constants   int
	HasFunction map[string]bool
	HasStruct   map[string]bool
	HasEnum     map[string]bool
	HasTrait    map[string]bool
}

// compareAccuracy compares parsing results between regex and syn implementations
func compareAccuracy(t *testing.T, testName string, expected AccuracyExpectations, 
	regexAST *types.RustASTInfo, regexErr error,
	synAST *types.RustASTInfo, synErr error) {
	
	t.Logf("=== Accuracy Comparison for %s ===", testName)
	
	// Error comparison
	if regexErr != nil && synErr == nil {
		t.Errorf("Regex parser failed but syn parser succeeded: regex error: %v", regexErr)
	} else if regexErr == nil && synErr != nil {
		t.Errorf("Syn parser failed but regex parser succeeded: syn error: %v", synErr)
	} else if regexErr != nil && synErr != nil {
		t.Logf("Both parsers failed: regex: %v, syn: %v", regexErr, synErr)
		return
	}

	if regexAST == nil || synAST == nil {
		t.Error("One or both AST results are nil")
		return
	}

	// Compare counts
	compareCounts := []struct {
		name     string
		expected int
		regex    int
		syn      int
	}{
		{"Functions", expected.Functions, len(regexAST.Functions), len(synAST.Functions)},
		{"Structs", expected.Structs, len(regexAST.Structs), len(synAST.Structs)},
		{"Enums", expected.Enums, len(regexAST.Enums), len(synAST.Enums)},
		{"Traits", expected.Traits, len(regexAST.Traits), len(synAST.Traits)},
		{"Impls", expected.Impls, len(regexAST.Impls), len(synAST.Impls)},
		{"Uses", expected.Uses, len(regexAST.Uses), len(synAST.Uses)},
		{"Macros", expected.Macros, len(regexAST.Macros), len(synAST.Macros)},
		{"Constants", expected.Constants, len(regexAST.Constants), len(synAST.Constants)},
	}

	for _, comp := range compareCounts {
		if comp.expected > 0 {
			t.Logf("%s - Expected: %d, Regex: %d, Syn: %d", comp.name, comp.expected, comp.regex, comp.syn)
			
			// Check if syn is more accurate
			regexAccuracy := calculateAccuracy(comp.expected, comp.regex)
			synAccuracy := calculateAccuracy(comp.expected, comp.syn)
			
			if synAccuracy > regexAccuracy {
				t.Logf("  Syn parser more accurate: %.1f%% vs %.1f%%", synAccuracy*100, regexAccuracy*100)
			} else if regexAccuracy > synAccuracy {
				t.Logf("  Regex parser more accurate: %.1f%% vs %.1f%%", regexAccuracy*100, synAccuracy*100)
			} else {
				t.Logf("  Equal accuracy: %.1f%%", synAccuracy*100)
			}
		}
	}

	// Check specific items
	checkSpecificItems(t, "Functions", expected.HasFunction, regexAST.Functions, synAST.Functions)
	checkSpecificItems(t, "Structs", expected.HasStruct, 
		convertStructsToNames(regexAST.Structs), convertStructsToNames(synAST.Structs))
	checkSpecificItems(t, "Enums", expected.HasEnum,
		convertEnumsToNames(regexAST.Enums), convertEnumsToNames(synAST.Enums))
	checkSpecificItems(t, "Traits", expected.HasTrait,
		convertTraitsToNames(regexAST.Traits), convertTraitsToNames(synAST.Traits))
}

// calculateAccuracy returns accuracy as a ratio (0.0 to 1.0)
func calculateAccuracy(expected, actual int) float64 {
	if expected == 0 {
		return 1.0 // Perfect if we expected nothing
	}
	
	diff := float64(expected - actual)
	if diff < 0 {
		diff = -diff
	}
	
	return 1.0 - (diff / float64(expected))
}

// checkSpecificItems compares specific named items between parsers
func checkSpecificItems(t *testing.T, itemType string, expected map[string]bool, regexItems, synItems []string) {
	if len(expected) == 0 {
		return
	}

	t.Logf("%s item comparison:", itemType)
	
	regexSet := make(map[string]bool)
	for _, item := range regexItems {
		regexSet[item] = true
	}
	
	synSet := make(map[string]bool)
	for _, item := range synItems {
		synSet[item] = true
	}

	for expectedItem := range expected {
		regexFound := regexSet[expectedItem]
		synFound := synSet[expectedItem]
		
		t.Logf("  %s - Regex: %v, Syn: %v", expectedItem, regexFound, synFound)
		
		if !regexFound && !synFound {
			t.Errorf("Both parsers missed expected %s: %s", itemType, expectedItem)
		} else if !synFound && regexFound {
			t.Errorf("Syn parser missed %s that regex found: %s", itemType, expectedItem)
		}
	}
}

// Helper functions for converting AST elements to names
func convertStructsToNames(structs []*types.RustStructInfo) []string {
	names := make([]string, len(structs))
	for i, s := range structs {
		names[i] = s.Name
	}
	return names
}

func convertEnumsToNames(enums []*types.RustEnumInfo) []string {
	names := make([]string, len(enums))
	for i, e := range enums {
		names[i] = e.Name
	}
	return names
}

func convertTraitsToNames(traits []*types.RustTraitInfo) []string {
	names := make([]string, len(traits))
	for i, t := range traits {
		names[i] = t.Name
	}
	return names
}

// BenchmarkRustParsing_SynVsRegex benchmarks performance comparison
func BenchmarkRustParsing_SynVsRegex(b *testing.B) {
	// Test code samples of varying complexity
	testCodes := map[string]string{
		"Simple": `
fn hello() -> String {
    "Hello, World!".to_string()
}

struct Point { x: f64, y: f64 }
`,
		"Medium": `
use std::collections::HashMap;

pub struct DataProcessor {
    data: HashMap<String, i32>,
    count: usize,
}

impl DataProcessor {
    pub fn new() -> Self {
        Self {
            data: HashMap::new(),
            count: 0,
        }
    }
    
    pub fn process(&mut self, key: String, value: i32) -> Option<i32> {
        self.count += 1;
        self.data.insert(key, value)
    }
    
    pub fn get_count(&self) -> usize {
        self.count
    }
}

pub enum ProcessResult {
    Success(i32),
    Error(String),
}
`,
		"Complex": generateComplexRustCode(),
	}

	for testName, code := range testCodes {
		b.Run(fmt.Sprintf("Regex_%s", testName), func(b *testing.B) {
			benchmarkRegexParsing(b, code, testName)
		})
		
		b.Run(fmt.Sprintf("Syn_%s", testName), func(b *testing.B) {
			benchmarkSynParsing(b, code, testName)
		})
	}
}

func benchmarkRegexParsing(b *testing.B, code, testName string) {
	analyzer := &RustASTAnalyzer{verbose: false, optimizer: nil}
	codeBytes := []byte(code)
	fileName := testName + ".rs"
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		_, err := analyzer.AnalyzeRustFile(fileName, codeBytes)
		if err != nil {
			b.Fatalf("Regex parsing failed: %v", err)
		}
	}
}

func benchmarkSynParsing(b *testing.B, code, testName string) {
	parser, err := NewRustSynParser(false)
	if err != nil {
		b.Skipf("Skipping syn benchmark: %v", err)
	}
	defer parser.Cleanup()
	
	codeBytes := []byte(code)
	fileName := testName + ".rs"
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		_, err := parser.ParseRustFile(codeBytes, fileName)
		if err != nil {
			b.Fatalf("Syn parsing failed: %v", err)
		}
	}
}

// generateComplexRustCode creates a complex Rust code sample for benchmarking
func generateComplexRustCode() string {
	var builder strings.Builder
	
	builder.WriteString(`
use std::collections::{HashMap, HashSet, BTreeMap};
use std::sync::{Arc, Mutex, RwLock};
use std::thread;
use std::time::{Duration, Instant};

/// Complex generic trait with associated types
pub trait DataProcessor<T, U> 
where 
    T: Clone + Send + Sync,
    U: std::fmt::Debug,
{
    type Output;
    type Error: std::error::Error;
    
    fn process(&self, input: T) -> Result<Self::Output, Self::Error>;
    fn batch_process(&self, inputs: Vec<T>) -> Vec<Result<Self::Output, Self::Error>>;
}

/// Complex generic structure with lifetime parameters
pub struct ProcessingEngine<'a, T, U> 
where 
    T: Clone + Send + Sync + 'static,
    U: std::fmt::Debug + Send + Sync,
{
    name: &'a str,
    data_cache: Arc<RwLock<HashMap<String, T>>>,
    processors: Vec<Box<dyn DataProcessor<T, U, Output = U, Error = ProcessingError> + Send + Sync>>,
    metrics: ProcessingMetrics,
    config: ProcessingConfig,
}

#[derive(Debug, Clone)]
pub struct ProcessingMetrics {
    total_processed: Arc<Mutex<u64>>,
    errors_count: Arc<Mutex<u64>>,
    average_time: Arc<Mutex<Duration>>,
    start_time: Instant,
}

#[derive(Debug, Clone)]
pub struct ProcessingConfig {
    max_concurrent: usize,
    timeout: Duration,
    retry_count: u32,
    cache_size: usize,
}

#[derive(Debug, Clone)]
pub enum ProcessingError {
    InvalidInput(String),
    Timeout,
    CacheOverflow,
    ProcessorError(String),
    ThreadError(String),
}

impl std::fmt::Display for ProcessingError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            ProcessingError::InvalidInput(msg) => write!(f, "Invalid input: {}", msg),
            ProcessingError::Timeout => write!(f, "Processing timeout"),
            ProcessingError::CacheOverflow => write!(f, "Cache overflow"),
            ProcessingError::ProcessorError(msg) => write!(f, "Processor error: {}", msg),
            ProcessingError::ThreadError(msg) => write!(f, "Thread error: {}", msg),
        }
    }
}

impl std::error::Error for ProcessingError {}

impl<'a, T, U> ProcessingEngine<'a, T, U> 
where 
    T: Clone + Send + Sync + 'static,
    U: std::fmt::Debug + Send + Sync,
{
    pub fn new(name: &'a str, config: ProcessingConfig) -> Self {
        Self {
            name,
            data_cache: Arc::new(RwLock::new(HashMap::new())),
            processors: Vec::new(),
            metrics: ProcessingMetrics {
                total_processed: Arc::new(Mutex::new(0)),
                errors_count: Arc::new(Mutex::new(0)),
                average_time: Arc::new(Mutex::new(Duration::from_millis(0))),
                start_time: Instant::now(),
            },
            config,
        }
    }
    
    pub async fn process_async(&self, key: String, data: T) -> Result<U, ProcessingError> {
        let start_time = Instant::now();
        
        // Cache the data
        {
            let mut cache = self.data_cache.write().map_err(|_| ProcessingError::ThreadError("Cache lock error".to_string()))?;
            if cache.len() >= self.config.cache_size {
                return Err(ProcessingError::CacheOverflow);
            }
            cache.insert(key.clone(), data.clone());
        }
        
        // Process with timeout
        let result = tokio::time::timeout(
            self.config.timeout,
            self.process_with_retry(data)
        ).await;
        
        match result {
            Ok(Ok(output)) => {
                self.update_metrics(start_time, false);
                Ok(output)
            }
            Ok(Err(e)) => {
                self.update_metrics(start_time, true);
                Err(e)
            }
            Err(_) => {
                self.update_metrics(start_time, true);
                Err(ProcessingError::Timeout)
            }
        }
    }
    
    async fn process_with_retry(&self, data: T) -> Result<U, ProcessingError> {
        let mut last_error = ProcessingError::ProcessorError("No processors available".to_string());
        
        for attempt in 0..=self.config.retry_count {
            for processor in &self.processors {
                match processor.process(data.clone()) {
                    Ok(output) => return Ok(output),
                    Err(e) => {
                        last_error = ProcessingError::ProcessorError(format!("Attempt {}: {}", attempt, e));
                        if attempt < self.config.retry_count {
                            tokio::time::sleep(Duration::from_millis(100 * (attempt as u64 + 1))).await;
                        }
                    }
                }
            }
        }
        
        Err(last_error)
    }
    
    fn update_metrics(&self, start_time: Instant, is_error: bool) {
        let duration = start_time.elapsed();
        
        if let Ok(mut total) = self.metrics.total_processed.lock() {
            *total += 1;
        }
        
        if is_error {
            if let Ok(mut errors) = self.metrics.errors_count.lock() {
                *errors += 1;
            }
        }
        
        if let Ok(mut avg_time) = self.metrics.average_time.lock() {
            *avg_time = (*avg_time + duration) / 2;
        }
    }
    
    pub fn get_statistics(&self) -> Result<HashMap<String, u64>, ProcessingError> {
        let mut stats = HashMap::new();
        
        let total = self.metrics.total_processed.lock()
            .map_err(|_| ProcessingError::ThreadError("Lock error".to_string()))?;
        let errors = self.metrics.errors_count.lock()
            .map_err(|_| ProcessingError::ThreadError("Lock error".to_string()))?;
        let avg_time = self.metrics.average_time.lock()
            .map_err(|_| ProcessingError::ThreadError("Lock error".to_string()))?;
        
        stats.insert("total_processed".to_string(), *total);
        stats.insert("errors_count".to_string(), *errors);
        stats.insert("average_time_ms".to_string(), avg_time.as_millis() as u64);
        stats.insert("uptime_seconds".to_string(), self.metrics.start_time.elapsed().as_secs());
        
        Ok(stats)
    }
}

// Macro for creating specialized processors
macro_rules! create_processor {
    ($name:ident, $input_type:ty, $output_type:ty, $process_fn:expr) => {
        pub struct $name;
        
        impl DataProcessor<$input_type, $output_type> for $name {
            type Output = $output_type;
            type Error = ProcessingError;
            
            fn process(&self, input: $input_type) -> Result<Self::Output, Self::Error> {
                $process_fn(input)
            }
            
            fn batch_process(&self, inputs: Vec<$input_type>) -> Vec<Result<Self::Output, Self::Error>> {
                inputs.into_iter().map(|input| self.process(input)).collect()
            }
        }
    };
}

create_processor!(StringProcessor, String, String, |s: String| {
    if s.is_empty() {
        Err(ProcessingError::InvalidInput("Empty string".to_string()))
    } else {
        Ok(s.to_uppercase())
    }
});

create_processor!(NumberProcessor, i32, f64, |n: i32| {
    if n < 0 {
        Err(ProcessingError::InvalidInput("Negative number".to_string()))
    } else {
        Ok(n as f64 * 1.5)
    }
});

/// Const function for compile-time calculations
const fn calculate_buffer_size(items: usize, item_size: usize) -> usize {
    items * item_size + (items / 8) // Add some overhead
}

/// Static configuration with complex initialization
static GLOBAL_CONFIG: std::sync::OnceLock<ProcessingConfig> = std::sync::OnceLock::new();

pub fn get_global_config() -> &'static ProcessingConfig {
    GLOBAL_CONFIG.get_or_init(|| ProcessingConfig {
        max_concurrent: std::thread::available_parallelism().map(|n| n.get()).unwrap_or(4),
        timeout: Duration::from_secs(30),
        retry_count: 3,
        cache_size: calculate_buffer_size(1000, std::mem::size_of::<String>()),
    })
}
`)

	// Add more functions to increase complexity
	for i := 0; i < 20; i++ {
		builder.WriteString(fmt.Sprintf(`
pub fn generated_function_%d(param1: i32, param2: String) -> Result<String, ProcessingError> {
    if param1 < 0 {
        return Err(ProcessingError::InvalidInput(format!("Invalid param1: {}", param1)));
    }
    
    let processed = format!("Function_{}: {} -> {}", %d, param1, param2);
    
    match param1 %% 3 {
        0 => Ok(processed.to_uppercase()),
        1 => Ok(processed.to_lowercase()),
        _ => Ok(processed),
    }
}
`, i, i))
	}

	return builder.String()
}

// TestRustParsing_DetailedFeatureComparison compares specific language feature parsing
func TestRustParsing_DetailedFeatureComparison(t *testing.T) {
	regexAnalyzer := &RustASTAnalyzer{verbose: false, optimizer: nil}
	
	synParser, err := NewRustSynParser(false)
	if err != nil {
		t.Skipf("Skipping detailed comparison: Rust syn parser not available: %v", err)
	}
	defer synParser.Cleanup()

	features := []struct {
		name     string
		code     string
		checkFn  func(t *testing.T, regexAST, synAST *types.RustASTInfo)
	}{
		{
			name: "FunctionSignatures",
			code: `
fn simple() {}
fn with_params(a: i32, b: String) {}
fn with_return() -> i32 { 42 }
fn complex<T: Clone>(x: T, y: &str) -> Result<T, Box<dyn std::error::Error>> { Ok(x) }
async fn async_function() -> Result<(), String> { Ok(()) }
unsafe fn unsafe_function() {}
const fn const_function() -> i32 { 42 }
`,
			checkFn: func(t *testing.T, regexAST, synAST *types.RustASTInfo) {
				t.Logf("Function signature parsing comparison:")
				t.Logf("  Regex found %d functions", len(regexAST.Functions))
				t.Logf("  Syn found %d functions", len(synAST.Functions))
				
				expectedFunctions := []string{"simple", "with_params", "with_return", "complex", "async_function", "unsafe_function", "const_function"}
				
				for _, expected := range expectedFunctions {
					regexFound := findFunctionByName(regexAST.Functions, expected) != nil
					synFound := findFunctionByName(synAST.Functions, expected) != nil
					t.Logf("    %s: regex=%v, syn=%v", expected, regexFound, synFound)
				}
			},
		},
		{
			name: "GenericTypes",
			code: `
struct Generic<T> { data: T }
struct MultiGeneric<T, U> where T: Clone, U: Send { t: T, u: U }
enum GenericEnum<T> { Some(T), None }
trait GenericTrait<T> { fn method(&self, arg: T); }
`,
			checkFn: func(t *testing.T, regexAST, synAST *types.RustASTInfo) {
				t.Logf("Generic type parsing comparison:")
				t.Logf("  Regex: %d structs, %d enums, %d traits", 
					len(regexAST.Structs), len(regexAST.Enums), len(regexAST.Traits))
				t.Logf("  Syn: %d structs, %d enums, %d traits", 
					len(synAST.Structs), len(synAST.Enums), len(synAST.Traits))
			},
		},
		{
			name: "VisibilityModifiers",
			code: `
pub struct PublicStruct {}
struct PrivateStruct {}
pub(crate) struct CrateStruct {}
pub(super) struct SuperStruct {}

impl PublicStruct {
    pub fn public_method(&self) {}
    fn private_method(&self) {}
    pub(crate) fn crate_method(&self) {}
}
`,
			checkFn: func(t *testing.T, regexAST, synAST *types.RustASTInfo) {
				t.Logf("Visibility modifier parsing comparison:")
				
				// Count public vs private structs
				regexPublicStructs := 0
				synPublicStructs := 0
				
				for _, s := range regexAST.Structs {
					if s.IsPublic {
						regexPublicStructs++
					}
				}
				
				for _, s := range synAST.Structs {
					if s.IsPublic {
						synPublicStructs++
					}
				}
				
				t.Logf("  Public structs: regex=%d, syn=%d", regexPublicStructs, synPublicStructs)
			},
		},
	}

	for _, feature := range features {
		t.Run(feature.name, func(t *testing.T) {
			// Parse with both analyzers
			regexAST, regexErr := regexAnalyzer.AnalyzeRustFile(feature.name+".rs", []byte(feature.code))
			synAST, synErr := synParser.ParseRustFile([]byte(feature.code), feature.name+".rs")
			
			if regexErr != nil || synErr != nil {
				t.Errorf("Parsing errors: regex=%v, syn=%v", regexErr, synErr)
				return
			}
			
			// Run feature-specific comparison
			feature.checkFn(t, regexAST, synAST)
		})
	}
}

// BenchmarkRustParsing_MemoryUsage compares memory usage between parsers
func BenchmarkRustParsing_MemoryUsage(b *testing.B) {
	code := generateComplexRustCode()
	codeBytes := []byte(code)
	
	b.Run("RegexMemoryUsage", func(b *testing.B) {
		analyzer := &RustASTAnalyzer{verbose: false, optimizer: nil}
		
		b.ResetTimer()
		b.ReportAllocs()
		
		for i := 0; i < b.N; i++ {
			_, err := analyzer.AnalyzeRustFile("complex.rs", codeBytes)
			if err != nil {
				b.Fatalf("Regex parsing failed: %v", err)
			}
		}
	})
	
	b.Run("SynMemoryUsage", func(b *testing.B) {
		parser, err := NewRustSynParser(false)
		if err != nil {
			b.Skipf("Skipping syn memory benchmark: %v", err)
		}
		defer parser.Cleanup()
		
		b.ResetTimer()
		b.ReportAllocs()
		
		for i := 0; i < b.N; i++ {
			_, err := parser.ParseRustFile(codeBytes, "complex.rs")
			if err != nil {
				b.Fatalf("Syn parsing failed: %v", err)
			}
		}
	})
}