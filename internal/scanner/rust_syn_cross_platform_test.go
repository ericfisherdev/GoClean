//go:build cgo && !no_rust
// +build cgo,!no_rust

package scanner

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// TestRustSynParser_CrossPlatformPaths tests handling of different path formats
func TestRustSynParser_CrossPlatformPaths(t *testing.T) {
	parser, err := NewRustSynParser(false)
	if err != nil {
		t.Skipf("Skipping test: Rust syn parser not available: %v", err)
	}
	defer parser.Cleanup()

	rustCode := `
fn cross_platform_test() {
    println!("Testing cross-platform paths");
}
`

	testPaths := []string{
		"simple.rs",
		"src/main.rs",
		"deep/nested/path/file.rs",
		"with-hyphens/test-file.rs",
		"with_underscores/test_file.rs",
		"123numeric/start123.rs",
	}

	// Add platform-specific paths
	switch runtime.GOOS {
	case "windows":
		testPaths = append(testPaths, 
			"C:\\Users\\test\\project\\main.rs",
			"D:\\development\\rust\\lib.rs",
			"\\\\server\\share\\code.rs",
		)
	case "linux", "darwin":
		testPaths = append(testPaths,
			"/home/user/project/main.rs",
			"/usr/local/src/rust/lib.rs",
			"/tmp/test.rs",
		)
	}

	for _, path := range testPaths {
		t.Run(fmt.Sprintf("Path_%s", strings.ReplaceAll(path, "/", "_")), func(t *testing.T) {
			astInfo, err := parser.ParseRustFile([]byte(rustCode), path)
			if err != nil {
				t.Errorf("Failed to parse with path %s: %v", path, err)
				return
			}

			if astInfo.FilePath != path {
				t.Errorf("File path mismatch: expected %s, got %s", path, astInfo.FilePath)
			}
		})
	}
}

// TestRustSynParser_UnicodePaths tests handling of Unicode in file paths
func TestRustSynParser_UnicodePaths(t *testing.T) {
	parser, err := NewRustSynParser(false)
	if err != nil {
		t.Skipf("Skipping test: Rust syn parser not available: %v", err)
	}
	defer parser.Cleanup()

	rustCode := `
fn unicode_test() {
    println!("Testing Unicode paths");
}
`

	unicodePaths := []string{
		"测试/文件.rs",           // Chinese
		"тест/файл.rs",          // Russian
		"テスト/ファイル.rs",         // Japanese
		"🦀/crab_code.rs",       // Emoji
		"café/élève.rs",         // Accented characters
		"español/niño.rs",       // Spanish
		"Ελληνικά/δοκιμή.rs",    // Greek
	}

	for _, path := range unicodePaths {
		t.Run(fmt.Sprintf("Unicode_%s", path), func(t *testing.T) {
			astInfo, err := parser.ParseRustFile([]byte(rustCode), path)
			if err != nil {
				t.Errorf("Failed to parse with Unicode path %s: %v", path, err)
				return
			}

			if astInfo.FilePath != path {
				t.Errorf("Unicode path mismatch: expected %s, got %s", path, astInfo.FilePath)
			}
		})
	}
}

// TestRustSynParser_PlatformSpecificCode tests parsing platform-specific Rust code
func TestRustSynParser_PlatformSpecificCode(t *testing.T) {
	parser, err := NewRustSynParser(false)
	if err != nil {
		t.Skipf("Skipping test: Rust syn parser not available: %v", err)
	}
	defer parser.Cleanup()

	// Test code with platform-specific attributes
	platformCode := `
#[cfg(target_os = "windows")]
fn windows_specific() {
    use std::os::windows::ffi::OsStringExt;
    println!("Windows-specific code");
}

#[cfg(target_os = "linux")]
fn linux_specific() {
    use std::os::unix::fs::PermissionsExt;
    println!("Linux-specific code");
}

#[cfg(target_os = "macos")]
fn macos_specific() {
    use std::os::unix::net::UnixStream;
    println!("macOS-specific code");
}

#[cfg(target_arch = "x86_64")]
fn x86_64_specific() {
    println!("x86_64 specific code");
}

#[cfg(target_arch = "aarch64")]
fn aarch64_specific() {
    println!("ARM64 specific code");
}

#[cfg(feature = "unstable")]
fn unstable_feature() {
    println!("Unstable feature code");
}

// Cross-platform function
fn cross_platform() {
    #[cfg(windows)]
    let separator = "\\";
    
    #[cfg(not(windows))]
    let separator = "/";
    
    println!("Path separator: {}", separator);
}
`

	astInfo, err := parser.ParseRustFile([]byte(platformCode), "platform_specific.rs")
	if err != nil {
		t.Fatalf("Failed to parse platform-specific code: %v", err)
	}

	// Verify that all functions are parsed (regardless of current platform)
	expectedFunctions := []string{
		"windows_specific",
		"linux_specific", 
		"macos_specific",
		"x86_64_specific",
		"aarch64_specific",
		"unstable_feature",
		"cross_platform",
	}

	for _, expectedFunc := range expectedFunctions {
		found := findFunctionByName(astInfo.Functions, expectedFunc)
		if found == nil {
			t.Errorf("Expected function %s not found in AST", expectedFunc)
		}
	}
}

// TestRustSynParser_PathNormalization tests path normalization across platforms
func TestRustSynParser_PathNormalization(t *testing.T) {
	parser, err := NewRustSynParser(false)
	if err != nil {
		t.Skipf("Skipping test: Rust syn parser not available: %v", err)
	}
	defer parser.Cleanup()

	rustCode := `fn test() {}`

	testCases := []struct {
		input    string
		expected string
	}{
		{"./test.rs", "./test.rs"},
		{"../test.rs", "../test.rs"},
		{"src/../test.rs", "src/../test.rs"},
		{"./src/./test.rs", "./src/./test.rs"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Normalize_%s", strings.ReplaceAll(tc.input, "/", "_")), func(t *testing.T) {
			astInfo, err := parser.ParseRustFile([]byte(rustCode), tc.input)
			if err != nil {
				t.Errorf("Failed to parse with path %s: %v", tc.input, err)
				return
			}

			// The parser should preserve the original path as given
			if astInfo.FilePath != tc.input {
				t.Errorf("Path normalization unexpected: input %s, expected %s, got %s", 
					tc.input, tc.expected, astInfo.FilePath)
			}
		})
	}
}

// TestRustSynParser_LargePathLengths tests handling of very long file paths
func TestRustSynParser_LargePathLengths(t *testing.T) {
	parser, err := NewRustSynParser(false)
	if err != nil {
		t.Skipf("Skipping test: Rust syn parser not available: %v", err)
	}
	defer parser.Cleanup()

	rustCode := `fn test() { println!("Testing long paths"); }`

	// Generate progressively longer paths
	lengths := []int{100, 500, 1000, 2000}
	
	for _, length := range lengths {
		t.Run(fmt.Sprintf("PathLength_%d", length), func(t *testing.T) {
			// Create a long path by repeating directory names
			var pathBuilder strings.Builder
			remaining := length
			
			for remaining > 0 {
				// Guard against underflow - need at least 3 chars for ".rs"
				if remaining <= 3 {
					break
				}
				
				segment := "very_long_directory_name_segment"
				if remaining < len(segment) {
					segment = segment[:remaining-3] // Leave room for ".rs"
				}
				
				pathBuilder.WriteString(segment)
				remaining -= len(segment)
				
				if remaining > 3 { // Room for "/" + ".rs"
					pathBuilder.WriteString("/")
					remaining--
				}
			}
			pathBuilder.WriteString(".rs")
			
			longPath := pathBuilder.String()
			
			// Test parsing with long path
			astInfo, err := parser.ParseRustFile([]byte(rustCode), longPath)
			if err != nil {
				// Very long paths might fail on some systems, which is acceptable
				t.Logf("Long path parsing failed (acceptable): %v", err)
				return
			}

			if astInfo.FilePath != longPath {
				t.Errorf("Long path mismatch: expected length %d, got length %d", 
					len(longPath), len(astInfo.FilePath))
			}
		})
	}
}

// TestRustSynParser_SpecialCharacterPaths tests paths with special characters
func TestRustSynParser_SpecialCharacterPaths(t *testing.T) {
	parser, err := NewRustSynParser(false)
	if err != nil {
		t.Skipf("Skipping test: Rust syn parser not available: %v", err)
	}
	defer parser.Cleanup()

	rustCode := `fn special_chars_test() {}`

	// Characters that might cause issues in paths
	specialPaths := []string{
		"spaces in path/file.rs",
		"path-with-hyphens/file-name.rs",
		"path_with_underscores/file_name.rs",
		"path.with.dots/file.name.rs",
		"path+with+plus/file+name.rs",
		"path@with@at/file@name.rs",
		"path#with#hash/file#name.rs",
		"path%with%percent/file%name.rs",
		"path&with&ampersand/file&name.rs",
	}

	// Skip problematic characters on Windows
	if runtime.GOOS != "windows" {
		specialPaths = append(specialPaths,
			"path:with:colon/file:name.rs",
			"path*with*star/file*name.rs",
			"path?with?question/file?name.rs",
			"path<with<bracket/file<name.rs",
			"path>with>bracket/file>name.rs",
			"path|with|pipe/file|name.rs",
		)
	}

	for _, path := range specialPaths {
		t.Run(fmt.Sprintf("Special_%s", strings.ReplaceAll(path, "/", "_")), func(t *testing.T) {
			astInfo, err := parser.ParseRustFile([]byte(rustCode), path)
			if err != nil {
				t.Errorf("Failed to parse with special character path %s: %v", path, err)
				return
			}

			if astInfo.FilePath != path {
				t.Errorf("Special character path mismatch: expected %s, got %s", path, astInfo.FilePath)
			}
		})
	}
}

// TestRustSynParser_ConcurrentCrossPlatform tests concurrent operations across platforms
func TestRustSynParser_ConcurrentCrossPlatform(t *testing.T) {
	parser, err := NewRustSynParser(false)
	if err != nil {
		t.Skipf("Skipping test: Rust syn parser not available: %v", err)
	}
	defer parser.Cleanup()

	rustCode := `
fn concurrent_cross_platform_test(id: usize) {
    println!("Concurrent test {}", id);
}
`

	// Test concurrent parsing with platform-appropriate paths
	var testPaths []string
	basePaths := []string{
		"src/lib.rs",
		"tests/integration.rs", 
		"examples/basic.rs",
		"benches/performance.rs",
	}

	// Add platform-specific prefixes
	switch runtime.GOOS {
	case "windows":
		for i, path := range basePaths {
			testPaths = append(testPaths, fmt.Sprintf("C:\\project%d\\%s", i, filepath.FromSlash(path)))
		}
	default:
		for i, path := range basePaths {
			testPaths = append(testPaths, fmt.Sprintf("/home/user/project%d/%s", i, path))
		}
	}

	const numGoroutines = 10
	errors := make(chan error, len(testPaths)*numGoroutines)
	done := make(chan bool)

	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			for j, path := range testPaths {
				filePath := fmt.Sprintf("%s_%d_%d", path, goroutineID, j)
				_, err := parser.ParseRustFile([]byte(rustCode), filePath)
				if err != nil {
					errors <- fmt.Errorf("goroutine %d, path %s: %w", goroutineID, filePath, err)
				}
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent cross-platform parsing error: %v", err)
	}

	// Verify parser state
	if err := parser.ValidateMemoryState(); err != nil {
		t.Errorf("Memory state validation failed: %v", err)
	}
}

// TestRustSynParser_FileSystemLimits tests behavior near filesystem limits
func TestRustSynParser_FileSystemLimits(t *testing.T) {
	parser, err := NewRustSynParser(false)
	if err != nil {
		t.Skipf("Skipping test: Rust syn parser not available: %v", err)
	}
	defer parser.Cleanup()

	rustCode := `fn filesystem_limits_test() {}`

	// Test very long file names (filesystem dependent)
	longFileName := strings.Repeat("very_long_file_name", 20) + ".rs"
	
	_, err = parser.ParseRustFile([]byte(rustCode), longFileName)
	if err != nil {
		// This might fail on some filesystems, which is acceptable
		t.Logf("Long filename test failed (acceptable): %v", err)
	}

	// Test deeply nested paths
	deepPath := strings.Repeat("deep/", 50) + "file.rs"
	
	_, err = parser.ParseRustFile([]byte(rustCode), deepPath)
	if err != nil {
		// This might fail due to path length limits, which is acceptable
		t.Logf("Deep path test failed (acceptable): %v", err)
	}
}

// BenchmarkRustSynParser_CrossPlatformPerformance benchmarks performance across platforms
func BenchmarkRustSynParser_CrossPlatformPerformance(b *testing.B) {
	parser, err := NewRustSynParser(false)
	if err != nil {
		b.Skipf("Skipping benchmark: Rust syn parser not available: %v", err)
	}
	defer parser.Cleanup()

	rustCode := []byte(`
fn cross_platform_benchmark() {
    println!("Cross-platform performance test");
}

struct CrossPlatformStruct {
    field1: i32,
    field2: String,
}

impl CrossPlatformStruct {
    fn new() -> Self {
        Self {
            field1: 42,
            field2: "test".to_string(),
        }
    }
}
`)

	// Platform-specific path format
	var testPath string
	switch runtime.GOOS {
	case "windows":
		testPath = "C:\\project\\src\\benchmark.rs"
	default:
		testPath = "/home/user/project/src/benchmark.rs"
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := parser.ParseRustFile(rustCode, testPath)
		if err != nil {
			b.Fatalf("Cross-platform benchmark failed: %v", err)
		}
	}

	b.StopTimer()
	
	// Report platform-specific information
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	b.ReportMetric(float64(m.Alloc), "bytes_allocated")
	b.ReportMetric(float64(runtime.NumGoroutine()), "goroutines")
	b.ReportMetric(float64(runtime.NumCPU()), "cpus")
	
	metrics := parser.GetMemorySafetyMetrics()
	b.Logf("Platform: %s, Arch: %s", runtime.GOOS, runtime.GOARCH)
	b.Logf("Memory metrics: %+v", metrics)
}

// TestRustSynParser_PlatformSpecificFeatures tests platform-specific parser features
func TestRustSynParser_PlatformSpecificFeatures(t *testing.T) {
	parser, err := NewRustSynParser(true)
	if err != nil {
		t.Skipf("Skipping test: Rust syn parser not available: %v", err)
	}
	defer parser.Cleanup()

	// Test getting capabilities and check for platform-specific features
	capabilities, err := parser.GetCapabilities()
	if err != nil {
		t.Fatalf("Failed to get capabilities: %v", err)
	}

	t.Logf("Platform: %s/%s", runtime.GOOS, runtime.GOARCH)
	t.Logf("Parser capabilities: %+v", capabilities)

	// Test version information
	version, err := parser.GetVersion()
	if err != nil {
		t.Fatalf("Failed to get version: %v", err)
	}

	t.Logf("Parser version info: %+v", version)

	// Platform-specific performance testing
	startTime := time.Now()
	rustCode := `fn performance_test() { println!("Testing"); }`
	
	for i := 0; i < 100; i++ {
		_, err := parser.ParseRustFile([]byte(rustCode), fmt.Sprintf("perf_%d.rs", i))
		if err != nil {
			t.Fatalf("Performance test failed: %v", err)
		}
	}
	
	duration := time.Since(startTime)
	t.Logf("Platform %s/%s: parsed 100 files in %v (%.2f files/sec)", 
		runtime.GOOS, runtime.GOARCH, duration, 100.0/duration.Seconds())
}

// TestRustSynParser_MemoryBehaviorAcrossPlatforms tests memory behavior differences
func TestRustSynParser_MemoryBehaviorAcrossPlatforms(t *testing.T) {
	parser, err := NewRustSynParser(true)
	if err != nil {
		t.Skipf("Skipping test: Rust syn parser not available: %v", err)
	}
	defer parser.Cleanup()

	var initialMem runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&initialMem)

	// Perform parsing operations and monitor memory
	rustCode := `
fn memory_test_function(param: i32) -> String {
    format!("Memory test: {}", param)
}

struct MemoryTestStruct {
    data: Vec<String>,
}
`

	const iterations = 200
	memSnapshots := make([]uint64, iterations)

	for i := 0; i < iterations; i++ {
		_, err := parser.ParseRustFile([]byte(rustCode), fmt.Sprintf("memory_test_%d.rs", i))
		if err != nil {
			t.Fatalf("Memory test parsing failed at iteration %d: %v", i, err)
		}

		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		memSnapshots[i] = m.Alloc

		// Force cleanup periodically
		if i%50 == 49 {
			parser.ForceCleanup()
			runtime.GC()
		}
	}

	var finalMem runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&finalMem)

	// Analyze memory pattern
	maxMem := memSnapshots[0]
	minMem := memSnapshots[0]
	for _, mem := range memSnapshots {
		if mem > maxMem {
			maxMem = mem
		}
		if mem < minMem {
			minMem = mem
		}
	}

	memIncrease := finalMem.Alloc - initialMem.Alloc
	memRange := maxMem - minMem

	t.Logf("Platform %s/%s memory behavior:", runtime.GOOS, runtime.GOARCH)
	t.Logf("  Initial memory: %d bytes", initialMem.Alloc)
	t.Logf("  Final memory: %d bytes", finalMem.Alloc)
	t.Logf("  Net increase: %d bytes", memIncrease)
	t.Logf("  Memory range: %d bytes", memRange)
	t.Logf("  Max memory: %d bytes", maxMem)

	// Verify memory state
	if err := parser.ValidateMemoryState(); err != nil {
		t.Errorf("Final memory state validation failed: %v", err)
	}

	// Platform-specific memory thresholds (adjust as needed)
	maxAcceptableIncrease := uint64(5 * 1024 * 1024) // 5MB
	if memIncrease > maxAcceptableIncrease {
		t.Errorf("Memory increase %d bytes exceeds threshold %d bytes on platform %s/%s",
			memIncrease, maxAcceptableIncrease, runtime.GOOS, runtime.GOARCH)
	}
}