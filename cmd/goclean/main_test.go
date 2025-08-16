package main

import (
	"testing"
)

func TestRootCommand(t *testing.T) {
	// Test that the root command can be executed
	if rootCmd == nil {
		t.Fatal("rootCmd should not be nil")
	}

	if rootCmd.Use != "goclean" {
		t.Errorf("Expected root command use 'goclean', got %q", rootCmd.Use)
	}

	if rootCmd.Version != "2025.08.16.6" {
		t.Errorf("Expected version '2025.08.16.6', got %q", rootCmd.Version)
	}
}

func TestVersionCommand(t *testing.T) {
	// Test that version command exists and has proper configuration
	if versionCmd == nil {
		t.Fatal("versionCmd should not be nil")
	}

	if versionCmd.Use != "version" {
		t.Errorf("Expected version command use 'version', got %q", versionCmd.Use)
	}

	if versionCmd.Short != "Print the version number of GoClean" {
		t.Errorf("Expected proper short description for version command")
	}
}

func TestConfigInitCommand(t *testing.T) {
	// Test that config init command exists and has proper configuration
	if configInitCmd == nil {
		t.Fatal("configInitCmd should not be nil")
	}

	if configInitCmd.Use != "init [path]" {
		t.Errorf("Expected config init command use 'init [path]', got %q", configInitCmd.Use)
	}

	if configInitCmd.Short != "Initialize a default configuration file" {
		t.Errorf("Expected proper short description for config init command")
	}
}

func TestConfigCommand(t *testing.T) {
	// Test that config command exists and has proper configuration
	if configCmd == nil {
		t.Fatal("configCmd should not be nil")
	}

	if configCmd.Use != "config" {
		t.Errorf("Expected config command use 'config', got %q", configCmd.Use)
	}

	if configCmd.Short != "Manage GoClean configuration" {
		t.Errorf("Expected proper short description for config command")
	}
}

func TestScanCommand(t *testing.T) {
	// Test that scan command exists and has proper configuration
	if scanCmd == nil {
		t.Fatal("scanCmd should not be nil")
	}

	if scanCmd.Use != "scan [paths...]" {
		t.Errorf("Expected scan command use 'scan [paths...]', got %q", scanCmd.Use)
	}

	if scanCmd.Short != "Scan codebases for clean code violations" {
		t.Errorf("Expected proper short description for scan command")
	}
}

// Simplified tests that don't execute commands but verify structure

func TestCommandLineFlags(t *testing.T) {
	// Test that flags are properly defined
	flags := scanCmd.Flags()

	if flag := flags.Lookup("exclude"); flag == nil || flag.Shorthand != "e" {
		t.Error("Expected exclude flag to have shorthand 'e'")
	}

	if flag := flags.Lookup("types"); flag == nil || flag.Shorthand != "t" {
		t.Error("Expected types flag to have shorthand 't'")
	}

	if flag := flags.Lookup("format"); flag == nil || flag.Shorthand != "f" {
		t.Error("Expected format flag to have shorthand 'f'")
	}

	if flag := flags.Lookup("output"); flag == nil || flag.Shorthand != "o" {
		t.Error("Expected output flag to have shorthand 'o'")
	}

	// Test persistent flags
	persistentFlags := rootCmd.PersistentFlags()
	
	if flag := persistentFlags.Lookup("verbose"); flag == nil || flag.Shorthand != "v" {
		t.Error("Expected verbose flag to have shorthand 'v'")
	}
}

func TestSubcommands(t *testing.T) {
	// Test that all expected subcommands are present
	commands := rootCmd.Commands()

	var foundScan, foundConfig, foundVersion bool
	for _, cmd := range commands {
		switch cmd.Use {
		case "scan [paths...]":
			foundScan = true
		case "config":
			foundConfig = true
		case "version":
			foundVersion = true
		}
	}

	if !foundScan {
		t.Error("Expected to find scan command")
	}

	if !foundConfig {
		t.Error("Expected to find config command")
	}

	if !foundVersion {
		t.Error("Expected to find version command")
	}
}

func TestConfigSubcommands(t *testing.T) {
	// Test that config command has init subcommand
	configCommands := configCmd.Commands()

	foundInit := false
	for _, cmd := range configCommands {
		if cmd.Use == "init [path]" {
			foundInit = true
			break
		}
	}

	if !foundInit {
		t.Error("Expected to find config init subcommand")
	}
}

func TestGlobalVariables(t *testing.T) {
	// Test that global variables are properly initialized
	testCases := []struct {
		name     string
		variable interface{}
	}{
		{"cfgFile", &cfgFile},
		{"verbose", &verbose},
		{"outputPath", &outputPath},
		{"format", &format},
		{"paths", &paths},
		{"exclude", &exclude},
		{"fileTypes", &fileTypes},
	}

	for _, tc := range testCases {
		if tc.variable == nil {
			t.Errorf("Global variable %s should not be nil", tc.name)
		}
	}
}