// Package morphology provides test coverage for the morphological analysis engine.
package morphology

import (
	"testing"
)

func TestMorphologyEngine_NewMorphologyEngine(t *testing.T) {
	engine := NewMorphologyEngine()
	
	if engine == nil {
		t.Fatal("Expected non-nil morphology engine")
	}
	
	if !engine.initialized {
		t.Error("Expected engine to be initialized")
	}
	
	stats := engine.GetStats()
	if stats["prefixes"] == 0 {
		t.Error("Expected prefixes to be loaded")
	}
	if stats["suffixes"] == 0 {
		t.Error("Expected suffixes to be loaded")
	}
	if stats["root_words"] == 0 {
		t.Error("Expected root words to be loaded")
	}
}

func TestMorphologyEngine_AnalyzeWord_CompleteWords(t *testing.T) {
	engine := NewMorphologyEngine()
	
	testCases := []struct {
		word               string
		expectedComplete   bool
		expectedConfidence float64
		description        string
	}{
		{"handler", true, 0.7, "handler should be recognized as complete"},
		{"manager", true, 0.7, "manager should be recognized as complete"},
		{"processor", true, 0.7, "processor should be recognized as complete"},
		{"validator", true, 0.7, "validator should be recognized as complete"},
		{"generator", true, 0.7, "generator should be recognized as complete"},
		{"configuration", true, 0.8, "configuration should be recognized as complete"},
		{"initialize", true, 0.8, "initialize should be recognized as complete"},
		{"serialize", true, 0.7, "serialize should be recognized as complete"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.word, func(t *testing.T) {
			result := engine.AnalyzeWord(tc.word)
			
			if result.IsComplete != tc.expectedComplete {
				t.Errorf("For word '%s': expected IsComplete=%v, got %v", 
					tc.word, tc.expectedComplete, result.IsComplete)
			}
			
			if result.Confidence < tc.expectedConfidence {
				t.Errorf("For word '%s': expected confidence >= %v, got %v", 
					tc.word, tc.expectedConfidence, result.Confidence)
			}
		})
	}
}

func TestMorphologyEngine_AnalyzeWord_Abbreviations(t *testing.T) {
	engine := NewMorphologyEngine()
	
	testCases := []struct {
		word        string
		description string
	}{
		{"cfg", "cfg should be recognized as abbreviation"},
		{"mgr", "mgr should be recognized as abbreviation"},
		{"proc", "proc should be recognized as abbreviation"},
		{"req", "req should be recognized as abbreviation"},
		{"res", "res should be recognized as abbreviation"},
		{"btn", "btn should be recognized as abbreviation"},
		{"str", "str should be recognized as abbreviation"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.word, func(t *testing.T) {
			result := engine.AnalyzeWord(tc.word)
			
			if result.IsComplete {
				t.Errorf("For word '%s': expected to be recognized as abbreviation, but IsComplete=true", tc.word)
			}
			
			if result.Confidence > 0.5 {
				t.Errorf("For word '%s': expected low confidence for abbreviation, got %v", 
					tc.word, result.Confidence)
			}
		})
	}
}

func TestMorphologyEngine_IsCompleteWord(t *testing.T) {
	engine := NewMorphologyEngine()
	
	testCases := []struct {
		word     string
		expected bool
	}{
		{"handler", true},
		{"manager", true},
		{"configuration", true},
		{"cfg", false},
		{"mgr", false},
		{"x", false},
		{"a", false},
		{"process", true},
		{"data", true},
	}
	
	for _, tc := range testCases {
		t.Run(tc.word, func(t *testing.T) {
			result := engine.IsCompleteWord(tc.word)
			if result != tc.expected {
				t.Errorf("For word '%s': expected %v, got %v", tc.word, tc.expected, result)
			}
		})
	}
}

func TestMorphologyEngine_IsProbableAbbreviation(t *testing.T) {
	engine := NewMorphologyEngine()
	
	testCases := []struct {
		word     string
		expected bool
	}{
		{"cfg", true},
		{"mgr", true},
		{"req", true},
		{"handler", false},
		{"manager", false},
		{"configuration", false},
		{"x", true},
		{"xy", true},
		{"abc", true},
	}
	
	for _, tc := range testCases {
		t.Run(tc.word, func(t *testing.T) {
			result := engine.IsProbableAbbreviation(tc.word)
			if result != tc.expected {
				t.Errorf("For word '%s': expected %v, got %v", tc.word, tc.expected, result)
			}
		})
	}
}

func TestMorphologyEngine_GetSuggestedExpansions(t *testing.T) {
	engine := NewMorphologyEngine()
	
	testCases := []struct {
		word              string
		expectedMinLength int
		description       string
	}{
		{"hand", 1, "hand should have handler as expansion"},
		{"manage", 1, "manage should have manager/management as expansions"},
		{"config", 1, "config should have configuration as expansion"},
		{"xyz", 0, "xyz should have no valid expansions"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.word, func(t *testing.T) {
			expansions := engine.GetSuggestedExpansions(tc.word)
			
			if len(expansions) < tc.expectedMinLength {
				t.Errorf("For word '%s': expected at least %d expansions, got %d: %v", 
					tc.word, tc.expectedMinLength, len(expansions), expansions)
			}
			
			// All expansions should be longer than the original word
			for _, expansion := range expansions {
				if len(expansion) <= len(tc.word) {
					t.Errorf("For word '%s': expansion '%s' should be longer than original word", 
						tc.word, expansion)
				}
			}
		})
	}
}

func TestMorphologyEngine_Cache(t *testing.T) {
	engine := NewMorphologyEngine()
	
	word := "handler"
	
	// First analysis
	result1 := engine.AnalyzeWord(word)
	
	// Second analysis (should use cache)
	result2 := engine.AnalyzeWord(word)
	stats2 := engine.GetStats()
	
	// Results should be identical
	if result1.Confidence != result2.Confidence {
		t.Error("Cached result should be identical to original")
	}
	
	// Cache should have entries
	if stats2["cached_analyses"] == 0 {
		t.Error("Expected cache to contain analyzed words")
	}
	
	// Clear cache
	engine.ClearCache()
	stats3 := engine.GetStats()
	
	if stats3["cached_analyses"] != 0 {
		t.Error("Expected cache to be empty after clearing")
	}
}

func TestMorphologyEngine_MorphemeExtraction(t *testing.T) {
	engine := NewMorphologyEngine()
	
	testCases := []struct {
		word            string
		expectedMinMorphemes int
		description     string
	}{
		{"handler", 2, "handler should have hand + er morphemes"},
		{"processor", 2, "processor should have process + or morphemes"},
		{"configuration", 2, "configuration should have config + ation morphemes"},
		{"initialize", 2, "initialize should have initial + ize morphemes"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.word, func(t *testing.T) {
			result := engine.AnalyzeWord(tc.word)
			
			if len(result.Morphemes) < tc.expectedMinMorphemes {
				t.Errorf("For word '%s': expected at least %d morphemes, got %d: %v", 
					tc.word, tc.expectedMinMorphemes, len(result.Morphemes), result.Morphemes)
			}
			
			// Check that morphemes are properly positioned
			for i, morpheme := range result.Morphemes {
				if i > 0 && morpheme.Position <= result.Morphemes[i-1].Position {
					t.Errorf("For word '%s': morpheme positions should be increasing", tc.word)
				}
			}
		})
	}
}

func TestProgrammingTermAnalyzer_AnalyzeProgrammingTerm(t *testing.T) {
	engine := NewMorphologyEngine()
	analyzer := NewProgrammingTermAnalyzer(engine)
	
	testCases := []struct {
		term               string
		expectedIsAcronym  bool
		expectedCommonTerm bool
		expectedCamelCase  bool
		expectedSnakeCase  bool
		expectedComponents int
		description        string
	}{
		{"getUserById", false, false, true, false, 4, "getUserById should be camelCase with 4 components"},
		{"HTTP", true, false, false, false, 1, "HTTP should be recognized as acronym"},
		{"user_id", false, false, false, true, 2, "user_id should be snake_case with 2 components"},
		{"validate", false, true, false, false, 1, "validate should be common term"},
		{"XMLHttpRequest", false, false, false, false, 3, "XMLHttpRequest should have 3 components"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.term, func(t *testing.T) {
			result := analyzer.AnalyzeProgrammingTerm(tc.term)
			
			if result.IsAcronym != tc.expectedIsAcronym {
				t.Errorf("For term '%s': expected IsAcronym=%v, got %v", 
					tc.term, tc.expectedIsAcronym, result.IsAcronym)
			}
			
			if result.IsCommonTerm != tc.expectedCommonTerm {
				t.Errorf("For term '%s': expected IsCommonTerm=%v, got %v", 
					tc.term, tc.expectedCommonTerm, result.IsCommonTerm)
			}
			
			if result.IsCamelCase != tc.expectedCamelCase {
				t.Errorf("For term '%s': expected IsCamelCase=%v, got %v", 
					tc.term, tc.expectedCamelCase, result.IsCamelCase)
			}
			
			if result.IsSnakeCase != tc.expectedSnakeCase {
				t.Errorf("For term '%s': expected IsSnakeCase=%v, got %v", 
					tc.term, tc.expectedSnakeCase, result.IsSnakeCase)
			}
			
			if len(result.WordComponents) != tc.expectedComponents {
				t.Errorf("For term '%s': expected %d components, got %d: %v", 
					tc.term, tc.expectedComponents, len(result.WordComponents), result.WordComponents)
			}
		})
	}
}

func TestProgrammingTermAnalyzer_NamingConventions(t *testing.T) {
	engine := NewMorphologyEngine()
	analyzer := NewProgrammingTermAnalyzer(engine)
	
	testCases := []struct {
		term     string
		camelCase bool
		snakeCase bool
		kebabCase bool
	}{
		{"getUserId", true, false, false},
		{"GetUserID", false, false, false}, // PascalCase
		{"get_user_id", false, true, false},
		{"get-user-id", false, false, true},
		{"getuserid", false, false, false}, // No case
		{"GET_USER_ID", false, false, false}, // SCREAMING_SNAKE_CASE
	}
	
	for _, tc := range testCases {
		t.Run(tc.term, func(t *testing.T) {
			result := analyzer.AnalyzeProgrammingTerm(tc.term)
			
			if result.IsCamelCase != tc.camelCase {
				t.Errorf("For term '%s': expected IsCamelCase=%v, got %v", 
					tc.term, tc.camelCase, result.IsCamelCase)
			}
			
			if result.IsSnakeCase != tc.snakeCase {
				t.Errorf("For term '%s': expected IsSnakeCase=%v, got %v", 
					tc.term, tc.snakeCase, result.IsSnakeCase)
			}
			
			if result.IsKebabCase != tc.kebabCase {
				t.Errorf("For term '%s': expected IsKebabCase=%v, got %v", 
					tc.term, tc.kebabCase, result.IsKebabCase)
			}
		})
	}
}

func TestProgrammingTermAnalyzer_AcronymRecognition(t *testing.T) {
	engine := NewMorphologyEngine()
	analyzer := NewProgrammingTermAnalyzer(engine)
	
	knownAcronyms := []string{"HTTP", "HTTPS", "URL", "API", "JSON", "XML", "SQL", "UUID"}
	unknownAcronyms := []string{"XYZ", "ABC", "DEF"}
	
	for _, acronym := range knownAcronyms {
		t.Run(acronym, func(t *testing.T) {
			result := analyzer.AnalyzeProgrammingTerm(acronym)
			
			if !result.IsAcronym {
				t.Errorf("Expected '%s' to be recognized as known acronym", acronym)
			}
			
			if result.Confidence < 0.9 {
				t.Errorf("Expected high confidence for known acronym '%s', got %v", 
					acronym, result.Confidence)
			}
		})
	}
	
	for _, acronym := range unknownAcronyms {
		t.Run(acronym, func(t *testing.T) {
			result := analyzer.AnalyzeProgrammingTerm(acronym)
			
			if result.Confidence > 0.5 {
				t.Errorf("Expected low confidence for unknown acronym '%s', got %v", 
					acronym, result.Confidence)
			}
		})
	}
}

func TestProgrammingTermAnalyzer_SuggestionGeneration(t *testing.T) {
	engine := NewMorphologyEngine()
	analyzer := NewProgrammingTermAnalyzer(engine)
	
	testCases := []struct {
		term                   string
		expectedSuggestions    bool
		description            string
	}{
		{"cfg", true, "cfg should get suggestions"},
		{"mgr", true, "mgr should get suggestions"},
		{"x", true, "single letter should get suggestions"},
		{"getUserById", false, "good camelCase should not need suggestions"},
		{"validate", false, "common terms should not need suggestions"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.term, func(t *testing.T) {
			result := analyzer.AnalyzeProgrammingTerm(tc.term)
			
			hasSuggestions := len(result.SuggestedFixes) > 0
			if hasSuggestions != tc.expectedSuggestions {
				t.Errorf("For term '%s': expected suggestions=%v, got %v (suggestions: %v)", 
					tc.term, tc.expectedSuggestions, hasSuggestions, result.SuggestedFixes)
			}
		})
	}
}

// Benchmark tests for performance validation
func BenchmarkMorphologyEngine_AnalyzeWord(b *testing.B) {
	engine := NewMorphologyEngine()
	word := "handler"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.AnalyzeWord(word)
	}
}

func BenchmarkMorphologyEngine_AnalyzeWordCached(b *testing.B) {
	engine := NewMorphologyEngine()
	word := "handler"
	
	// Warm up cache
	engine.AnalyzeWord(word)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.AnalyzeWord(word)
	}
}

func BenchmarkProgrammingTermAnalyzer_AnalyzeTerm(b *testing.B) {
	engine := NewMorphologyEngine()
	analyzer := NewProgrammingTermAnalyzer(engine)
	term := "getUserById"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyzer.AnalyzeProgrammingTerm(term)
	}
}

// TestProgrammingTermAnalyzer_UnicodeSupport tests Unicode handling in camelCase detection and word component extraction
func TestProgrammingTermAnalyzer_UnicodeSupport(t *testing.T) {
	engine := NewMorphologyEngine()
	analyzer := NewProgrammingTermAnalyzer(engine)
	
	testCases := []struct {
		term               string
		expectedComponents []string
		description        string
	}{
		{"GödelNumber", []string{"gödel", "number"}, "German umlaut should be handled correctly"},
		{"ΔTime", []string{"δ", "time"}, "Greek letters should be handled correctly"},
		{"naïveBayes", []string{"naïve", "bayes"}, "French accent should be handled correctly"},
		{"HTTPServer", []string{"http", "server"}, "Acronym followed by word should split correctly"},
		{"userID", []string{"user", "id"}, "Word followed by acronym should split correctly"},
		{"SHA256Sum", []string{"sha", "256", "sum"}, "Acronym-number-word should split correctly"},
		{"userID123", []string{"user", "id", "123"}, "Word-acronym-number should split correctly"},
		{"getUser2Factor", []string{"get", "user", "2", "factor"}, "Number in middle should create boundary"},
		{"XMLHttpRequest", []string{"xml", "http", "request"}, "Multiple acronyms should split correctly"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.term, func(t *testing.T) {
			result := analyzer.AnalyzeProgrammingTerm(tc.term)
			
			if len(result.WordComponents) != len(tc.expectedComponents) {
				t.Errorf("For term '%s': expected %d components %v, got %d: %v", 
					tc.term, len(tc.expectedComponents), tc.expectedComponents, 
					len(result.WordComponents), result.WordComponents)
				return
			}
			
			for i, expected := range tc.expectedComponents {
				if i < len(result.WordComponents) && result.WordComponents[i] != expected {
					t.Errorf("For term '%s': expected component[%d]='%s', got '%s'", 
						tc.term, i, expected, result.WordComponents[i])
				}
			}
		})
	}
}

// TestProgrammingTermAnalyzer_CamelCaseUnicode tests Unicode support in camelCase detection
func TestProgrammingTermAnalyzer_CamelCaseUnicode(t *testing.T) {
	engine := NewMorphologyEngine()
	analyzer := NewProgrammingTermAnalyzer(engine)
	
	testCases := []struct {
		term           string
		expectedCamel  bool
		description    string
	}{
		{"naïveBayes", true, "French accent in camelCase should be detected"},
		{"getUserNaïve", true, "camelCase with Unicode should be detected"},
		{"δTime", true, "Lowercase Unicode first character should be camelCase"},
		{"gödelNumber", true, "Lowercase Unicode first character should be camelCase"},
		{"ΔTime", false, "Uppercase Greek first character should not be camelCase"},
		{"αβγTest", true, "Lowercase Greek followed by uppercase should be camelCase"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.term, func(t *testing.T) {
			result := analyzer.AnalyzeProgrammingTerm(tc.term)
			
			if result.IsCamelCase != tc.expectedCamel {
				t.Errorf("For term '%s': expected IsCamelCase=%v, got %v", 
					tc.term, tc.expectedCamel, result.IsCamelCase)
			}
		})
	}
}

// TestProgrammingTermAnalyzer_CamelCaseFixedLogic tests the fixed camelCase logic for separators and single characters
func TestProgrammingTermAnalyzer_CamelCaseFixedLogic(t *testing.T) {
	engine := NewMorphologyEngine()
	analyzer := NewProgrammingTermAnalyzer(engine)
	
	testCases := []struct {
		term           string
		expectedCamel  bool
		description    string
	}{
		// Single character cases
		{"a", true, "Single lowercase letter should be camelCase"},
		{"x", true, "Single lowercase letter should be camelCase"},
		{"A", false, "Single uppercase letter should not be camelCase"},
		
		// Separator cases (should be false)
		{"foo_bar", false, "Terms with underscores should not be camelCase"},
		{"foo-bar", false, "Terms with hyphens should not be camelCase"},
		{"user_id", false, "snake_case should not be camelCase"},
		{"get-user", false, "kebab-case should not be camelCase"},
		
		// Multi-character without uppercase (should be false)
		{"foobar", false, "Multi-character lowercase without uppercase should not be camelCase"},
		{"hello", false, "Multi-character lowercase without uppercase should not be camelCase"},
		{"test", false, "Multi-character lowercase without uppercase should not be camelCase"},
		
		// Valid camelCase (should be true)
		{"getUserId", true, "Proper camelCase should be detected"},
		{"camelCase", true, "Proper camelCase should be detected"},
		{"myVariable", true, "Proper camelCase should be detected"},
		{"a1", false, "Single letter followed by digit should not be camelCase without uppercase"},
		
		// Invalid camelCase (should be false)  
		{"GetUserId", false, "PascalCase should not be camelCase"},
		{"CONSTANT", false, "All uppercase should not be camelCase"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.term, func(t *testing.T) {
			result := analyzer.AnalyzeProgrammingTerm(tc.term)
			
			if result.IsCamelCase != tc.expectedCamel {
				t.Errorf("For term '%s': expected IsCamelCase=%v, got %v", 
					tc.term, tc.expectedCamel, result.IsCamelCase)
			}
		})
	}
}

// TestProgrammingTermAnalyzer_NilMorphEngineHandling tests that nil morphEngine is handled gracefully
func TestProgrammingTermAnalyzer_NilMorphEngineHandling(t *testing.T) {
	// Create analyzer with nil morphEngine
	analyzer := NewProgrammingTermAnalyzer(nil)
	
	testCases := []string{
		"getUserById",
		"HTTPServer", 
		"validate",
		"naïveBayes",
	}
	
	for _, term := range testCases {
		t.Run(term, func(t *testing.T) {
			// This should not panic even with nil morphEngine
			result := analyzer.AnalyzeProgrammingTerm(term)
			
			// Should still have basic analysis
			if result == nil {
				t.Errorf("Expected non-nil result for term '%s' even with nil morphEngine", term)
				return
			}
			
			// Morphological info should be empty when morphEngine is nil
			if len(result.MorphologicalInfo) != 0 {
				t.Errorf("Expected empty MorphologicalInfo when morphEngine is nil, got %v", 
					result.MorphologicalInfo)
			}
		})
	}
}