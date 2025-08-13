package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetDefaultConfig(t *testing.T) {
	cfg := GetDefaultConfig()

	// Test scan configuration defaults
	if len(cfg.Scan.Paths) == 0 {
		t.Error("Expected default scan paths, got none")
	}
	expectedPaths := []string{"./src", "./internal", "./cmd"}
	for i, path := range expectedPaths {
		if i >= len(cfg.Scan.Paths) || cfg.Scan.Paths[i] != path {
			t.Errorf("Expected scan path %s at index %d, got %v", path, i, cfg.Scan.Paths)
		}
	}

	// Test thresholds defaults
	if cfg.Thresholds.FunctionLines != 25 {
		t.Errorf("Expected function lines threshold 25, got %d", cfg.Thresholds.FunctionLines)
	}
	if cfg.Thresholds.CyclomaticComplexity != 8 {
		t.Errorf("Expected cyclomatic complexity threshold 8, got %d", cfg.Thresholds.CyclomaticComplexity)
	}
	if cfg.Thresholds.Parameters != 4 {
		t.Errorf("Expected parameters threshold 4, got %d", cfg.Thresholds.Parameters)
	}

	// Test output configuration defaults
	if cfg.Output.HTML.Path != "./reports/clean-code-report.html" {
		t.Errorf("Expected HTML path './reports/clean-code-report.html', got %s", cfg.Output.HTML.Path)
	}
	if cfg.Output.HTML.AutoRefresh {
		t.Error("Expected HTML auto refresh to be disabled by default")
	}
	if cfg.Output.HTML.RefreshInterval != 10 {
		t.Errorf("Expected HTML refresh interval 10, got %d", cfg.Output.HTML.RefreshInterval)
	}

	// Test logging configuration defaults
	if cfg.Logging.Level != "info" {
		t.Errorf("Expected logging level 'info', got %s", cfg.Logging.Level)
	}
	if cfg.Logging.Format != "structured" {
		t.Errorf("Expected logging format 'structured', got %s", cfg.Logging.Format)
	}
}

func TestLoadConfigFromFile(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.yaml")

	configContent := `scan:
  paths:
    - "./test"
  exclude:
    - "*.tmp"
  file_types:
    - ".go"

thresholds:
  function_lines: 30
  cyclomatic_complexity: 10
  parameters: 6
  nesting_depth: 4
  class_lines: 200

output:
  html:
    path: "./test-reports/report.html"
    auto_refresh: false
    refresh_interval: 5
    theme: "dark"

logging:
  level: "debug"
  format: "plain"
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Load configuration
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify loaded values
	if len(cfg.Scan.Paths) != 1 || cfg.Scan.Paths[0] != "./test" {
		t.Errorf("Expected scan paths [./test], got %v", cfg.Scan.Paths)
	}
	if cfg.Thresholds.FunctionLines != 30 {
		t.Errorf("Expected function lines 30, got %d", cfg.Thresholds.FunctionLines)
	}
	if cfg.Output.HTML.Theme != "dark" {
		t.Errorf("Expected theme 'dark', got %s", cfg.Output.HTML.Theme)
	}
	if cfg.Logging.Level != "debug" {
		t.Errorf("Expected logging level 'debug', got %s", cfg.Logging.Level)
	}
}

func TestLoadConfigNonExistent(t *testing.T) {
	// Try to load non-existent config file
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Expected no error for non-existent config, got: %v", err)
	}

	// Should return default config
	defaultCfg := GetDefaultConfig()
	if cfg.Thresholds.FunctionLines != defaultCfg.Thresholds.FunctionLines {
		t.Error("Expected default configuration when no config file found")
	}
}

func TestSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "save-test.yaml")

	cfg := GetDefaultConfig()
	cfg.Thresholds.FunctionLines = 20
	cfg.Logging.Level = "warn"

	// Save configuration
	err := Save(cfg, configPath)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Load and verify
	loadedCfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if loadedCfg.Thresholds.FunctionLines != 20 {
		t.Errorf("Expected function lines 20, got %d", loadedCfg.Thresholds.FunctionLines)
	}
	if loadedCfg.Logging.Level != "warn" {
		t.Errorf("Expected logging level 'warn', got %s", loadedCfg.Logging.Level)
	}
}

func TestConfigValidation(t *testing.T) {
	testCases := []struct {
		name        string
		modifyFunc  func(*Config)
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			modifyFunc: func(cfg *Config) {
				// Default config should be valid
			},
			expectError: false,
		},
		{
			name: "empty paths",
			modifyFunc: func(cfg *Config) {
				cfg.Scan.Paths = []string{}
			},
			expectError: true,
			errorMsg:    "at least one scan path must be specified",
		},
		{
			name: "invalid function lines threshold",
			modifyFunc: func(cfg *Config) {
				cfg.Thresholds.FunctionLines = 0
			},
			expectError: true,
			errorMsg:    "function_lines threshold must be positive",
		},
		{
			name: "invalid complexity threshold",
			modifyFunc: func(cfg *Config) {
				cfg.Thresholds.CyclomaticComplexity = -1
			},
			expectError: true,
			errorMsg:    "cyclomatic_complexity threshold must be positive",
		},
		{
			name: "invalid logging level",
			modifyFunc: func(cfg *Config) {
				cfg.Logging.Level = "invalid"
			},
			expectError: true,
			errorMsg:    "invalid logging level: invalid (must be one of: debug, info, warn, error)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := GetDefaultConfig()
			// Set paths to current directory for validation to pass
			cfg.Scan.Paths = []string{"."}
			
			tc.modifyFunc(cfg)

			err := cfg.Validate()
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error for %s, got none", tc.name)
				} else if tc.errorMsg != "" && err.Error() != tc.errorMsg {
					t.Errorf("Expected error message '%s', got '%s'", tc.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for %s, got: %v", tc.name, err)
				}
			}
		})
	}
}

func TestMergeWithDefaults(t *testing.T) {
	cfg := &Config{}
	mergeWithDefaults(cfg)

	defaults := GetDefaultConfig()

	// Test that empty config gets filled with defaults
	if len(cfg.Scan.Paths) != len(defaults.Scan.Paths) {
		t.Error("Expected scan paths to be filled with defaults")
	}
	if cfg.Thresholds.FunctionLines != defaults.Thresholds.FunctionLines {
		t.Error("Expected function lines threshold to be filled with default")
	}
	if cfg.Logging.Level != defaults.Logging.Level {
		t.Error("Expected logging level to be filled with default")
	}
}

func TestGetConfigPaths(t *testing.T) {
	paths := GetConfigPaths()

	expectedPaths := []string{
		"goclean.yaml",
		"goclean.yml",
		".goclean.yaml",
		".goclean.yml",
		"configs/goclean.yaml",
		"configs/goclean.yml",
	}

	if len(paths) != len(expectedPaths) {
		t.Errorf("Expected %d config paths, got %d", len(expectedPaths), len(paths))
	}

	for i, expected := range expectedPaths {
		if i >= len(paths) || paths[i] != expected {
			t.Errorf("Expected config path %s at index %d, got %v", expected, i, paths)
		}
	}
}

func TestPartialConfigFile(t *testing.T) {
	// Create temporary config file with only some settings
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "partial-config.yaml")

	configContent := `thresholds:
  function_lines: 15

logging:
  level: "error"
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create partial config file: %v", err)
	}

	// Load configuration
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load partial config: %v", err)
	}

	// Verify partial values are loaded
	if cfg.Thresholds.FunctionLines != 15 {
		t.Errorf("Expected function lines 15, got %d", cfg.Thresholds.FunctionLines)
	}
	if cfg.Logging.Level != "error" {
		t.Errorf("Expected logging level 'error', got %s", cfg.Logging.Level)
	}

	// Verify defaults are used for missing values
	defaults := GetDefaultConfig()
	if cfg.Thresholds.CyclomaticComplexity != defaults.Thresholds.CyclomaticComplexity {
		t.Error("Expected default cyclomatic complexity when not specified in partial config")
	}
	if len(cfg.Scan.Paths) != len(defaults.Scan.Paths) {
		t.Error("Expected default scan paths when not specified in partial config")
	}
}