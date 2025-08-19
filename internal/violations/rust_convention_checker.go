package violations

import (
	"regexp"
	"strings"
	"unicode"
)

// RustConventionChecker validates Rust naming conventions
type RustConventionChecker struct {
	// Common Rust keywords that should not be used as identifiers
	keywords map[string]bool
	
	// Common acronyms in Rust that need special handling
	commonAcronyms map[string]string
}

// NewRustConventionChecker creates a new Rust convention checker
func NewRustConventionChecker() *RustConventionChecker {
	return &RustConventionChecker{
		keywords: initRustKeywords(),
		commonAcronyms: initRustAcronyms(),
	}
}

// IsValidFunctionName checks if a function name follows Rust conventions (snake_case)
func (c *RustConventionChecker) IsValidFunctionName(name string) bool {
	// Rust functions should be snake_case
	// Special cases: new, default, from_*, into_*, as_*, to_*, is_*, has_*
	
	if c.isKeyword(name) {
		return false
	}
	
	// Check if it's snake_case
	return c.isSnakeCase(name)
}

// IsValidMethodName checks if a method name follows Rust conventions (snake_case)
func (c *RustConventionChecker) IsValidMethodName(name string) bool {
	// Methods follow the same rules as functions
	return c.IsValidFunctionName(name)
}

// IsValidVariableName checks if a variable name follows Rust conventions (snake_case)
func (c *RustConventionChecker) IsValidVariableName(name string) bool {
	// Variables should be snake_case
	// Exception: _ for unused variables
	if name == "_" {
		return true
	}
	
	if c.isKeyword(name) {
		return false
	}
	
	return c.isSnakeCase(name)
}

// IsValidTypeName checks if a type name follows Rust conventions (PascalCase)
func (c *RustConventionChecker) IsValidTypeName(name string) bool {
	// Types (structs, enums, traits) should be PascalCase
	
	if c.isKeyword(name) {
		return false
	}
	
	// Check if it's PascalCase with proper acronym handling
	return c.isPascalCase(name)
}

// IsValidConstantName checks if a constant name follows Rust conventions (SCREAMING_SNAKE_CASE)
func (c *RustConventionChecker) IsValidConstantName(name string) bool {
	// Constants should be SCREAMING_SNAKE_CASE
	
	if c.isKeyword(name) {
		return false
	}
	
	return c.isScreamingSnakeCase(name)
}

// IsValidStaticName checks if a static variable name follows Rust conventions (SCREAMING_SNAKE_CASE)
func (c *RustConventionChecker) IsValidStaticName(name string) bool {
	// Static variables follow the same rules as constants
	return c.IsValidConstantName(name)
}

// IsValidModuleName checks if a module name follows Rust conventions (snake_case)
func (c *RustConventionChecker) IsValidModuleName(name string) bool {
	// Modules should be snake_case
	
	if c.isKeyword(name) {
		return false
	}
	
	return c.isSnakeCase(name)
}

// IsValidCrateName checks if a crate name follows Rust conventions (snake_case, may contain hyphens)
func (c *RustConventionChecker) IsValidCrateName(name string) bool {
	// Crate names can be snake_case and may contain hyphens
	// They should start with a letter
	
	if len(name) == 0 {
		return false
	}
	
	// Must start with a letter
	if !unicode.IsLetter(rune(name[0])) {
		return false
	}
	
	// Check if it contains only valid characters
	pattern := regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)
	return pattern.MatchString(name)
}

// ToSnakeCase converts a name to snake_case
func (c *RustConventionChecker) ToSnakeCase(name string) string {
	// Handle already snake_case
	if c.isSnakeCase(name) {
		return name
	}
	
	// Convert from PascalCase or camelCase
	var result []rune
	for i, r := range name {
		if i > 0 && unicode.IsUpper(r) {
			// Check if previous char is lowercase or next char is lowercase (for acronyms)
			if i > 0 && unicode.IsLower(rune(name[i-1])) {
				result = append(result, '_')
			} else if i < len(name)-1 && unicode.IsLower(rune(name[i+1])) && i > 0 {
				result = append(result, '_')
			}
		}
		result = append(result, unicode.ToLower(r))
	}
	
	return string(result)
}

// ToPascalCase converts a name to PascalCase
func (c *RustConventionChecker) ToPascalCase(name string) string {
	// Handle snake_case
	if strings.Contains(name, "_") {
		parts := strings.Split(name, "_")
		var result strings.Builder
		for _, part := range parts {
			if len(part) > 0 {
				// Handle acronyms properly
				if c.isCommonAcronym(part) {
					result.WriteString(c.formatAcronymForPascalCase(part))
				} else {
					result.WriteString(c.capitalize(part))
				}
			}
		}
		return result.String()
	}
	
	// Handle camelCase - just capitalize first letter
	if len(name) > 0 && unicode.IsLower(rune(name[0])) {
		return string(unicode.ToUpper(rune(name[0]))) + name[1:]
	}
	
	return name
}

// ToScreamingSnakeCase converts a name to SCREAMING_SNAKE_CASE
func (c *RustConventionChecker) ToScreamingSnakeCase(name string) string {
	// First convert to snake_case
	snakeCase := c.ToSnakeCase(name)
	
	// Then convert to uppercase
	return strings.ToUpper(snakeCase)
}

// Helper methods

func (c *RustConventionChecker) isSnakeCase(name string) bool {
	// snake_case: lowercase letters, numbers, and underscores
	// Should not start or end with underscore (except for special cases like _unused)
	// Should not have consecutive underscores
	
	if len(name) == 0 {
		return false
	}
	
	// Allow leading underscore for unused variables
	if strings.HasPrefix(name, "_") && len(name) > 1 {
		name = name[1:]
	}
	
	// Check pattern - allow single letter names as valid snake_case
	pattern := regexp.MustCompile(`^[a-z]([a-z0-9_]*[a-z0-9])?$`)
	if !pattern.MatchString(name) {
		return false
	}
	
	// Check for consecutive underscores
	if strings.Contains(name, "__") {
		return false
	}
	
	return true
}

func (c *RustConventionChecker) isPascalCase(name string) bool {
	// PascalCase: starts with uppercase, followed by letters and numbers
	// No underscores, proper acronym handling
	
	if len(name) == 0 {
		return false
	}
	
	// Must start with uppercase
	if !unicode.IsUpper(rune(name[0])) {
		return false
	}
	
	// Should not contain underscores
	if strings.Contains(name, "_") {
		return false
	}
	
	// Check for proper acronym handling
	// In Rust, acronyms should be like "Http" not "HTTP" when not at the end
	consecutiveUpperCount := 0
	for i, r := range name {
		if unicode.IsUpper(r) {
			consecutiveUpperCount++
			// Allow consecutive uppercase only at the end or for 2-letter acronyms at the start
			if consecutiveUpperCount > 2 && i < len(name)-1 {
				// Check if there's a lowercase letter after
				if i < len(name)-1 && unicode.IsLower(rune(name[i+1])) {
					return false
				}
			}
		} else {
			consecutiveUpperCount = 0
		}
	}
	
	return true
}

func (c *RustConventionChecker) isScreamingSnakeCase(name string) bool {
	// SCREAMING_SNAKE_CASE: uppercase letters, numbers, and underscores
	
	if len(name) == 0 {
		return false
	}
	
	// Check pattern
	pattern := regexp.MustCompile(`^[A-Z][A-Z0-9_]*[A-Z0-9]$|^[A-Z][A-Z0-9]*$`)
	if !pattern.MatchString(name) {
		return false
	}
	
	// Check for consecutive underscores
	if strings.Contains(name, "__") {
		return false
	}
	
	return true
}

func (c *RustConventionChecker) isKeyword(name string) bool {
	return c.keywords[name]
}

func (c *RustConventionChecker) isCommonAcronym(word string) bool {
	upper := strings.ToUpper(word)
	_, exists := c.commonAcronyms[upper]
	return exists
}

func (c *RustConventionChecker) formatAcronymForPascalCase(acronym string) string {
	upper := strings.ToUpper(acronym)
	if formatted, ok := c.commonAcronyms[upper]; ok {
		return formatted
	}
	// Default: capitalize first letter only
	return c.capitalize(acronym)
}

func (c *RustConventionChecker) capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	// Capitalize first letter, lowercase the rest
	return string(unicode.ToUpper(rune(s[0]))) + strings.ToLower(s[1:])
}

// initRustKeywords initializes the set of Rust keywords
func initRustKeywords() map[string]bool {
	return map[string]bool{
		// Strict keywords
		"as": true, "break": true, "const": true, "continue": true,
		"crate": true, "else": true, "enum": true, "extern": true,
		"false": true, "fn": true, "for": true, "if": true,
		"impl": true, "in": true, "let": true, "loop": true,
		"match": true, "mod": true, "move": true, "mut": true,
		"pub": true, "ref": true, "return": true, "self": true,
		"Self": true, "static": true, "struct": true, "super": true,
		"trait": true, "true": true, "type": true, "unsafe": true,
		"use": true, "where": true, "while": true,
		
		// Strict keywords (2018+)
		"async": true, "await": true, "dyn": true,
		
		// Reserved keywords
		"abstract": true, "become": true, "box": true, "do": true,
		"final": true, "macro": true, "override": true, "priv": true,
		"typeof": true, "unsized": true, "virtual": true, "yield": true,
		
		// Weak keywords (context-dependent, but we'll be conservative)
		"union": true, "try": true,
	}
}

// initRustAcronyms initializes common acronyms and their proper casing
func initRustAcronyms() map[string]string {
	return map[string]string{
		// Format: UPPERCASE -> ProperCase for PascalCase context
		"HTTP":  "Http",
		"HTTPS": "Https",
		"URL":   "Url",
		"URI":   "Uri",
		"UUID":  "Uuid",
		"XML":   "Xml",
		"JSON":  "Json",
		"HTML":  "Html",
		"CSS":   "Css",
		"SQL":   "Sql",
		"API":   "Api",
		"REST":  "Rest",
		"RPC":   "Rpc",
		"TCP":   "Tcp",
		"UDP":   "Udp",
		"IP":    "Ip",
		"DNS":   "Dns",
		"TLS":   "Tls",
		"SSL":   "Ssl",
		"SSH":   "Ssh",
		"FTP":   "Ftp",
		"IO":    "Io",
		"OS":    "Os",
		"FS":    "Fs",
		"DB":    "Db",
		"CPU":   "Cpu",
		"GPU":   "Gpu",
		"RAM":   "Ram",
		"ROM":   "Rom",
		"GUI":   "Gui",
		"CLI":   "Cli",
		"TUI":   "Tui",
		"UI":    "Ui",
		"ID":    "Id",
		"UTF":   "Utf",
		"ASCII": "Ascii",
		"YAML":  "Yaml",
		"TOML":  "Toml",
		"CSV":   "Csv",
		"PDF":   "Pdf",
		"PNG":   "Png",
		"JPG":   "Jpg",
		"JPEG":  "Jpeg",
		"GIF":   "Gif",
		"SVG":   "Svg",
	}
}