package violations

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/scanner"
)

func TestNewFunctionDetector(t *testing.T) {
	detector := NewFunctionDetector(nil)
	
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

func TestFunctionDetector_LongFunction(t *testing.T) {
	config := &DetectorConfig{
		MaxFunctionLines: 5, // Very low threshold for testing
	}
	detector := NewFunctionDetector(config)
	
	// Create a file info
	fileInfo := &models.FileInfo{
		Path: "test.go",
	}
	
	// Create AST info with a long function
	astInfo := &scanner.GoASTInfo{
		Functions: []*scanner.FunctionInfo{
			{
				Name:      "LongFunction",
				StartLine: 10,
				LineCount: 15, // Exceeds threshold of 5
				Parameters: []scanner.ParameterInfo{
					{Name: "x", Type: "int"},
				},
				IsExported: true,
			},
		},
	}
	
	violations := detector.Detect(fileInfo, astInfo)
	
	if len(violations) == 0 {
		t.Error("Expected violations for long function")
	}
	
	found := false
	for _, v := range violations {
		if v.Type == models.ViolationTypeFunctionLength {
			found = true
			if v.Severity == models.SeverityLow {
				t.Error("Expected higher severity for very long function")
			}
			if v.Line != 10 {
				t.Errorf("Expected line 10, got %d", v.Line)
			}
			if v.Rule != "function-length" {
				t.Errorf("Expected rule 'function-length', got %s", v.Rule)
			}
		}
	}
	
	if !found {
		t.Error("Expected function length violation")
	}
}

func TestFunctionDetector_ComplexFunction(t *testing.T) {
	config := &DetectorConfig{
		MaxCyclomaticComplexity: 3,
	}
	detector := NewFunctionDetector(config)
	
	fileInfo := &models.FileInfo{
		Path: "test.go",
	}
	
	astInfo := &scanner.GoASTInfo{
		Functions: []*scanner.FunctionInfo{
			{
				Name:       "ComplexFunction",
				StartLine:  10,
				LineCount:  20,
				Complexity: 8, // Exceeds threshold of 3
				Parameters: []scanner.ParameterInfo{
					{Name: "x", Type: "int"},
				},
			},
		},
	}
	
	violations := detector.Detect(fileInfo, astInfo)
	
	found := false
	for _, v := range violations {
		if v.Type == models.ViolationTypeCyclomaticComplexity {
			found = true
			if v.Severity == models.SeverityLow {
				t.Error("Expected higher severity for very complex function")
			}
		}
	}
	
	if !found {
		t.Error("Expected cyclomatic complexity violation")
	}
}

func TestFunctionDetector_TooManyParameters(t *testing.T) {
	config := &DetectorConfig{
		MaxParameters: 2,
	}
	detector := NewFunctionDetector(config)
	
	fileInfo := &models.FileInfo{
		Path: "test.go",
	}
	
	astInfo := &scanner.GoASTInfo{
		Functions: []*scanner.FunctionInfo{
			{
				Name:      "FunctionWithManyParams",
				StartLine: 10,
				LineCount: 5,
				Parameters: []scanner.ParameterInfo{
					{Name: "a", Type: "int"},
					{Name: "b", Type: "string"},
					{Name: "c", Type: "bool"},
					{Name: "d", Type: "float64"},
				}, // 4 parameters, exceeds threshold of 2
			},
		},
	}
	
	violations := detector.Detect(fileInfo, astInfo)
	
	found := false
	for _, v := range violations {
		if v.Type == models.ViolationTypeParameterCount {
			found = true
			if !contains(v.Message, "4") {
				t.Errorf("Expected message to contain parameter count, got: %s", v.Message)
			}
		}
	}
	
	if !found {
		t.Error("Expected parameter count violation")
	}
}

func TestFunctionDetector_MissingDocumentation(t *testing.T) {
	config := &DetectorConfig{
		RequireCommentsForPublic: true,
	}
	detector := NewFunctionDetector(config)
	
	fileInfo := &models.FileInfo{
		Path: "test.go",
	}
	
	astInfo := &scanner.GoASTInfo{
		Functions: []*scanner.FunctionInfo{
			{
				Name:        "PublicFunction",
				StartLine:   10,
				LineCount:   5,
				IsExported:  true,
				HasComments: false, // Missing documentation
				Parameters: []scanner.ParameterInfo{
					{Name: "x", Type: "int"},
				},
			},
		},
	}
	
	violations := detector.Detect(fileInfo, astInfo)
	
	found := false
	for _, v := range violations {
		if v.Type == models.ViolationTypeMissingDocumentation {
			found = true
			if v.Severity != models.SeverityMedium {
				t.Errorf("Expected medium severity, got %s", v.Severity.String())
			}
		}
	}
	
	if !found {
		t.Error("Expected missing documentation violation")
	}
}

func TestFunctionDetector_PrivateFunction_NoDocumentationRequired(t *testing.T) {
	config := &DetectorConfig{
		RequireCommentsForPublic: true,
	}
	detector := NewFunctionDetector(config)
	
	fileInfo := &models.FileInfo{
		Path: "test.go",
	}
	
	astInfo := &scanner.GoASTInfo{
		Functions: []*scanner.FunctionInfo{
			{
				Name:        "privateFunction", // Not exported
				StartLine:   10,
				LineCount:   5,
				IsExported:  false,
				HasComments: false,
				Parameters: []scanner.ParameterInfo{
					{Name: "x", Type: "int"},
				},
			},
		},
	}
	
	violations := detector.Detect(fileInfo, astInfo)
	
	// Should not have missing documentation violation for private functions
	for _, v := range violations {
		if v.Type == models.ViolationTypeMissingDocumentation {
			t.Error("Should not require documentation for private functions")
		}
	}
}

func TestFunctionDetector_MultipleViolations(t *testing.T) {
	config := &DetectorConfig{
		MaxFunctionLines:        10,
		MaxCyclomaticComplexity: 5,
		MaxParameters:          2,
		RequireCommentsForPublic: true,
	}
	detector := NewFunctionDetector(config)
	
	fileInfo := &models.FileInfo{
		Path: "test.go",
	}
	
	astInfo := &scanner.GoASTInfo{
		Functions: []*scanner.FunctionInfo{
			{
				Name:       "BadFunction",
				StartLine:  10,
				LineCount:  25,  // Too long
				Complexity: 12,  // Too complex
				IsExported: true,
				HasComments: false, // Missing docs
				Parameters: []scanner.ParameterInfo{
					{Name: "a", Type: "int"},
					{Name: "b", Type: "string"},
					{Name: "c", Type: "bool"},
					{Name: "d", Type: "float64"},
				}, // Too many parameters
			},
		},
	}
	
	violations := detector.Detect(fileInfo, astInfo)
	
	// Should have multiple violations
	if len(violations) < 4 {
		t.Errorf("Expected at least 4 violations, got %d", len(violations))
	}
	
	types := make(map[models.ViolationType]bool)
	for _, v := range violations {
		types[v.Type] = true
	}
	
	expectedTypes := []models.ViolationType{
		models.ViolationTypeFunctionLength,
		models.ViolationTypeCyclomaticComplexity,
		models.ViolationTypeParameterCount,
		models.ViolationTypeMissingDocumentation,
	}
	
	for _, expected := range expectedTypes {
		if !types[expected] {
			t.Errorf("Expected violation type %s", expected)
		}
	}
}

func TestFunctionDetector_NoAST(t *testing.T) {
	detector := NewFunctionDetector(nil)
	
	fileInfo := &models.FileInfo{
		Path: "test.js", // Non-Go file
	}
	
	// No AST info for non-Go files
	violations := detector.Detect(fileInfo, nil)
	
	if len(violations) != 0 {
		t.Errorf("Expected no violations for non-Go files, got %d", len(violations))
	}
}

func TestGenerateFunctionSignature(t *testing.T) {
	detector := NewFunctionDetector(nil)
	
	tests := []struct {
		name     string
		function *scanner.FunctionInfo
		expected string
	}{
		{
			name: "simple function",
			function: &scanner.FunctionInfo{
				Name: "TestFunc",
				Parameters: []scanner.ParameterInfo{
					{Name: "x", Type: "int"},
					{Name: "y", Type: "string"},
				},
				Results: []string{"error"},
			},
			expected: "func TestFunc(x int, y string) error",
		},
		{
			name: "method with receiver",
			function: &scanner.FunctionInfo{
				Name:         "Method",
				IsMethod:     true,
				ReceiverType: "*Struct",
				Parameters: []scanner.ParameterInfo{
					{Name: "value", Type: "int"},
				},
				Results: []string{"int", "error"},
			},
			expected: "func (*Struct) Method(value int) (int, error)",
		},
		{
			name: "function with no parameters",
			function: &scanner.FunctionInfo{
				Name:       "NoParams",
				Parameters: []scanner.ParameterInfo{},
				Results:    []string{},
			},
			expected: "func NoParams()",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signature := detector.generateFunctionSignature(tt.function)
			if signature != tt.expected {
				t.Errorf("Expected signature %q, got %q", tt.expected, signature)
			}
		})
	}
}

func TestSeverityCalculation(t *testing.T) {
	config := &DetectorConfig{
		MaxFunctionLines:        10,
		MaxCyclomaticComplexity: 5,
		MaxParameters:          3,
		MaxNestingDepth:        2,
	}
	detector := NewFunctionDetector(config)
	
	// Test function length severity
	if detector.getSeverityForFunctionLength(12) != models.SeverityLow {
		t.Error("Expected low severity for slightly long function")
	}
	if detector.getSeverityForFunctionLength(18) != models.SeverityMedium {
		t.Error("Expected medium severity for moderately long function")
	}
	if detector.getSeverityForFunctionLength(25) != models.SeverityHigh {
		t.Error("Expected high severity for very long function")
	}
	
	// Test complexity severity
	if detector.getSeverityForComplexity(6) != models.SeverityLow {
		t.Error("Expected low severity for slightly complex function")
	}
	if detector.getSeverityForComplexity(9) != models.SeverityMedium {
		t.Error("Expected medium severity for moderately complex function")
	}
	if detector.getSeverityForComplexity(12) != models.SeverityHigh {
		t.Error("Expected high severity for very complex function")
	}
}

func TestFunctionDetector_NestingDepth(t *testing.T) {
	config := &DetectorConfig{
		MaxNestingDepth: 2,
	}
	detector := NewFunctionDetector(config)

	// Create code with excessive nesting
	code := `package main

func deeplyNestedFunction() {
	if true {
		if true {
			if true { // This exceeds MaxNestingDepth of 2
				println("deeply nested")
			}
		}
	}
}`

	astInfo := parseGoCodeForFunction(t, code)
	fileInfo := &models.FileInfo{
		Path:     "test.go",
		Language: "Go",
	}

	violations := detector.Detect(fileInfo, astInfo)

	// Should detect nesting depth violation
	found := false
	for _, v := range violations {
		if v.Type == models.ViolationTypeNestingDepth {
			found = true
			if !strings.Contains(v.Message, "nesting depth") {
				t.Errorf("Expected nesting depth message, got: %s", v.Message)
			}
			break
		}
	}
	if !found {
		t.Error("Expected to find nesting depth violation")
	}
}

func TestFunctionDetector_CalculateNestingDepth(t *testing.T) {
	detector := NewFunctionDetector(nil)

	tests := []struct {
		name     string
		code     string
		expected int
	}{
		{
			name: "simple function",
			code: `package main
func simple() {
	println("hello")
}`,
			expected: 0,
		},
		{
			name: "single if",
			code: `package main
func singleIf() {
	if true {
		println("nested")
	}
}`,
			expected: 1,
		},
		{
			name: "nested if statements",
			code: `package main
func nestedIfs() {
	if true {
		if true {
			println("double nested")
		}
	}
}`,
			expected: 2,
		},
		{
			name: "for loop with if",
			code: `package main
func forWithIf() {
	for i := 0; i < 10; i++ {
		if i%2 == 0 {
			println(i)
		}
	}
}`,
			expected: 2,
		},
		{
			name: "switch statement",
			code: `package main
func switchStatement() {
	switch x := 1; x {
	case 1:
		if true {
			println("nested in switch")
		}
	}
}`,
			expected: 1, // Switch increases depth by 1, case is not a separate nesting level
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			astInfo := parseGoCodeForFunction(t, test.code)
			if len(astInfo.Functions) == 0 {
				t.Fatal("No functions found in parsed code")
			}
			
			fn := astInfo.Functions[0]
			if fn.ASTNode == nil {
				t.Fatal("Function AST node is nil")
			}
			
			depth := detector.calculateNestingDepth(fn.ASTNode)
			if depth != test.expected {
				t.Errorf("Expected nesting depth %d, got %d", test.expected, depth)
			}
		})
	}
}

func TestFunctionDetector_GetNestingDepthSeverity(t *testing.T) {
	config := &DetectorConfig{
		MaxNestingDepth: 3,
	}
	detector := NewFunctionDetector(config)

	tests := []struct {
		depth    int
		expected models.Severity
	}{
		{3, models.SeverityLow},    // At threshold
		{4, models.SeverityLow},    // Just above threshold (1.33x)
		{5, models.SeverityMedium}, // 1.67x threshold  
		{6, models.SeverityMedium}, // 2x threshold (but condition uses >, so still Medium)
		{7, models.SeverityHigh},   // Above 2x threshold
		{9, models.SeverityHigh},   // Way above threshold
	}

	for _, test := range tests {
		result := detector.getSeverityForNestingDepth(test.depth)
		if result != test.expected {
			t.Errorf("getSeverityForNestingDepth(%d) = %v, expected %v", test.depth, result, test.expected)
		}
	}
}

func TestFunctionDetector_GetNestingDepthSuggestion(t *testing.T) {
	detector := NewFunctionDetector(nil)

	suggestion := detector.getNestingDepthSuggestion("testFunction", 5)
	if suggestion == "" {
		t.Error("Expected non-empty suggestion")
	}
	if !strings.Contains(suggestion, "testFunction") {
		t.Error("Expected suggestion to contain function name")
	}
	if !strings.Contains(suggestion, "5") {
		t.Error("Expected suggestion to contain depth value")
	}
}

// parseGoCodeForFunction is a helper function to parse Go code for testing
func parseGoCodeForFunction(t *testing.T, code string) *scanner.GoASTInfo {
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, "test.go", code, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse Go code: %v", err)
	}

	astInfo := &scanner.GoASTInfo{
		FilePath:    "test.go",
		PackageName: file.Name.Name,
		AST:         file,
		FileSet:     fileSet,
		Functions:   make([]*scanner.FunctionInfo, 0),
		Types:       make([]*scanner.TypeInfo, 0),
		Imports:     make([]*scanner.ImportInfo, 0),
		Variables:   make([]*scanner.VariableInfo, 0),
		Constants:   make([]*scanner.ConstantInfo, 0),
	}

	// Analyze the AST to populate function information
	ast.Inspect(file, func(n ast.Node) bool {
		if funcDecl, ok := n.(*ast.FuncDecl); ok && funcDecl.Name != nil {
			pos := fileSet.Position(funcDecl.Pos())
			end := fileSet.Position(funcDecl.End())
			
			funcInfo := &scanner.FunctionInfo{
				Name:        funcDecl.Name.Name,
				StartLine:   pos.Line,
				EndLine:     end.Line,
				StartColumn: pos.Column,
				EndColumn:   end.Column,
				IsExported:  ast.IsExported(funcDecl.Name.Name),
				IsMethod:    funcDecl.Recv != nil,
				ASTNode:     funcDecl,
			}
			
			astInfo.Functions = append(astInfo.Functions, funcInfo)
		}
		return true
	})

	return astInfo
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}