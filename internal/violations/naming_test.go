package violations
import (
	"strings"
	"testing"
	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/types"
)
func TestNewNamingDetector(t *testing.T) {
	detector := NewNamingDetector(nil)
	if detector == nil {
		t.Fatal("Expected detector to be created")
	}
	if detector.config == nil {
		t.Error("Expected config to be initialized")
	}
	if detector.Name() == "" {
		t.Error("Expected detector to have a name")
	}
	if detector.Description() == "" {
		t.Error("Expected detector to have a description")
	}
}
func TestNamingDetector_NonDescriptiveFunctionName(t *testing.T) {
	detector := NewNamingDetector(nil)
	fileInfo := &models.FileInfo{Path: "test.go"}
	astInfo := &types.GoASTInfo{
		Functions: []*types.FunctionInfo{
			{
				Name:      "data", // Non-descriptive
				StartLine: 10,
				IsExported: true,
			},
		},
	}
	violations := detector.Detect(fileInfo, astInfo)
	found := false
	for _, v := range violations {
		if v.Type == models.ViolationTypeNaming && v.Rule == "non-descriptive-function-name" {
			found = true
			if !contains(v.Message, "non-descriptive") {
				t.Errorf("Expected message about non-descriptive name, got: %s", v.Message)
			}
		}
	}
	if !found {
		t.Error("Expected non-descriptive function name violation")
	}
}
func TestNamingDetector_ImproperGoCasing(t *testing.T) {
	detector := NewNamingDetector(nil)
	fileInfo := &models.FileInfo{Path: "test.go"}
	astInfo := &types.GoASTInfo{
		Functions: []*types.FunctionInfo{
			{
				Name:       "my_function", // Should be myFunction or MyFunction
				StartLine:  10,
				IsExported: false,
			},
			{
				Name:       "another_Func", // Inconsistent casing
				StartLine:  20,
				IsExported: true,
			},
		},
	}
	violations := detector.Detect(fileInfo, astInfo)
	caseSViolations := 0
	for _, v := range violations {
		if v.Type == models.ViolationTypeNaming && v.Rule == "go-function-case" {
			caseSViolations++
		}
	}
	if caseSViolations != 2 {
		t.Errorf("Expected 2 casing violations, got %d", caseSViolations)
	}
}
func TestNamingDetector_SingleLetterParameters(t *testing.T) {
	config := &DetectorConfig{
		AllowSingleLetterVars: false, // Disallow single letter vars
	}
	detector := NewNamingDetector(config)
	fileInfo := &models.FileInfo{Path: "test.go"}
	astInfo := &types.GoASTInfo{
		Functions: []*types.FunctionInfo{
			{
				Name:      "TestFunc",
				StartLine: 10,
				Parameters: []types.ParameterInfo{
					{Name: "x", Type: "int"},    // Single letter (should violate if not allowed)
					{Name: "data", Type: "string"}, // Acceptable
					{Name: "i", Type: "int"},    // Common loop counter (might be acceptable)
				},
			},
		},
	}
	violations := detector.Detect(fileInfo, astInfo)
	singleLetterViolations := 0
	for _, v := range violations {
		if v.Type == models.ViolationTypeNaming && v.Rule == "single-letter-parameter" {
			singleLetterViolations++
		}
	}
	// Should have violations for single letter params when not allowed
	if singleLetterViolations == 0 {
		t.Error("Expected single letter parameter violations")
	}
}
func TestNamingDetector_AllowSingleLetterParameters(t *testing.T) {
	config := &DetectorConfig{
		AllowSingleLetterVars: true, // Allow single letter vars
	}
	detector := NewNamingDetector(config)
	fileInfo := &models.FileInfo{Path: "test.go"}
	astInfo := &types.GoASTInfo{
		Functions: []*types.FunctionInfo{
			{
				Name:      "TestFunc",
				StartLine: 10,
				Parameters: []types.ParameterInfo{
					{Name: "x", Type: "int"},
					{Name: "y", Type: "int"},
					{Name: "i", Type: "int"},
				},
			},
		},
	}
	violations := detector.Detect(fileInfo, astInfo)
	// Should not have single letter violations when allowed
	for _, v := range violations {
		if v.Type == models.ViolationTypeNaming && v.Rule == "single-letter-parameter" {
			t.Error("Should not have single letter violations when allowed")
		}
	}
}
func TestNamingDetector_TypeNaming(t *testing.T) {
	detector := NewNamingDetector(nil)
	fileInfo := &models.FileInfo{Path: "test.go"}
	astInfo := &types.GoASTInfo{
		Types: []*types.TypeInfo{
			{
				Name:       "data",    // Non-descriptive
				StartLine:  10,
				IsExported: false,
			},
			{
				Name:       "my_type", // Wrong casing
				StartLine:  20,
				IsExported: false,
			},
		},
	}
	violations := detector.Detect(fileInfo, astInfo)
	typeViolations := 0
	for _, v := range violations {
		if v.Type == models.ViolationTypeNaming && 
		   (v.Rule == "non-descriptive-type-name" || v.Rule == "go-type-case") {
			typeViolations++
		}
	}
	if typeViolations < 2 {
		t.Errorf("Expected at least 2 type naming violations, got %d", typeViolations)
	}
}
func TestNamingDetector_VariableNaming(t *testing.T) {
	config := &DetectorConfig{
		AllowSingleLetterVars: false,
	}
	detector := NewNamingDetector(config)
	fileInfo := &models.FileInfo{Path: "test.go"}
	astInfo := &types.GoASTInfo{
		Variables: []*types.VariableInfo{
			{
				Name:       "x",        // Single letter (not allowed)
				Line:       10,
				IsExported: false,
			},
			{
				Name:       "My_Var",   // Wrong casing
				Line:       20,
				IsExported: true,
			},
		},
	}
	violations := detector.Detect(fileInfo, astInfo)
	variableViolations := 0
	for _, v := range violations {
		if v.Type == models.ViolationTypeNaming && 
		   (v.Rule == "single-letter-variable" || v.Rule == "go-variable-case") {
			variableViolations++
		}
	}
	if variableViolations < 2 {
		t.Errorf("Expected at least 2 variable naming violations, got %d", variableViolations)
	}
}
func TestNamingDetector_ConstantNaming(t *testing.T) {
	detector := NewNamingDetector(nil)
	fileInfo := &models.FileInfo{Path: "test.go"}
	astInfo := &types.GoASTInfo{
		Constants: []*types.ConstantInfo{
			{
				Name:       "my_constant", // Wrong casing
				Line:       10,
				IsExported: false,
			},
			{
				Name:       "MAX_SIZE", // ALL_CAPS acceptable for constants
				Line:       20,
				IsExported: true,
			},
			{
				Name:       "DefaultTimeout", // PascalCase acceptable for exported
				Line:       30,
				IsExported: true,
			},
		},
	}
	violations := detector.Detect(fileInfo, astInfo)
	constantViolations := 0
	for _, v := range violations {
		if v.Type == models.ViolationTypeNaming && v.Rule == "go-constant-case" {
			constantViolations++
		}
	}
	// Should only violate the first one (my_constant)
	if constantViolations != 1 {
		t.Errorf("Expected 1 constant naming violation, got %d", constantViolations)
	}
}
func TestNamingDetector_NoAST(t *testing.T) {
	detector := NewNamingDetector(nil)
	fileInfo := &models.FileInfo{Path: "test.js"} // Non-Go file
	violations := detector.Detect(fileInfo, nil)
	if len(violations) != 0 {
		t.Errorf("Expected no violations for non-Go files, got %d", len(violations))
	}
}
func TestIsNonDescriptiveName(t *testing.T) {
	detector := NewNamingDetector(nil)
	tests := []struct {
		name     string
		expected bool
	}{
		{"data", true},
		{"info", true},
		{"item", true},
		{"temp", true},
		{"tmp", true},
		{"foo", true},
		{"bar", true},
		{"x", true},
		{"a1", true},
		{"getUserData", false},
		{"processRequest", false},
		{"counter", false},
		{"userName", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.isNonDescriptiveName(tt.name)
			if result != tt.expected {
				t.Errorf("isNonDescriptiveName(%q) = %v, expected %v", tt.name, result, tt.expected)
			}
		})
	}
}
func TestIsSingleLetterVar(t *testing.T) {
	detector := NewNamingDetector(nil)
	tests := []struct {
		name     string
		expected bool
	}{
		{"x", true},
		{"i", true},
		{"A", true},
		{"xy", false},
		{"x1", false},
		{"", false},
		{"123", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.isSingleLetterVar(tt.name)
			if result != tt.expected {
				t.Errorf("isSingleLetterVar(%q) = %v, expected %v", tt.name, result, tt.expected)
			}
		})
	}
}
func TestIsCommonShortParam(t *testing.T) {
	detector := NewNamingDetector(nil)
	tests := []struct {
		name     string
		expected bool
	}{
		{"i", true},
		{"j", true},
		{"x", true},
		{"y", true},
		{"ok", true},
		{"id", true},
		{"r", true},
		{"w", true},
		{"data", false},
		{"count", false},
		{"xyz", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.isCommonShortParam(tt.name)
			if result != tt.expected {
				t.Errorf("isCommonShortParam(%q) = %v, expected %v", tt.name, result, tt.expected)
			}
		})
	}
}
func TestHasInappropriateAbbreviation(t *testing.T) {
	detector := NewNamingDetector(nil)
	tests := []struct {
		name     string
		expected bool
		reason   string
	}{
		// Original test cases
		{"processMgr", true, "mgr is a problematic abbreviation for manager"},
		{"calcTotal", true, "calc is a problematic abbreviation for calculate"},
		{"getUserStr", true, "str is a problematic abbreviation for string"},
		{"getAddr", true, "addr is a problematic abbreviation for address"},
		{"processManager", false, "manager is a complete word, not abbreviation"},
		{"calculateTotal", false, "calculate is a complete word, not abbreviation"},
		{"getUserString", false, "string is a complete word, not abbreviation"},
		{"getAddress", false, "address is a complete word, not abbreviation"},
		
		// New comprehensive test cases for req/requires scenario
		{"req", true, "req is a problematic abbreviation when standalone"},
		{"sendReq", true, "req is a problematic abbreviation at word boundary"},
		{"reqData", true, "req is a problematic abbreviation at start"},
		{"handleReqResponse", true, "req is a problematic abbreviation in compound"},
		{"requires", false, "requires is a complete word containing req"},
		{"requiresStaff", false, "requiresStaff contains complete word requires"},
		{"requestsCatering", false, "requestsCatering contains complete word requests"},
		{"requestHandler", false, "request is a complete word, not abbreviation"},
		{"required", false, "required is a complete word containing req"},
		{"requirement", false, "requirement is a complete word containing req"},
		
		// Additional test cases for other abbreviations
		{"comp", true, "comp is problematic abbreviation when standalone"},
		{"completeGeneration", false, "complete is a complete word, not abbreviation"},
		{"selectFeatures", false, "no problematic abbreviations present"},
		{"compData", true, "comp is problematic abbreviation at start"},
		{"component", false, "component is a complete word containing comp"},
		{"computation", false, "computation is a complete word containing comp"},
		{"compare", false, "compare is a complete word containing comp"},
		{"completion", false, "completion is a complete word containing comp"},
		
		// Processing related tests
		{"proc", true, "proc is problematic abbreviation when standalone"},
		{"procData", true, "proc is problematic abbreviation at start"},
		{"process", false, "process is a complete word containing proc"},
		{"processing", false, "processing is a complete word containing proc"},
		{"processor", false, "processor is a complete word containing proc"},
		{"procedure", false, "procedure is a complete word containing proc"},
		
		// Management related tests
		{"mgr", true, "mgr is problematic abbreviation when standalone"},
		{"mgrData", true, "mgr is problematic abbreviation at start"},
		{"manager", false, "manager is a complete word containing mgr"},
		{"management", false, "management is a complete word containing mgmt"},
		
		// Configuration related tests  
		{"cfg", true, "cfg is problematic abbreviation when standalone"},
		{"cfgData", true, "cfg is problematic abbreviation at start"},
		{"config", false, "config is a complete word containing cfg"},
		{"configuration", false, "configuration is a complete word containing cfg"},
		{"conf", true, "conf is problematic abbreviation when standalone"},
		{"conference", false, "conference is a complete word containing conf"},
		
		// String/structure related tests
		{"str", true, "str is problematic abbreviation when standalone"},
		{"strData", true, "str is problematic abbreviation at start"},
		{"string", false, "string is a complete word containing str"},
		{"structure", false, "structure is a complete word containing str"},
		{"stream", false, "stream is a complete word containing str"},
		
		// Response/result related tests  
		{"res", true, "res is problematic abbreviation when standalone"},
		{"resData", true, "res is problematic abbreviation at start"},
		{"response", false, "response is a complete word containing res"},
		{"result", false, "result is a complete word containing res"},
		{"resource", false, "resource is a complete word containing res"},
		{"resp", true, "resp is problematic abbreviation when standalone"},
		{"respHandler", true, "resp is problematic abbreviation at start"},
		{"responsible", false, "responsible is a complete word containing resp"},
		{"responsibility", false, "responsibility is a complete word containing resp"},
		
		// Number related tests
		{"num", true, "num is problematic abbreviation when standalone"},
		{"numItems", true, "num is problematic abbreviation at start"},
		{"number", false, "number is a complete word containing num"},
		{"numeric", false, "numeric is a complete word containing num"},
		{"numerical", false, "numerical is a complete word containing num"},
		
		// Edge cases with mixed casing
		{"RequiresAuth", false, "Requires is a complete word regardless of casing"},
		{"REQUIRES_STAFF", false, "REQUIRES is a complete word regardless of casing"},
		{"requestsData", false, "requests is a complete word containing req"},
		
		// Cases that should NOT be flagged as containing complete words
		{"something", false, "no problematic abbreviations present"},
		{"staff", false, "staff is a complete word, no problematic abbreviations"},
		{"catering", false, "catering is a complete word, no problematic abbreviations"},
		{"features", false, "features is a complete word, no problematic abbreviations"},
		{"generation", false, "generation is a complete word, no problematic abbreviations"},
		{"selector", false, "no problematic abbreviations present"},
		{"complete", false, "complete is a complete word, no problematic abbreviations"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.hasInappropriateAbbreviation(tt.name)
			if result != tt.expected {
				t.Errorf("hasInappropriateAbbreviation(%q) = %v, expected %v\nReason: %s", 
					tt.name, result, tt.expected, tt.reason)
			}
		})
	}
}
func TestIsProperGoFunctionCase(t *testing.T) {
	detector := NewNamingDetector(nil)
	tests := []struct {
		name       string
		isExported bool
		expected   bool
	}{
		// Exported functions
		{"GetUser", true, true},
		{"ProcessData", true, true},
		{"HTTPHandler", true, true},
		{"getUserData", true, false}, // Should start with uppercase
		{"get_user", true, false},    // Should not have underscores
		// Unexported functions
		{"getUserData", false, true},
		{"processData", false, true},
		{"httpHandler", false, true},
		{"GetUserData", false, false}, // Should start with lowercase
		{"get_user", false, false},    // Should not have underscores
		// Edge cases
		{"", true, false},
		{"", false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.isProperGoFunctionCase(tt.name, tt.isExported)
			if result != tt.expected {
				t.Errorf("isProperGoFunctionCase(%q, %v) = %v, expected %v", 
					tt.name, tt.isExported, result, tt.expected)
			}
		})
	}
}
func TestIsCamelCase(t *testing.T) {
	detector := NewNamingDetector(nil)
	tests := []struct {
		name     string
		expected bool
	}{
		{"getUserData", true},
		{"GetUserData", true},
		{"HTTPHandler", true}, // Acronym acceptable
		{"XMLParser", true},   // Acronym acceptable
		{"get_user_data", false}, // Underscores not allowed
		{"getUser_Data", false},  // Mixed style not allowed
		{"getUSERData", false},   // All caps in middle not allowed (unless acronym)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.isCamelCase(tt.name)
			if result != tt.expected {
				t.Errorf("isCamelCase(%q) = %v, expected %v", tt.name, result, tt.expected)
			}
		})
	}
}
func TestIsAllCapsWithUnderscores(t *testing.T) {
	detector := NewNamingDetector(nil)
	tests := []struct {
		name     string
		expected bool
	}{
		{"MAX_SIZE", true},
		{"DEFAULT_TIMEOUT", true},
		{"API_KEY", true},
		{"maxSize", false},
		{"Max_Size", false},
		{"max_size", false},
		{"MAX_SIZE_123", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.isAllCapsWithUnderscores(tt.name)
			if result != tt.expected {
				t.Errorf("isAllCapsWithUnderscores(%q) = %v, expected %v", tt.name, result, tt.expected)
			}
		})
	}
}
func TestContainsCompleteWords(t *testing.T) {
	detector := NewNamingDetector(nil)
	tests := []struct {
		name     string
		expected bool
		reason   string
	}{
		// Should be detected as complete words
		{"generation", true, "contains complete word generation"},
		{"completeGeneration", true, "contains complete word complete and generation"},
		{"selectFeatures", true, "contains complete words select and features"},
		{"requiresStaff", true, "contains complete words requires and staff"},
		{"requestsCatering", true, "contains complete words requests and catering"},
		{"processData", true, "contains complete word process"},
		{"management", true, "contains complete word management"},
		{"configuration", true, "contains complete word configuration"},
		{"structure", true, "contains complete word structure"},
		{"response", true, "contains complete word response"},
		{"request", true, "contains complete word request"},
		{"number", true, "contains complete word number"},
		{"address", true, "contains complete word address"},
		{"string", true, "contains complete word string"},
		{"button", true, "contains complete word button"},
		
		// Should NOT be detected as complete words
		{"req", false, "standalone abbreviation, no complete words"},
		{"mgr", false, "standalone abbreviation, no complete words"},
		{"calc", false, "standalone abbreviation, no complete words"},
		{"cfg", false, "standalone abbreviation, no complete words"},
		{"proc", false, "standalone abbreviation, no complete words"},
		{"reqData", false, "no complete words, just abbreviation + Data"},
		{"mgrHandler", false, "no complete words, just abbreviation + Handler"},
		{"calcTotal", false, "no complete words, just abbreviation + Total"},
		{"something", false, "no recognized complete words"},
		{"randomName", false, "no recognized complete words"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.containsCompleteWords(tt.name)
			if result != tt.expected {
				t.Errorf("containsCompleteWords(%q) = %v, expected %v\nReason: %s", 
					tt.name, result, tt.expected, tt.reason)
			}
		})
	}
}

func TestHasProblematicAbbreviation(t *testing.T) {
	detector := NewNamingDetector(nil)
	
	// Create test abbreviation info for 'req'
	reqInfo := AbbreviationInfo{
		fullWords: []string{"request", "requests", "require", "requires", "required", "requirement"},
		minLength: 3,
	}
	
	tests := []struct {
		name     string
		abbrev   string
		info     AbbreviationInfo
		expected bool
		reason   string
	}{
		// Should detect problematic abbreviations
		{"req", "req", reqInfo, true, "standalone req abbreviation"},
		{"reqdata", "req", reqInfo, true, "req at start of compound word"},
		{"sendreq", "req", reqInfo, true, "req at end of compound word"},
		{"handlereqresponse", "req", reqInfo, true, "req in middle of compound word"},
		
		// Should NOT detect when part of complete words
		{"requires", "req", reqInfo, false, "req is part of complete word requires"},
		{"requiresstaff", "req", reqInfo, false, "req is part of complete word requires"},
		{"request", "req", reqInfo, false, "req is part of complete word request"},
		{"requestsdata", "req", reqInfo, false, "req is part of complete word requests"},
		{"required", "req", reqInfo, false, "req is part of complete word required"},
		{"requirement", "req", reqInfo, false, "req is part of complete word requirement"},
		
		// Edge cases
		{"requesthandler", "req", reqInfo, false, "req is part of complete word request"},
		{"requirements", "req", reqInfo, false, "req is part of complete word requirement"},
		{"requestscatering", "req", reqInfo, false, "req is part of complete word requests"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.hasProblematicAbbreviation(strings.ToLower(tt.name), tt.abbrev, tt.info)
			if result != tt.expected {
				t.Errorf("hasProblematicAbbreviation(%q, %q) = %v, expected %v\nReason: %s", 
					tt.name, tt.abbrev, result, tt.expected, tt.reason)
			}
		})
	}
}

func TestIsAbbrevPartOfCompleteWord(t *testing.T) {
	detector := NewNamingDetector(nil)
	
	reqInfo := AbbreviationInfo{
		fullWords: []string{"request", "requests", "require", "requires", "required", "requirement"},
		minLength: 3,
	}
	
	tests := []struct {
		name     string
		abbrev   string
		info     AbbreviationInfo
		expected bool
		reason   string
	}{
		// Should be detected as part of complete words
		{"requires", "req", reqInfo, true, "req is part of requires"},
		{"request", "req", reqInfo, true, "req is part of request"},
		{"required", "req", reqInfo, true, "req is part of required"},
		{"requirement", "req", reqInfo, true, "req is part of requirement"},
		
		// Should NOT be detected as part of complete words
		{"req", "req", reqInfo, false, "standalone req is not part of larger word"},
		{"reqdata", "req", reqInfo, false, "req + data is not a recognized complete word"},
		{"sendreq", "req", reqInfo, false, "send + req is not a recognized complete word"},
		{"something", "req", reqInfo, false, "req is not present in this word"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.isAbbrevPartOfCompleteWord(strings.ToLower(tt.name), tt.abbrev, tt.info)
			if result != tt.expected {
				t.Errorf("isAbbrevPartOfCompleteWord(%q, %q) = %v, expected %v\nReason: %s", 
					tt.name, tt.abbrev, result, tt.expected, tt.reason)
			}
		})
	}
}

func TestExtractWordsFromName(t *testing.T) {
	detector := NewNamingDetector(nil)
	tests := []struct {
		name     string
		expected []string
	}{
		{"getUserData", []string{"get", "user", "data"}},
		{"refreshOptions", []string{"refresh", "options"}},
		{"requiresStaff", []string{"requires", "staff"}},
		{"selectFeatures", []string{"select", "features"}},
		{"completeGeneration", []string{"complete", "generation"}},
		{"req", []string{"req"}},
		{"reqData", []string{"req", "data"}},
		{"calculateComplexity", []string{"calculate", "complexity"}},
		{"compilePatterns", []string{"compile", "patterns"}},
		{"simple", []string{"simple"}},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.extractWordsFromName(tt.name)
			if !equalSlices(result, tt.expected) {
				t.Errorf("extractWordsFromName(%q) = %v, expected %v", tt.name, result, tt.expected)
			}
		})
	}
}

func TestIsLikelyCompleteWord(t *testing.T) {
	detector := NewNamingDetector(nil)
	tests := []struct {
		word     string
		expected bool
		reason   string
	}{
		// Should be recognized as complete words
		{"generate", true, "6+ characters, good vowel distribution"},
		{"complete", true, "6+ characters, good vowel distribution"},
		{"features", true, "6+ characters, good vowel distribution"},
		{"requires", true, "6+ characters, good vowel distribution"},
		{"staff", true, "known complete word"},
		{"refresh", true, "6+ characters, good vowel distribution"},
		{"options", true, "6+ characters, good vowel distribution"},
		{"progress", true, "6+ characters, good vowel distribution"},
		{"compile", true, "6+ characters, good vowel distribution"},
		{"calculate", true, "6+ characters, good vowel distribution"},
		{"pattern", true, "6+ characters, good vowel distribution"},
		{"patterns", true, "6+ characters, good vowel distribution"},
		{"instruction", true, "suffix -tion"},
		{"validation", true, "suffix -tion"},
		{"complexity", true, "suffix -ity"},
		{"rendering", true, "suffix -ing"},
		{"advanced", true, "suffix -ed"},
		{"scaling", true, "suffix -ing"},
		{"constraints", true, "plural form"},
		{"data", true, "known complete word"},
		{"key", true, "known complete word"},
		{"age", true, "known complete word"},
		
		// Should NOT be recognized as complete words (likely abbreviations)
		{"req", false, "too short, no vowels"},
		{"mgr", false, "too short, no vowels"},
		{"calc", false, "too short, poor vowel distribution"},
		{"cfg", false, "too short, no vowels"},
		{"proc", false, "too short, poor vowel distribution"},
		{"str", false, "too short, no vowels"},
		{"res", false, "too short, poor vowel distribution"},
		{"resp", false, "too short, poor vowel distribution"},
		{"num", false, "too short, poor vowel distribution"},
		{"addr", false, "too short, poor vowel distribution"},
		{"btn", false, "too short, no vowels"},
		{"img", false, "too short, no vowels"},
	}
	
	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			result := detector.isLikelyCompleteWord(tt.word)
			if result != tt.expected {
				t.Errorf("isLikelyCompleteWord(%q) = %v, expected %v\nReason: %s", 
					tt.word, result, tt.expected, tt.reason)
			}
		})
	}
}

func TestHasGoodVowelDistribution(t *testing.T) {
	detector := NewNamingDetector(nil)
	tests := []struct {
		word     string
		expected bool
	}{
		// Good vowel distribution
		{"generate", true},
		{"complete", true},
		{"feature", true},
		{"require", true},
		{"compile", true},
		{"pattern", true},
		
		// Poor vowel distribution (likely abbreviations)
		{"mgr", false},
		{"cfg", false},
		{"str", false},
		{"btn", false},
		{"img", false},
		{"req", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			result := detector.hasGoodVowelDistribution(tt.word)
			if result != tt.expected {
				t.Errorf("hasGoodVowelDistribution(%q) = %v, expected %v", tt.word, result, tt.expected)
			}
		})
	}
}

func TestMatchesProgrammingWordPatterns(t *testing.T) {
	detector := NewNamingDetector(nil)
	tests := []struct {
		word     string
		expected bool
		reason   string
	}{
		// Should match programming patterns
		{"validation", true, "ends with -tion"},
		{"instruction", true, "ends with -tion"},
		{"generation", true, "ends with -tion"},
		{"management", true, "ends with -ment"},
		{"completion", true, "ends with -tion"},
		{"rendering", true, "ends with -ing"},
		{"scaling", true, "ends with -ing"},
		{"advanced", true, "ends with -ed"},
		{"complexity", true, "ends with -ity"},
		{"responsibility", true, "ends with -ity"},
		{"reliable", true, "ends with -able"},
		{"responsible", true, "ends with -ible"},
		{"useful", true, "ends with -ful"},
		{"quickly", true, "ends with -ly"},
		{"procedure", true, "ends with -ure"},
		{"storage", true, "ends with -age"},
		{"initialize", true, "ends with -ize"},
		{"recreate", true, "starts with re-"},
		{"preprocess", true, "starts with pre-"},
		{"subprocess", true, "starts with sub-"},
		
		// Should NOT match programming patterns
		{"req", false, "too short, no patterns"},
		{"mgr", false, "no programming patterns"},
		{"calc", false, "no programming patterns"},
		{"cfg", false, "no programming patterns"},
		{"data", false, "no programming patterns (but still complete word)"},
		{"user", false, "no programming patterns (but still complete word)"},
	}
	
	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			result := detector.matchesProgrammingWordPatterns(tt.word)
			if result != tt.expected {
				t.Errorf("matchesProgrammingWordPatterns(%q) = %v, expected %v\nReason: %s", 
					tt.word, result, tt.expected, tt.reason)
			}
		})
	}
}

func TestProgrammaticContainsCompleteWords(t *testing.T) {
	detector := NewNamingDetector(nil)
	tests := []struct {
		name     string
		expected bool
		reason   string
	}{
		// Should be detected as containing complete words
		{"refreshOptions", true, "contains 'refresh' and 'options'"},
		{"selectFeatures", true, "contains 'select' and 'features'"},
		{"requiresStaff", true, "contains 'requires' and 'staff'"},
		{"completeGeneration", true, "contains 'complete' and 'generation'"},
		{"calculateComplexity", true, "contains 'calculate' and 'complexity'"},
		{"compilePatterns", true, "contains 'compile' and 'patterns'"},
		{"renderProgress", true, "contains 'render' and 'progress'"},
		{"validateConstraints", true, "contains 'validate' and 'constraints'"},
		{"updateConfiguration", true, "contains 'update' and 'configuration'"},
		
		// Should NOT be detected as containing complete words (because they contain known abbreviations)
		{"reqData", false, "req is known abbreviation, should reject whole name"},
		{"mgrHandler", false, "mgr is known abbreviation, should reject whole name"},
		{"calcTotal", false, "calc is known abbreviation, should reject whole name"},
		{"cfgParser", false, "cfg is known abbreviation, should reject whole name"},
		{"strBuilder", false, "str is known abbreviation, should reject whole name"},
		{"req", false, "standalone abbreviation"},
		{"mgr", false, "standalone abbreviation"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.containsCompleteWords(tt.name)
			if result != tt.expected {
				t.Errorf("containsCompleteWords(%q) = %v, expected %v\nReason: %s", 
					tt.name, result, tt.expected, tt.reason)
			}
		})
	}
}

// Helper function for comparing string slices
func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestGenerateCodeSnippets(t *testing.T) {
	detector := NewNamingDetector(nil)
	// Test function snippet generation
	fn := &types.FunctionInfo{
		Name:         "TestFunc",
		IsMethod:     false,
		ReceiverType: "",
	}
	snippet := detector.generateFunctionNameSnippet(fn)
	expected := "func TestFunc"
	if snippet != expected {
		t.Errorf("generateFunctionNameSnippet() = %q, expected %q", snippet, expected)
	}
	// Test method snippet generation
	method := &types.FunctionInfo{
		Name:         "Method",
		IsMethod:     true,
		ReceiverType: "*Struct",
	}
	methodSnippet := detector.generateFunctionNameSnippet(method)
	expectedMethod := "func (*Struct) Method"
	if methodSnippet != expectedMethod {
		t.Errorf("generateFunctionNameSnippet() = %q, expected %q", methodSnippet, expectedMethod)
	}
}