// Package morphology provides programming-specific term analysis for code contexts.
package morphology

import (
	"fmt"
	"strings"
	"unicode"
)

// ProgrammingTermAnalyzer provides specialized analysis for programming terminology
type ProgrammingTermAnalyzer struct {
	morphEngine *MorphologyEngine
	acronyms    map[string]string
	commonTerms map[string]bool
}

// NewProgrammingTermAnalyzer creates a new programming term analyzer
func NewProgrammingTermAnalyzer(morphEngine *MorphologyEngine) *ProgrammingTermAnalyzer {
	analyzer := &ProgrammingTermAnalyzer{
		morphEngine: morphEngine,
		acronyms:    make(map[string]string),
		commonTerms: make(map[string]bool),
	}

	analyzer.initializeTermDatabases()
	return analyzer
}

// AnalyzeProgrammingTerm performs specialized analysis for programming terms
func (p *ProgrammingTermAnalyzer) AnalyzeProgrammingTerm(term string) *ProgrammingTermResult {
	result := &ProgrammingTermResult{
		Term:           term,
		IsAcronym:      p.isKnownAcronym(term),
		IsCommonTerm:   p.isCommonProgrammingTerm(term),
		IsCamelCase:    p.isCamelCase(term),
		IsPascalCase:   p.isPascalCase(term),
		IsSnakeCase:    p.isSnakeCase(term),
		IsKebabCase:    p.isKebabCase(term),
		WordComponents: p.extractWordComponents(term),
		SuggestedFixes: []string{},
		Confidence:     0.0,
	}

	// Populate acronym expansion if applicable
	if result.IsAcronym {
		if exp, ok := p.acronyms[strings.ToUpper(term)]; ok {
			result.AcronymExpansion = exp
		}
	}

	// Perform morphological analysis if not an acronym
	if !result.IsAcronym && p.morphEngine != nil {
		for _, component := range result.WordComponents {
			morphInfo := p.morphEngine.AnalyzeWord(component)
			result.MorphologicalInfo = append(result.MorphologicalInfo, morphInfo)
		}
	}

	// Calculate overall confidence
	result.Confidence = p.calculateTermConfidence(result)

	// Generate suggestions if needed
	if result.Confidence < 0.5 {
		result.SuggestedFixes = p.generateSuggestions(term, result)
	}

	return result
}

// ProgrammingTermResult contains the analysis result for a programming term
type ProgrammingTermResult struct {
	Term              string       `json:"term"`
	IsAcronym         bool         `json:"is_acronym"`
	IsCommonTerm      bool         `json:"is_common_term"`
	IsCamelCase       bool         `json:"is_camel_case"`
	IsPascalCase      bool         `json:"is_pascal_case"`
	IsSnakeCase       bool         `json:"is_snake_case"`
	IsKebabCase       bool         `json:"is_kebab_case"`
	WordComponents    []string     `json:"word_components"`
	MorphologicalInfo []*MorphInfo `json:"morphological_info"`
	SuggestedFixes    []string     `json:"suggested_fixes"`
	Confidence        float64      `json:"confidence"`
	AcronymExpansion  string       `json:"acronym_expansion,omitempty"`
}

// initializeTermDatabases loads programming-specific databases
func (p *ProgrammingTermAnalyzer) initializeTermDatabases() {
	p.loadAcronymDatabase()
	p.loadCommonTermsDatabase()
}

// loadAcronymDatabase initializes known programming acronyms
func (p *ProgrammingTermAnalyzer) loadAcronymDatabase() {
	acronyms := map[string]string{
		// Web and networking
		"HTTP":  "HyperText Transfer Protocol",
		"HTTPS": "HyperText Transfer Protocol Secure",
		"URL":   "Uniform Resource Locator",
		"URI":   "Uniform Resource Identifier",
		"API":   "Application Programming Interface",
		"REST":  "Representational State Transfer",
		"SOAP":  "Simple Object Access Protocol",
		"JSON":  "JavaScript Object Notation",
		"XML":   "eXtensible Markup Language",
		"HTML":  "HyperText Markup Language",
		"CSS":   "Cascading Style Sheets",
		"DNS":   "Domain Name System",
		"TCP":   "Transmission Control Protocol",
		"UDP":   "User Datagram Protocol",
		"IP":    "Internet Protocol",
		"FTP":   "File Transfer Protocol",
		"SSH":   "Secure Shell",
		"SSL":   "Secure Sockets Layer",
		"TLS":   "Transport Layer Security",

		// Databases
		"SQL":   "Structured Query Language",
		"CRUD":  "Create, Read, Update, Delete",
		"ACID":  "Atomicity, Consistency, Isolation, Durability",
		"ORM":   "Object-Relational Mapping",
		"DBMS":  "Database Management System",
		"NoSQL": "Not Only SQL",

		// Programming concepts
		"OOP":    "Object-Oriented Programming",
		"MVC":    "Model-View-Controller",
		"MVP":    "Model-View-Presenter",
		"MVVM":   "Model-View-ViewModel",
		"DRY":    "Don't Repeat Yourself",
		"SOLID":  "Single Responsibility, Open-Closed, Liskov Substitution, Interface Segregation, Dependency Inversion",
		"KISS":   "Keep It Simple, Stupid",
		"YAGNI":  "You Aren't Gonna Need It",
		"TDD":    "Test-Driven Development",
		"BDD":    "Behavior-Driven Development",
		"CI":     "Continuous Integration",
		"CD":     "Continuous Deployment",
		"DevOps": "Development Operations",

		// Data structures and algorithms
		"AST":  "Abstract Syntax Tree",
		"BFS":  "Breadth-First Search",
		"DFS":  "Depth-First Search",
		"LIFO": "Last In, First Out",
		"FIFO": "First In, First Out",
		"LRU":  "Least Recently Used",
		"UUID": "Universally Unique Identifier",
		"GUID": "Globally Unique Identifier",

		// File formats and encoding
		"CSV":   "Comma-Separated Values",
		"TSV":   "Tab-Separated Values",
		"UTF":   "Unicode Transformation Format",
		"ASCII": "American Standard Code for Information Interchange",
		"YAML":  "YAML Ain't Markup Language",
		"TOML":  "Tom's Obvious, Minimal Language",
		"PDF":   "Portable Document Format",

		// Version control
		"VCS": "Version Control System",
		"SCM": "Source Code Management",
		"Git": "Global Information Tracker",
		"SVN": "Subversion",
		"CVS": "Concurrent Versions System",

		// Security
		"JWT":   "JSON Web Token",
		"OAuth": "Open Authorization",
		"SAML":  "Security Assertion Markup Language",
		"LDAP":  "Lightweight Directory Access Protocol",
		"AES":   "Advanced Encryption Standard",
		"RSA":   "Rivest-Shamir-Adleman",
		"SHA":   "Secure Hash Algorithm",
		"MD5":   "Message Digest Algorithm 5",

		// Go-specific
		"Ctx":  "Context",
		"Cfg":  "Configuration",
		"Repo": "Repository",
		"Auth": "Authentication",
		"Impl": "Implementation",
	}

	// Canonicalize acronym keys to uppercase for consistent lookup
	p.acronyms = make(map[string]string)
	for k, v := range acronyms {
		p.acronyms[strings.ToUpper(k)] = v
	}
}

// loadCommonTermsDatabase initializes common programming terms
func (p *ProgrammingTermAnalyzer) loadCommonTermsDatabase() {
	commonTerms := map[string]bool{
		// Action verbs
		"add": true, "append": true, "apply": true, "assign": true, "attach": true,
		"build": true, "call": true, "check": true, "clear": true, "clone": true,
		"close": true, "compare": true, "compile": true, "compute": true, "connect": true,
		"convert": true, "copy": true, "create": true, "decode": true, "delete": true,
		"deploy": true, "destroy": true, "detach": true, "disconnect": true, "encode": true,
		"execute": true, "export": true, "extract": true, "fetch": true, "filter": true,
		"find": true, "format": true, "generate": true, "get": true, "handle": true,
		"import": true, "initialize": true, "insert": true, "install": true, "invoke": true,
		"join": true, "launch": true, "load": true, "lock": true, "log": true,
		"merge": true, "modify": true, "move": true, "notify": true, "open": true,
		"parse": true, "process": true, "publish": true, "push": true, "read": true,
		"receive": true, "register": true, "remove": true, "replace": true, "reset": true,
		"resolve": true, "retrieve": true, "return": true, "run": true, "save": true,
		"search": true, "send": true, "serialize": true, "set": true, "sort": true,
		"split": true, "start": true, "stop": true, "store": true, "submit": true,
		"sync": true, "transform": true, "update": true, "upload": true, "validate": true,
		"write": true,

		// State verbs
		"is": true, "has": true, "can": true, "should": true, "will": true, "was": true,
		"are": true, "have": true, "could": true, "would": true, "been": true,

		// Data types and structures
		"array": true, "list": true, "map": true, "stack": true, "queue": true,
		"tree": true, "graph": true, "node": true, "edge": true, "vertex": true,
		"table": true, "record": true, "field": true, "column": true, "row": true,
		"key": true, "value": true, "pair": true, "entry": true, "item": true,
		"element": true, "object": true, "entity": true, "model": true, "schema": true,

		// Programming concepts
		"class": true, "struct": true, "interface": true, "enum": true, "type": true,
		"function": true, "method": true, "procedure": true, "routine": true, "callback": true,
		"event": true, "handler": true, "listener": true, "observer": true, "subscriber": true,
		"publisher": true, "producer": true, "consumer": true, "worker": true, "task": true,
		"job": true, "thread": true, "goroutine": true,
		"service": true, "component": true, "module": true, "package": true, "library": true,
		"framework": true, "plugin": true, "extension": true, "addon": true, "middleware": true,

		// System concepts
		"file": true, "directory": true, "folder": true, "path": true, "location": true,
		"address": true, "pointer": true, "reference": true, "link": true, "url": true,
		"connection": true, "session": true, "transaction": true, "request": true, "response": true,
		"message": true, "signal": true, "notification": true, "alert": true, "warning": true,
		"error": true, "exception": true, "failure": true, "success": true, "result": true,
		"output": true, "input": true, "stream": true, "buffer": true, "cache": true,
		"memory": true, "storage": true, "disk": true, "network": true, "protocol": true,

		// Quality attributes
		"secure": true, "safe": true, "reliable": true, "stable": true, "robust": true,
		"scalable": true, "flexible": true, "maintainable": true, "testable": true, "portable": true,
		"efficient": true, "optimal": true, "fast": true, "slow": true, "quick": true,
		"valid": true, "invalid": true, "correct": true, "wrong": true, "broken": true,
		"active": true, "inactive": true, "enabled": true, "disabled": true, "visible": true,
		"hidden": true, "public": true, "private": true, "protected": true, "internal": true,

		// Quantities and measurements
		"count": true, "number": true, "amount": true, "quantity": true, "size": true,
		"length": true, "width": true, "height": true, "depth": true, "capacity": true,
		"limit": true, "maximum": true, "minimum": true, "average": true, "total": true,
		"partial": true, "complete": true, "full": true, "empty": true, "available": true,
		"remaining": true, "current": true, "previous": true, "next": true, "last": true,
		"first": true, "initial": true, "final": true, "temporary": true, "permanent": true,

		// Time-related
		"time": true, "date": true, "timestamp": true, "duration": true, "interval": true,
		"timeout": true, "delay": true, "schedule": true, "timer": true, "clock": true,
		"end": true, "begin": true, "finish": true,
		"expire": true,
	}

	p.commonTerms = commonTerms
}

// isKnownAcronym checks if a term is a known programming acronym
func (p *ProgrammingTermAnalyzer) isKnownAcronym(term string) bool {
	upperTerm := strings.ToUpper(term)
	_, exists := p.acronyms[upperTerm]
	return exists
}

// isCommonProgrammingTerm checks if a term is a common programming term
func (p *ProgrammingTermAnalyzer) isCommonProgrammingTerm(term string) bool {
	lowerTerm := strings.ToLower(term)
	return p.commonTerms[lowerTerm]
}

// isCamelCase checks if a term follows camelCase convention
func (p *ProgrammingTermAnalyzer) isCamelCase(term string) bool {
	if len(term) == 0 {
		return false
	}

	runes := []rune(term)
	if len(runes) == 0 {
		return false
	}

	// First character should be lowercase
	if !unicode.IsLower(runes[0]) {
		return false
	}

	// Return false if the term contains separator characters
	if strings.ContainsAny(term, "_-") {
		return false
	}

	// Single lowercase rune is valid camelCase
	if len(runes) == 1 {
		return true
	}

	// Scan for uppercase letters
	hasUpper := false
	for i := 1; i < len(runes); i++ {
		if unicode.IsUpper(runes[i]) {
			hasUpper = true
			break
		}
	}

	// For multi-rune terms, only return true if uppercase was found
	// (multi-rune with no uppercase but no separators should return false)
	return hasUpper
}

// isPascalCase checks if a term follows PascalCase convention
func (p *ProgrammingTermAnalyzer) isPascalCase(term string) bool {
	if len(term) == 0 {
		return false
	}
	runes := []rune(term)
	// First rune must be uppercase, and no separators
	if !unicode.IsUpper(runes[0]) || strings.ContainsAny(term, "_-") {
		return false
	}
	// Avoid treating all-uppercase acronyms as PascalCase
	hasLower := false
	for _, r := range runes[1:] {
		if unicode.IsLower(r) {
			hasLower = true
			break
		}
	}
	return hasLower
}

// isSnakeCase checks if a term follows snake_case convention
func (p *ProgrammingTermAnalyzer) isSnakeCase(term string) bool {
	if len(term) == 0 {
		return false
	}

	// Snake case requires at least one underscore for compound words
	if !strings.Contains(term, "_") {
		return false
	}

	// Should be all lowercase with underscores
	for _, char := range term {
		if !unicode.IsLower(char) && char != '_' && !unicode.IsDigit(char) {
			return false
		}
	}

	// Should not start or end with underscore
	if strings.HasPrefix(term, "_") || strings.HasSuffix(term, "_") {
		return false
	}

	// Should not have consecutive underscores
	if strings.Contains(term, "__") {
		return false
	}

	return true
}

// isKebabCase checks if a term follows kebab-case convention
func (p *ProgrammingTermAnalyzer) isKebabCase(term string) bool {
	if len(term) == 0 {
		return false
	}

	// Kebab case requires at least one hyphen for compound words
	if !strings.Contains(term, "-") {
		return false
	}

	// Should be all lowercase with hyphens
	for _, char := range term {
		if !unicode.IsLower(char) && char != '-' && !unicode.IsDigit(char) {
			return false
		}
	}

	// Should not start or end with hyphen
	if strings.HasPrefix(term, "-") || strings.HasSuffix(term, "-") {
		return false
	}

	// Should not have consecutive hyphens
	if strings.Contains(term, "--") {
		return false
	}

	return true
}

// extractWordComponents extracts individual words from a compound term
func (p *ProgrammingTermAnalyzer) extractWordComponents(term string) []string {
	var components []string

	if p.isSnakeCase(term) {
		components = strings.Split(term, "_")
	} else if p.isKebabCase(term) {
		components = strings.Split(term, "-")
	} else {
		// Handle camelCase and PascalCase with proper Unicode support
		runes := []rune(term)
		if len(runes) == 0 {
			return components
		}

		var currentWord strings.Builder
		for i := 0; i < len(runes); i++ {
			char := runes[i]

			// Check for word boundaries
			shouldSplit := false
			if i > 0 {
				prevChar := runes[i-1]

				// Lower to upper transition (camelCase boundary)
				if unicode.IsLower(prevChar) && unicode.IsUpper(char) {
					shouldSplit = true
				}

				// Acronym handling: Upper to upper followed by lower (HTTPServer -> HTTP Server)
				if i < len(runes)-1 && unicode.IsUpper(prevChar) && unicode.IsUpper(char) && unicode.IsLower(runes[i+1]) {
					shouldSplit = true
				}

				// Letter to digit transition (userID123 -> user ID 123)
				if unicode.IsLetter(prevChar) && unicode.IsDigit(char) {
					shouldSplit = true
				}

				// Digit to letter transition (SHA256Sum -> SHA 256 Sum)
				if unicode.IsDigit(prevChar) && unicode.IsLetter(char) {
					shouldSplit = true
				}
			}

			if shouldSplit && currentWord.Len() > 0 {
				components = append(components, strings.ToLower(currentWord.String()))
				currentWord.Reset()
			}

			currentWord.WriteRune(unicode.ToLower(char))
		}

		if currentWord.Len() > 0 {
			components = append(components, currentWord.String())
		}
	}

	return components
}

// calculateTermConfidence calculates confidence score for a programming term
func (p *ProgrammingTermAnalyzer) calculateTermConfidence(result *ProgrammingTermResult) float64 {
	confidence := 0.0

	// Acronyms get high confidence if they're known
	if result.IsAcronym {
		if p.isKnownAcronym(result.Term) {
			return 0.95
		}
		return 0.3 // Unknown acronyms are suspicious
	}

	// Common terms get baseline confidence
	if result.IsCommonTerm {
		confidence += 0.4
	}

	// Proper naming convention adds confidence
	if result.IsCamelCase || result.IsSnakeCase || result.IsPascalCase {
		confidence += 0.3
	}

	// Morphological analysis of components
	if len(result.MorphologicalInfo) > 0 {
		totalMorphConfidence := 0.0
		for _, morphInfo := range result.MorphologicalInfo {
			totalMorphConfidence += morphInfo.Confidence
		}
		avgMorphConfidence := totalMorphConfidence / float64(len(result.MorphologicalInfo))
		confidence += 0.4 * avgMorphConfidence
	}

	// Penalty for very short terms that aren't acronyms
	if len(result.Term) <= 3 && !result.IsAcronym && !result.IsCommonTerm {
		confidence *= 0.5
	}

	// Ensure confidence is within bounds
	if confidence > 1.0 {
		confidence = 1.0
	}
	if confidence < 0.0 {
		confidence = 0.0
	}

	return confidence
}

// generateSuggestions generates improvement suggestions for low-confidence terms
func (p *ProgrammingTermAnalyzer) generateSuggestions(term string, result *ProgrammingTermResult) []string {
	var suggestions []string

	// If it's a potential acronym, suggest expansions
	if len(term) <= 5 && strings.ToUpper(term) == term {
		if expansion, exists := p.acronyms[term]; exists {
			suggestions = append(suggestions, fmt.Sprintf("Consider using full form: %s", expansion))
		} else {
			suggestions = append(suggestions, "Consider using a more descriptive name instead of this acronym")
		}
	}

	// If components have low morphological confidence, suggest alternatives
	for i, component := range result.WordComponents {
		if i < len(result.MorphologicalInfo) && result.MorphologicalInfo[i].Confidence < 0.5 {
			expansions := p.morphEngine.GetSuggestedExpansions(component)
			for _, expansion := range expansions {
				if len(expansions) > 0 && len(expansion) > len(component) {
					suggestions = append(suggestions, fmt.Sprintf("Consider '%s' instead of '%s'", expansion, component))
				}
			}
		}
	}

	// If naming convention is inconsistent
	if !result.IsCamelCase && !result.IsSnakeCase && !result.IsKebabCase && !result.IsPascalCase {
		suggestions = append(suggestions, "Consider using camelCase, PascalCase, or snake_case naming convention (per context)")
	}

	// If term is very short and not descriptive
	if len(term) <= 3 && !result.IsCommonTerm && !result.IsAcronym {
		suggestions = append(suggestions, "Consider using a more descriptive name")
	}

	return suggestions
}
