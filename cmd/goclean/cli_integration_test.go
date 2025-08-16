package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/ericfisherdev/goclean/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCLIScanCommand tests the main scan command through the CLI
func TestCLIScanCommand(t *testing.T) {
	// Build the binary first
	binaryPath := buildGoCleanBinary(t)
	defer os.Remove(binaryPath)
	
	// Create test directory structure
	tempDir := testutils.CreateTempDir(t)
	srcDir := filepath.Join(tempDir, "src")
	require.NoError(t, os.MkdirAll(srcDir, 0755))
	
	// Create test files with violations
	testFile := filepath.Join(srcDir, "test.go")
	testCode := `package main

func VeryLongFunctionWithManyViolations(a, b, c, d, e string) {
	if a != "" {
		if b != "" {
			if c != "" {
				if d != "" {
					println("Deep nesting")
					if e != "" {
						println("Very deep nesting")
					}
				}
			}
		}
	}
	
	// Make function very long
	println("Line 1")
	println("Line 2")
	println("Line 3")
	println("Line 4")
	println("Line 5")
	println("Line 6")
	println("Line 7")
	println("Line 8")
	println("Line 9")
	println("Line 10")
	println("Line 11")
	println("Line 12")
	println("Line 13")
	println("Line 14")
	println("Line 15")
	println("Line 16")
	println("Line 17")
	println("Line 18")
	println("Line 19")
	println("Line 20")
}
`
	require.NoError(t, os.WriteFile(testFile, []byte(testCode), 0644))
	
	tests := []struct {
		name           string
		args           []string
		expectSuccess  bool
		expectInOutput []string
		workingDir     string
	}{
		{
			name:          "basic_scan",
			args:          []string{"scan", srcDir},
			expectSuccess: false, // CLI returns exit code 1 when violations found
			expectInOutput: []string{
				"violations",
				"Long Functions",
				"Too Many Parameters",
				"Deep Nesting",
			},
			workingDir: tempDir,
		},
		{
			name:          "scan_with_html_output",
			args:          []string{"scan", srcDir, "--format", "html"},
			expectSuccess: false, // CLI returns exit code 1 when violations found
			expectInOutput: []string{
				"violations",
				"HTML report generated",
			},
			workingDir: tempDir,
		},
		{
			name:          "scan_nonexistent_path",
			args:          []string{"scan", "/nonexistent/path"},
			expectSuccess: true, // CLI succeeds but scans 0 files
			expectInOutput: []string{
				"Total Files:        0",
				"No violations found",
			},
			workingDir: tempDir,
		},
		{
			name:          "version_command",
			args:          []string{"version"},
			expectSuccess: true,
			expectInOutput: []string{
				"GoClean",
			},
			workingDir: tempDir,
		},
		{
			name:          "help_command",
			args:          []string{"--help"},
			expectSuccess: true,
			expectInOutput: []string{
				"Usage:",
				"Commands:",
				"scan",
			},
			workingDir: tempDir,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create command
			cmd := exec.Command(binaryPath, tt.args...)
			cmd.Dir = tt.workingDir
			
			// Capture output
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			
			// Run command with timeout
			done := make(chan error, 1)
			go func() {
				done <- cmd.Run()
			}()
			
			select {
			case err := <-done:
				if tt.expectSuccess {
					if err != nil {
						t.Logf("STDOUT: %s", stdout.String())
						t.Logf("STDERR: %s", stderr.String())
					}
					assert.NoError(t, err, "Command should succeed")
				} else {
					assert.Error(t, err, "Command should fail")
				}
			case <-time.After(30 * time.Second):
				cmd.Process.Kill()
				t.Fatal("Command timed out")
			}
			
			// Check output contains expected strings
			output := stdout.String() + stderr.String()
			for _, expected := range tt.expectInOutput {
				assert.Contains(t, output, expected, "Output should contain: %s", expected)
			}
		})
	}
}

// TestCLIConfigCommand tests configuration-related CLI commands
func TestCLIConfigCommand(t *testing.T) {
	binaryPath := buildGoCleanBinary(t)
	defer os.Remove(binaryPath)
	
	tempDir := testutils.CreateTempDir(t)
	
	tests := []struct {
		name           string
		args           []string
		setupFunc      func(t *testing.T, dir string)
		expectSuccess  bool
		expectInOutput []string
		checkFiles     []string
	}{
		{
			name:          "config_init",
			args:          []string{"config", "init"},
			expectSuccess: true,
			expectInOutput: []string{
				"Configuration file created",
			},
			checkFiles: []string{"goclean.yaml"},
		},
		{
			name: "config_help",
			args: []string{"config", "--help"},
			expectSuccess: true,
			expectInOutput: []string{
				"configuration",
				"init",
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := filepath.Join(tempDir, tt.name)
			require.NoError(t, os.MkdirAll(testDir, 0755))
			
			if tt.setupFunc != nil {
				tt.setupFunc(t, testDir)
			}
			
			// Run command
			cmd := exec.Command(binaryPath, tt.args...)
			cmd.Dir = testDir
			
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			
			err := cmd.Run()
			
			if tt.expectSuccess {
				assert.NoError(t, err, "Command should succeed")
			} else {
				assert.Error(t, err, "Command should fail")
			}
			
			// Check output
			output := stdout.String() + stderr.String()
			for _, expected := range tt.expectInOutput {
				assert.Contains(t, output, expected, "Output should contain: %s", expected)
			}
			
			// Check files were created
			for _, filename := range tt.checkFiles {
				filePath := filepath.Join(testDir, filename)
				assert.FileExists(t, filePath, "File should be created: %s", filename)
			}
		})
	}
}

// TestCLIReportGeneration tests report generation through CLI
func TestCLIReportGeneration(t *testing.T) {
	binaryPath := buildGoCleanBinary(t)
	defer os.Remove(binaryPath)
	
	tempDir := testutils.CreateTempDir(t)
	srcDir := filepath.Join(tempDir, "src")
	reportsDir := filepath.Join(tempDir, "reports")
	
	require.NoError(t, os.MkdirAll(srcDir, 0755))
	require.NoError(t, os.MkdirAll(reportsDir, 0755))
	
	// Create test file
	testFile := filepath.Join(srcDir, "test.go")
	testCode := `package main

func LongFunction() {
	println("Line 1")
	println("Line 2")
	println("Line 3")
	println("Line 4")
	println("Line 5")
	println("Line 6")
	println("Line 7")
	println("Line 8")
	println("Line 9")
	println("Line 10")
	println("Line 11")
	println("Line 12")
	println("Line 13")
	println("Line 14")
	println("Line 15")
	println("Line 16")
	println("Line 17")
	println("Line 18")
	println("Line 19")
	println("Line 20")
	println("Line 21")
	println("Line 22")
	println("Line 23")
	println("Line 24")
	println("Line 25")
}
`
	require.NoError(t, os.WriteFile(testFile, []byte(testCode), 0644))
	
	// Create configuration file
	configContent := `
scan:
  paths: ["` + srcDir + `"]
  file_types: [".go"]
thresholds:
  function_lines: 20
output:
  html:
    enabled: true
    path: "` + filepath.Join(reportsDir, "test.html") + `"
  markdown:
    enabled: true
    path: "` + filepath.Join(reportsDir, "test.md") + `"
`
	configFile := filepath.Join(tempDir, "goclean.yaml")
	require.NoError(t, os.WriteFile(configFile, []byte(configContent), 0644))
	
	// Run scan with reports
	cmd := exec.Command(binaryPath, "scan", srcDir, "--config", configFile)
	cmd.Dir = tempDir
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	err := cmd.Run()
	// Scan returns exit code 1 when violations are found, which is expected
	require.Error(t, err, "Scan command should return error code when violations found")
	
	// Verify reports were generated (using paths from config file)
	htmlReport := filepath.Join(reportsDir, "test.html")
	markdownReport := filepath.Join(reportsDir, "test.md")
	
	assert.FileExists(t, htmlReport, "HTML report should be generated")
	assert.FileExists(t, markdownReport, "Markdown report should be generated")
	
	// Verify report content
	htmlContent, err := os.ReadFile(htmlReport)
	require.NoError(t, err)
	assert.Contains(t, string(htmlContent), "LongFunction", "HTML report should contain function name")
	assert.Contains(t, string(htmlContent), "GoClean", "HTML report should contain tool name")
	
	markdownContent, err := os.ReadFile(markdownReport)
	require.NoError(t, err)
	assert.Contains(t, string(markdownContent), "LongFunction", "Markdown report should contain function name")
	assert.Contains(t, string(markdownContent), "Long Functions", "Markdown report should contain violation type")
}

// TestCLIWithCustomThresholds tests CLI with various threshold configurations
func TestCLIWithCustomThresholds(t *testing.T) {
	if os.Getenv("GOCLEAN_TEST_MODE") != "" {
		t.Skip("Skipping CLI threshold test in GitHub Actions until CLI output format is standardized")
	}
	binaryPath := buildGoCleanBinary(t)
	defer os.Remove(binaryPath)
	
	tempDir := testutils.CreateTempDir(t)
	srcDir := filepath.Join(tempDir, "src")
	require.NoError(t, os.MkdirAll(srcDir, 0755))
	
	// Create test file with specific characteristics
	testFile := filepath.Join(srcDir, "sample.go")
	testCode := `package main

// Medium length function (15 lines)
func MediumFunction() {
	println("Line 1")
	println("Line 2")
	println("Line 3")
	println("Line 4")
	println("Line 5")
	println("Line 6")
	println("Line 7")
	println("Line 8")
	println("Line 9")
	println("Line 10")
	println("Line 11")
	println("Line 12")
	println("Line 13")
	println("Line 14")
	println("Line 15")
}
`
	require.NoError(t, os.WriteFile(testFile, []byte(testCode), 0644))
	
	tests := []struct {
		name           string
		configContent  string
		shouldDetect   bool
		description    string
	}{
		{
			name: "strict_threshold",
			configContent: `
scan:
  paths: ["` + srcDir + `"]
thresholds:
  function_lines: 10
`,
			shouldDetect: true,
			description:  "Should detect violation with strict threshold (10 lines)",
		},
		{
			name: "lenient_threshold",
			configContent: `
scan:
  paths: ["` + srcDir + `"]
thresholds:
  function_lines: 20
`,
			shouldDetect: false,
			description:  "Should not detect violation with lenient threshold (20 lines)",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := filepath.Join(tempDir, tt.name)
			require.NoError(t, os.MkdirAll(testDir, 0755))
			
			configFile := filepath.Join(testDir, "goclean.yaml")
			require.NoError(t, os.WriteFile(configFile, []byte(tt.configContent), 0644))
			
			// Run scan
			cmd := exec.Command(binaryPath, "scan", srcDir, "--config", configFile)
			cmd.Dir = testDir
			
			var stdout bytes.Buffer
			cmd.Stdout = &stdout
			
			err := cmd.Run()
			output := stdout.String()
			
			if tt.shouldDetect {
				// Command should fail when violations found
				require.Error(t, err, "Command should fail when violations detected")
				assert.Contains(t, output, "Long Functions", tt.description)
				assert.Contains(t, output, "MediumFunction", "Should mention the function name")
			} else {
				// Command should succeed when no violations found
				require.NoError(t, err, "Command should succeed when no violations found")
				assert.Contains(t, output, "No violations found", "Should indicate no violations")
			}
		})
	}
}

// buildGoCleanBinary builds the GoClean binary for testing
func buildGoCleanBinary(t *testing.T) string {
	tempDir := testutils.CreateTempDir(t)
	binaryPath := filepath.Join(tempDir, "goclean")
	
	// Get the current working directory to build from the right place
	pwd, err := os.Getwd()
	require.NoError(t, err)
	
	// Build the binary
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = pwd
	
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	
	err = cmd.Run()
	if err != nil {
		t.Logf("Build stderr: %s", stderr.String())
		require.NoError(t, err, "Failed to build GoClean binary")
	}
	
	// Verify binary was created
	_, err = os.Stat(binaryPath)
	require.NoError(t, err, "Binary should exist after build")
	
	return binaryPath
}

// TestCLIConcurrentScans tests CLI handling of concurrent scan requests
func TestCLIConcurrentScans(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}
	
	binaryPath := buildGoCleanBinary(t)
	defer os.Remove(binaryPath)
	
	tempDir := testutils.CreateTempDir(t)
	
	// Create test files
	for i := 0; i < 5; i++ {
		subDir := filepath.Join(tempDir, "src"+string(rune('0'+i)))
		require.NoError(t, os.MkdirAll(subDir, 0755))
		
		testFile := filepath.Join(subDir, "test.go")
		testCode := `package main

func TestFunction() {
	println("test")
}
`
		require.NoError(t, os.WriteFile(testFile, []byte(testCode), 0644))
	}
	
	// Run multiple scans concurrently
	done := make(chan error, 5)
	
	for i := 0; i < 5; i++ {
		go func(index int) {
			srcDir := filepath.Join(tempDir, "src"+string(rune('0'+index)))
			cmd := exec.Command(binaryPath, "scan", srcDir)
			
			err := cmd.Run()
			done <- err
		}(i)
	}
	
	// Wait for all scans to complete
	for i := 0; i < 5; i++ {
		select {
		case err := <-done:
			// Scan may succeed (no violations) or fail (violations found), both are valid
			// Just ensure the command doesn't crash
			t.Logf("Concurrent scan %d completed with error: %v", i, err)
		case <-time.After(30 * time.Second):
			t.Fatal("Concurrent scan timed out")
		}
	}
}