// Package violations provides violation detection capabilities for GoClean.
// This file implements a comprehensive magic number whitelist system to reduce false positives.
package violations

import (
	"fmt"
	"strings"
)

// MagicNumberWhitelist manages acceptable magic numbers based on context
type MagicNumberWhitelist struct {
	// System constants
	SystemConstants map[string]string
	
	// File permissions
	FilePermissions map[string]string
	
	// Network/HTTP constants
	NetworkConstants map[string]string
	
	// Mathematical constants
	MathConstants map[string]string
	
	// Common array/slice indices
	CommonIndices map[string]string
	
	// Framework-specific constants
	FrameworkConstants map[string]map[string]string
}

// Whitelist management constants
const (
	MaxWhitelistEntries = 1000 // Maximum number of whitelist entries to prevent memory issues
)

// DefaultMagicNumberWhitelist returns a comprehensive whitelist for common magic numbers
func DefaultMagicNumberWhitelist() *MagicNumberWhitelist {
	return &MagicNumberWhitelist{
		SystemConstants: map[string]string{
			"0":    "Zero value/false",
			"1":    "One/true/first item",
			"-1":   "Not found/error indicator",
			"2":    "Common small number",
			"3":    "Common small number", 
			"4":    "Common small number",
			"5":    "Common small number",
			"8":    "Standard Go byte size",
			"10":   "Decimal base",
			"16":   "Standard Go int16 size",
			"32":   "Standard Go float32 bit size",
			"64":   "Standard Go float64 bit size",
			"100":  "Percentage base",
			"1000": "Metric base",
			"1024": "Standard memory unit (1KB)",
			"2048": "Standard memory unit (2KB)",
			"4096": "Standard memory unit (4KB)",
			"8192": "Standard memory unit (8KB)",
		},
		FilePermissions: map[string]string{
			"0644": "Standard file permissions (rw-r--r--)",
			"0755": "Standard directory permissions (rwxr-xr-x)",
			"0600": "Secure file permissions (rw-------)",
			"0700": "Secure directory permissions (rwx------)",
			"0666": "World writable file permissions",
			"0777": "World accessible permissions",
		},
		NetworkConstants: map[string]string{
			"80":   "HTTP port",
			"443":  "HTTPS port",
			"22":   "SSH port",
			"21":   "FTP port",
			"25":   "SMTP port",
			"53":   "DNS port",
			"8080": "Alternative HTTP port",
			"3000": "Development server port",
			"5000": "Development server port",
			"8000": "Development server port",
		},
		MathConstants: map[string]string{
			"0.0":   "Zero float",
			"1.0":   "One float",
			"0.5":   "Half",
			"2.0":   "Two float",
			"360":   "Degrees in circle",
			"180":   "Half circle degrees",
			"90":    "Quarter circle degrees",
			"24":    "Hours in day",
			"60":    "Minutes/seconds",
			"7":     "Days in week",
			"12":    "Months in year",
			"365":   "Days in year",
		},
		CommonIndices: map[string]string{
			"0": "First array index",
			"1": "Second array index", 
			"2": "Third array index",
		},
		FrameworkConstants: map[string]map[string]string{
			"testing": {
				"1.5": "Common test multiplier",
				"2.0": "Common test multiplier", 
				"3.0": "Common test multiplier",
				"0.5": "Common test fraction",
				"0.8": "Common test threshold",
				"1.2": "Common test multiplier",
			},
			"http": {
				"200": "HTTP OK",
				"201": "HTTP Created",
				"400": "HTTP Bad Request",
				"401": "HTTP Unauthorized", 
				"403": "HTTP Forbidden",
				"404": "HTTP Not Found",
				"500": "HTTP Internal Server Error",
			},
			"mathematical": {
				"1.5": "Mathematical factor/ratio",
				"2.0": "Mathematical multiplier",
				"0.5": "Mathematical fraction",
				"0.25": "Quarter value",
				"0.75": "Three-quarter value",
				"3.14": "Pi approximation",
				"2.71": "e approximation",
			},
			"benchmark": {
				"50": "Benchmark iteration count",
				"20": "Benchmark threshold",
				"1000": "Benchmark operation count",
				"100": "Benchmark sample size",
			},
		},
	}
}

// IsWhitelistedMagicNumber checks if a number should be ignored based on context
func (w *MagicNumberWhitelist) IsWhitelistedMagicNumber(value string, context string) (bool, string) {
	// Check system constants
	if reason, exists := w.SystemConstants[value]; exists {
		return true, reason
	}
	
	// Check file permissions
	if reason, exists := w.FilePermissions[value]; exists {
		return true, reason
	}
	
	// Check for octal file permissions
	if w.IsOctalFilePermission(value) {
		return true, "Octal file permission constant"
	}
	
	// Check network constants
	if reason, exists := w.NetworkConstants[value]; exists {
		return true, reason
	}
	
	// Check math constants
	if reason, exists := w.MathConstants[value]; exists {
		return true, reason
	}
	
	// Check common indices
	if reason, exists := w.CommonIndices[value]; exists {
		return true, reason
	}
	
	// Check framework-specific constants
	for framework, constants := range w.FrameworkConstants {
		if strings.Contains(context, framework) {
			if reason, exists := constants[value]; exists {
				return true, fmt.Sprintf("%s (%s context)", reason, framework)
			}
		}
	}
	
	return false, ""
}

// IsOctalFilePermission checks if a value represents a valid octal file permission
func (w *MagicNumberWhitelist) IsOctalFilePermission(value string) bool {
	// Common octal file permissions
	octalPerms := []string{
		"0644", "0755", "0600", "0700", "0666", "0777",
		"0640", "0750", "0660", "0770", "0440", "0550",
		"0444", "0555", "644", "755", "600", "700", "666", "777",
	}
	
	for _, perm := range octalPerms {
		if value == perm {
			return true
		}
	}
	return false
}

// AddCustomWhitelistEntry adds a custom magic number to the whitelist
func (w *MagicNumberWhitelist) AddCustomWhitelistEntry(value, reason, category string) error {
	// Validate input
	if value == "" || reason == "" {
		return fmt.Errorf("value and reason cannot be empty")
	}
	
	// Check total entries to prevent memory issues
	totalEntries := len(w.SystemConstants) + len(w.FilePermissions) + 
					len(w.NetworkConstants) + len(w.MathConstants) + len(w.CommonIndices)
	
	for _, constants := range w.FrameworkConstants {
		totalEntries += len(constants)
	}
	
	if totalEntries >= MaxWhitelistEntries {
		return fmt.Errorf("maximum whitelist entries (%d) reached", MaxWhitelistEntries)
	}
	
	// Add to appropriate category
	switch strings.ToLower(category) {
	case "system":
		w.SystemConstants[value] = reason
	case "file", "permission", "permissions":
		w.FilePermissions[value] = reason
	case "network", "port", "ports":
		w.NetworkConstants[value] = reason
	case "math", "mathematical":
		w.MathConstants[value] = reason
	case "index", "indices", "array":
		w.CommonIndices[value] = reason
	default:
		// Add to framework constants with the category as framework name
		if w.FrameworkConstants[category] == nil {
			w.FrameworkConstants[category] = make(map[string]string)
		}
		w.FrameworkConstants[category][value] = reason
	}
	
	return nil
}

// GetWhitelistStatistics returns statistics about the whitelist
func (w *MagicNumberWhitelist) GetWhitelistStatistics() map[string]int {
	stats := map[string]int{
		"system":     len(w.SystemConstants),
		"file":       len(w.FilePermissions),
		"network":    len(w.NetworkConstants),
		"math":       len(w.MathConstants),
		"indices":    len(w.CommonIndices),
		"frameworks": len(w.FrameworkConstants),
	}
	
	totalFrameworkEntries := 0
	for _, constants := range w.FrameworkConstants {
		totalFrameworkEntries += len(constants)
	}
	stats["framework_entries"] = totalFrameworkEntries
	
	stats["total"] = stats["system"] + stats["file"] + stats["network"] + 
					stats["math"] + stats["indices"] + stats["framework_entries"]
	
	return stats
}

// RemoveWhitelistEntry removes a magic number from the whitelist
func (w *MagicNumberWhitelist) RemoveWhitelistEntry(value string) bool {
	removed := false
	
	// Remove from all categories
	if _, exists := w.SystemConstants[value]; exists {
		delete(w.SystemConstants, value)
		removed = true
	}
	
	if _, exists := w.FilePermissions[value]; exists {
		delete(w.FilePermissions, value)
		removed = true
	}
	
	if _, exists := w.NetworkConstants[value]; exists {
		delete(w.NetworkConstants, value)
		removed = true
	}
	
	if _, exists := w.MathConstants[value]; exists {
		delete(w.MathConstants, value)
		removed = true
	}
	
	if _, exists := w.CommonIndices[value]; exists {
		delete(w.CommonIndices, value)
		removed = true
	}
	
	// Remove from framework constants
	for framework, constants := range w.FrameworkConstants {
		if _, exists := constants[value]; exists {
			delete(constants, value)
			removed = true
			
			// Clean up empty framework maps
			if len(constants) == 0 {
				delete(w.FrameworkConstants, framework)
			}
		}
	}
	
	return removed
}

// GetAllWhitelistedValues returns all whitelisted values for debugging
func (w *MagicNumberWhitelist) GetAllWhitelistedValues() []string {
	var values []string
	
	for value := range w.SystemConstants {
		values = append(values, value)
	}
	
	for value := range w.FilePermissions {
		values = append(values, value)
	}
	
	for value := range w.NetworkConstants {
		values = append(values, value)
	}
	
	for value := range w.MathConstants {
		values = append(values, value)
	}
	
	for value := range w.CommonIndices {
		values = append(values, value)
	}
	
	for _, constants := range w.FrameworkConstants {
		for value := range constants {
			values = append(values, value)
		}
	}
	
	return values
}