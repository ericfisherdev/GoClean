// Package morphology provides English morphological analysis for intelligent abbreviation detection.
package morphology

import (
	"fmt"
	"strings"
	"sync"
)

// MorphologyEngine provides morphological analysis capabilities for word decomposition and abbreviation detection
type MorphologyEngine struct {
	prefixDB    map[string]PrefixInfo
	suffixDB    map[string]SuffixInfo
	rootWordDB  map[string]RootInfo
	cache       map[string]*MorphInfo
	cacheMutex  sync.RWMutex
	initialized bool
}

// MorphInfo contains the result of morphological analysis
type MorphInfo struct {
	Prefix      string      `json:"prefix"`
	Root        string      `json:"root"`
	Suffix      string      `json:"suffix"`
	IsComplete  bool        `json:"is_complete"`
	Confidence  float64     `json:"confidence"`
	Morphemes   []Morpheme  `json:"morphemes"`
	WordType    WordCategory `json:"word_type"`
}

// Morpheme represents a single meaningful unit of language
type Morpheme struct {
	Text     string        `json:"text"`
	Type     MorphemeType  `json:"type"`
	Meaning  string        `json:"meaning"`
	Position int           `json:"position"`
}

// MorphemeType defines the type of morpheme
type MorphemeType string

const (
	MorphemeTypePrefix MorphemeType = "prefix"
	MorphemeTypeRoot   MorphemeType = "root"
	MorphemeTypeSuffix MorphemeType = "suffix"
	MorphemeTypeInfix  MorphemeType = "infix"
)

// WordCategory defines the grammatical category of a word
type WordCategory string

const (
	WordCategoryNoun      WordCategory = "noun"
	WordCategoryVerb      WordCategory = "verb"
	WordCategoryAdjective WordCategory = "adjective"
	WordCategoryAdverb    WordCategory = "adverb"
	WordCategoryUnknown   WordCategory = "unknown"
	WordCategoryTechnical WordCategory = "technical"
)

// PrefixInfo contains information about a prefix
type PrefixInfo struct {
	Prefix       string  `json:"prefix"`
	Meaning      string  `json:"meaning"`
	Origin       string  `json:"origin"`
	Productivity float64 `json:"productivity"`
	Examples     []string `json:"examples"`
}

// SuffixInfo contains information about a suffix
type SuffixInfo struct {
	Suffix       string       `json:"suffix"`
	Meaning      string       `json:"meaning"`
	Category     WordCategory `json:"category"`
	Productivity float64      `json:"productivity"`
	Examples     []string     `json:"examples"`
}

// RootInfo contains information about a root word
type RootInfo struct {
	Word       string       `json:"word"`
	Frequency  int          `json:"frequency"`
	Category   WordCategory `json:"category"`
	Variations []string     `json:"variations"`
	IsTechnical bool        `json:"is_technical"`
}

// NewMorphologyEngine creates a new morphology engine with preloaded databases
func NewMorphologyEngine() *MorphologyEngine {
	engine := &MorphologyEngine{
		prefixDB:   make(map[string]PrefixInfo),
		suffixDB:   make(map[string]SuffixInfo),
		rootWordDB: make(map[string]RootInfo),
		cache:      make(map[string]*MorphInfo),
	}
	
	// Initialize databases
	engine.initializeDatabases()
	engine.initialized = true
	
	return engine
}

// AnalyzeWord performs morphological analysis on a word
func (e *MorphologyEngine) AnalyzeWord(word string) *MorphInfo {
	if !e.initialized {
		return &MorphInfo{
			IsComplete: false,
			Confidence: 0.0,
		}
	}
	
	// Normalize cache key to avoid duplicate analyses by case
	key := strings.ToLower(word)

	// Check cache first
	e.cacheMutex.RLock()
	if cached, exists := e.cache[key]; exists {
		e.cacheMutex.RUnlock()
		return cached
	}
	e.cacheMutex.RUnlock()
	
	// Perform analysis
	result := e.performAnalysis(word)
	
	// Cache result
	e.cacheMutex.Lock()
	e.cache[key] = result
	e.cacheMutex.Unlock()
	
	return result
}

// IsCompleteWord determines if a word is morphologically complete (not an abbreviation)
func (e *MorphologyEngine) IsCompleteWord(word string) bool {
	analysis := e.AnalyzeWord(word)
	return analysis.IsComplete && analysis.Confidence > 0.7
}

// IsProbableAbbreviation determines if a word is likely an abbreviation
func (e *MorphologyEngine) IsProbableAbbreviation(word string) bool {
	analysis := e.AnalyzeWord(word)
	
	// Short words with no recognized morphemes are likely abbreviations
	recognizedParts := 0
	if analysis.Prefix != "" {
		if _, ok := e.prefixDB[analysis.Prefix]; ok {
			recognizedParts++
		}
	}
	if analysis.Root != "" {
		if _, ok := e.rootWordDB[analysis.Root]; ok {
			recognizedParts++
		}
	}
	if analysis.Suffix != "" {
		if _, ok := e.suffixDB[analysis.Suffix]; ok {
			recognizedParts++
		}
	}
	if len(word) <= 3 && recognizedParts == 0 {
		return true
	}
	
	// Words with very low confidence are likely abbreviations
	if analysis.Confidence < 0.3 {
		return true
	}
	
	return false
}

// GetSuggestedExpansions returns possible expansions for an abbreviation
func (e *MorphologyEngine) GetSuggestedExpansions(word string) []string {
	var suggestions []string
	
	// Look for roots that start with the abbreviation
	for rootWord, info := range e.rootWordDB {
		if strings.HasPrefix(rootWord, word) && len(rootWord) > len(word) {
			suggestions = append(suggestions, rootWord)
		}
		
		// Check variations
		for _, variation := range info.Variations {
			if strings.HasPrefix(variation, word) && len(variation) > len(word) {
				suggestions = append(suggestions, variation)
			}
		}
	}
	
	return suggestions
}

// performAnalysis conducts the actual morphological analysis
func (e *MorphologyEngine) performAnalysis(word string) *MorphInfo {
	word = strings.ToLower(word)
	
	info := &MorphInfo{
		Confidence: 0.0,
		Morphemes:  []Morpheme{},
		WordType:   WordCategoryUnknown,
	}
	
	// Try to decompose the word
	prefix, root, suffix := e.decomposeWord(word)
	
	info.Prefix = prefix
	info.Root = root
	info.Suffix = suffix
	
	// Calculate confidence based on recognized components
	confidence := e.calculateConfidence(prefix, root, suffix, word)
	info.Confidence = confidence
	
	// Determine if word is complete
	info.IsComplete = e.determineCompleteness(prefix, root, suffix, word, confidence)
	
	// Extract morphemes
	info.Morphemes = e.extractMorphemes(prefix, root, suffix)
	
	// Determine word category
	info.WordType = e.determineWordCategory(suffix, root)
	
	return info
}

// decomposeWord attempts to break a word into prefix, root, and suffix
func (e *MorphologyEngine) decomposeWord(word string) (prefix, root, suffix string) {
	// Try different prefix lengths
	for prefixLen := 1; prefixLen <= len(word)-1; prefixLen++ {
		potentialPrefix := word[:prefixLen]
		if _, exists := e.prefixDB[potentialPrefix]; exists {
			remainder := word[prefixLen:]
			
			// Try to find suffix in remainder
			for suffixLen := 1; suffixLen <= len(remainder)-1; suffixLen++ {
				potentialSuffix := remainder[len(remainder)-suffixLen:]
				if _, exists := e.suffixDB[potentialSuffix]; exists {
					potentialRoot := remainder[:len(remainder)-suffixLen]
					if len(potentialRoot) >= 2 { // Minimum root length
						return potentialPrefix, potentialRoot, potentialSuffix
					}
				}
			}
			
			// No suffix found, check if remainder is a valid root
			if _, exists := e.rootWordDB[remainder]; exists {
				return potentialPrefix, remainder, ""
			}
		}
	}
	
	// Try suffix-only decomposition
	for suffixLen := 1; suffixLen <= len(word)-2; suffixLen++ {
		potentialSuffix := word[len(word)-suffixLen:]
		if _, exists := e.suffixDB[potentialSuffix]; exists {
			potentialRoot := word[:len(word)-suffixLen]
			if _, exists := e.rootWordDB[potentialRoot]; exists {
				return "", potentialRoot, potentialSuffix
			}
			// Special case for "handl" -> "handle" (drop 'e' variants)
			if potentialRoot+"e" != "" {
				if _, exists := e.rootWordDB[potentialRoot+"e"]; exists {
					return "", potentialRoot+"e", potentialSuffix
				}
			}
		}
	}
	
	// Check if entire word is a known root
	if _, exists := e.rootWordDB[word]; exists {
		return "", word, ""
	}
	
	// No decomposition found
	return "", word, ""
}

// calculateConfidence determines confidence in the morphological analysis
func (e *MorphologyEngine) calculateConfidence(prefix, root, suffix, originalWord string) float64 {
	confidence := 0.0
	
	// Special boost for common programming terms
	if e.isCommonProgrammingTerm(originalWord) {
		confidence += 0.75 // High base confidence for programming terms
	}
	
	// Base confidence for recognized components
	if prefix != "" {
		if info, exists := e.prefixDB[prefix]; exists {
			confidence += 0.3 * info.Productivity
		}
	}
	
	if root != "" {
		if info, exists := e.rootWordDB[root]; exists {
			// Higher confidence for more frequent words
			freqBonus := float64(info.Frequency) / 10000.0
			if freqBonus > 0.4 {
				freqBonus = 0.4
			}
			confidence += 0.5 + freqBonus
		}
	}
	
	if suffix != "" {
		if info, exists := e.suffixDB[suffix]; exists {
			confidence += 0.2 * info.Productivity
		}
	}
	
	// Penalty for very short words that aren't complete decompositions
	if len(originalWord) <= 3 && (prefix == "" || root == "" || len(root) <= 1) {
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

// determineCompleteness decides if a word is morphologically complete
func (e *MorphologyEngine) determineCompleteness(prefix, root, suffix, originalWord string, confidence float64) bool {
	// Common programming terms are always considered complete
	if e.isCommonProgrammingTerm(originalWord) {
		return true
	}
	
	// High confidence indicates a complete word
	if confidence > 0.7 {
		return true
	}
	
	// Known root words are likely complete
	if root != "" {
		if info, exists := e.rootWordDB[root]; exists && info.Frequency > 100 {
			return true
		}
	}
	
	// Words longer than 6 characters with some recognizable morphology
	if len(originalWord) > 6 && confidence > 0.4 {
		return true
	}
	
	return false
}

// extractMorphemes creates a list of morphemes from the decomposition
func (e *MorphologyEngine) extractMorphemes(prefix, root, suffix string) []Morpheme {
	var morphemes []Morpheme
	position := 0
	
	if prefix != "" {
		if info, exists := e.prefixDB[prefix]; exists {
			morphemes = append(morphemes, Morpheme{
				Text:     prefix,
				Type:     MorphemeTypePrefix,
				Meaning:  info.Meaning,
				Position: position,
			})
			position += len(prefix)
		}
	}
	
	if root != "" {
		meaning := "base word"
		if info, exists := e.rootWordDB[root]; exists {
			meaning = fmt.Sprintf("%s (category: %s)", root, info.Category)
		}
		morphemes = append(morphemes, Morpheme{
			Text:     root,
			Type:     MorphemeTypeRoot,
			Meaning:  meaning,
			Position: position,
		})
		position += len(root)
	}
	
	if suffix != "" {
		if info, exists := e.suffixDB[suffix]; exists {
			morphemes = append(morphemes, Morpheme{
				Text:     suffix,
				Type:     MorphemeTypeSuffix,
				Meaning:  info.Meaning,
				Position: position,
			})
		}
	}
	
	return morphemes
}

// determineWordCategory determines the grammatical category of the word
func (e *MorphologyEngine) determineWordCategory(suffix, root string) WordCategory {
	// Check suffix first
	if suffix != "" {
		if info, exists := e.suffixDB[suffix]; exists {
			return info.Category
		}
	}
	
	// Check root
	if root != "" {
		if info, exists := e.rootWordDB[root]; exists {
			return info.Category
		}
	}
	
	return WordCategoryUnknown
}

// isCommonProgrammingTerm checks if a word is a common programming term
func (e *MorphologyEngine) isCommonProgrammingTerm(word string) bool {
	commonTerms := map[string]bool{
		"config":        true,
		"configuration": true,
		"database":      true,
		"server":        true,
		"client":        true,
		"request":       true,
		"response":      true,
		"handler":       true,
		"manager":       true,
		"service":       true,
		"interface":     true,
		"struct":        true,
		"function":      true,
		"method":        true,
		"variable":      true,
		"constant":      true,
		"parameter":     true,
		"argument":      true,
		"return":        true,
		"process":       true,
		"processor":     true,
		"execute":       true,
		"executor":      true,
		"initialize":    true,
		"initializer":   true,
		"validate":      true,
		"validator":     true,
		"generate":      true,
		"generator":     true,
		"calculate":     true,
		"calculator":    true,
		"transform":     true,
		"transformer":   true,
		"convert":       true,
		"converter":     true,
		"parse":         true,
		"parser":        true,
		"serialize":     true,
		"serializer":    true,
		"deserialize":   true,
		"deserializer":  true,
	}
	
	return commonTerms[strings.ToLower(word)]
}

// ClearCache clears the analysis cache
func (e *MorphologyEngine) ClearCache() {
	e.cacheMutex.Lock()
	defer e.cacheMutex.Unlock()
	e.cache = make(map[string]*MorphInfo)
}

// GetStats returns statistics about the morphology databases
func (e *MorphologyEngine) GetStats() map[string]int {
	return map[string]int{
		"prefixes":    len(e.prefixDB),
		"suffixes":    len(e.suffixDB),
		"root_words":  len(e.rootWordDB),
		"cached_analyses": len(e.cache),
	}
}