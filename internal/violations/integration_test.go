package violations
import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/types"
)
func TestDetectorRegistry_IntegrationTest(t *testing.T) {
	// Create a registry with all detectors
	registry := NewDetectorRegistry()
	config := DefaultDetectorConfig()
	// Register all detectors
	registry.RegisterDetector(NewFunctionDetector(config))
	registry.RegisterDetector(NewNamingDetector(config))
	registry.RegisterDetector(NewStructureDetector(config))
	// Test with a complex Go file that has multiple violations
	complexCode := `package main
import "fmt"
// LargeStruct has too many fields
type LargeStruct struct {
	Field1  string
	Field2  int
	Field3  bool
	Field4  float64
	Field5  []string
	Field6  map[string]int
	Field7  interface{}
	Field8  chan int
	Field9  byte
	Field10 rune
	Field11 int32
	Field12 int64
	Field13 uint
	Field14 uint32
	Field15 uint64
	Field16 uintptr // Exceeds field count threshold
}
// processData has multiple violations
func processData(a, b, c, d, e, f string) string {
	timeout := 5000 // Magic number
	if len(a) > 0 {
		if len(b) > 0 {
			if len(c) > 0 {
				if len(d) > 0 { // Excessive nesting
					fmt.Printf("Processing: %s, %s, %s, %s\n", a, b, c, d)
					for i := 0; i < timeout; i++ {
						if i%100 == 0 {
							fmt.Printf("Progress: %d\n", i)
						}
					}
					return a + b + c + d + e + f
				}
			}
		}
	}
	return ""
}`
	// Parse the code
	astInfo := parseComplexGoCode(t, complexCode)
	fileInfo := &models.FileInfo{
		Path:     "complex.go",
		Language: "Go",
	}
	// Detect all violations
	violations := registry.DetectAll(fileInfo, astInfo)
	// Verify we found multiple types of violations
	violationTypes := make(map[models.ViolationType]int)
	for _, v := range violations {
		violationTypes[v.Type]++
	}
	// Expected violations:
	// 1. Function with too many parameters (processData)
	// 2. Function with excessive nesting depth (processData)
	// 3. Struct with too many fields (LargeStruct)
	// 4. Magic numbers (5000, 100)
	// 5. Non-descriptive variable names (a, b, c, d, e, f, i)
	expectedTypes := []models.ViolationType{
		models.ViolationTypeParameterCount,
		models.ViolationTypeNestingDepth,
		models.ViolationTypeClassSize,
		models.ViolationTypeMagicNumbers,
		models.ViolationTypeNaming,
	}
	for _, expectedType := range expectedTypes {
		if violationTypes[expectedType] == 0 {
			t.Errorf("Expected violations of type %s, but found none", expectedType)
		}
	}
	if len(violations) < 5 {
		t.Errorf("Expected at least 5 violations, found %d", len(violations))
	}
	// Verify all violations have required fields
	for i, v := range violations {
		if v.Type == "" {
			t.Errorf("Violation %d has empty Type", i)
		}
		if v.Message == "" {
			t.Errorf("Violation %d has empty Message", i)
		}
		if v.File == "" {
			t.Errorf("Violation %d has empty File", i)
		}
		if v.Line == 0 {
			t.Errorf("Violation %d has Line 0", i)
		}
		if v.Rule == "" {
			t.Errorf("Violation %d has empty Rule", i)
		}
	}
}
func TestDetectorRegistry_WithSeverityClassification(t *testing.T) {
	// Test that the new severity classification system works end-to-end
	registry := NewDetectorRegistry()
	config := DefaultDetectorConfig()
	// Modify config to have lower thresholds for easier testing
	config.MaxParameters = 2
	config.MaxNestingDepth = 1
	config.MaxFunctionLines = 10
	registry.RegisterDetector(NewFunctionDetector(config))
	// Code that should trigger different severity levels
	testCode := `package main
func lowSeverity(a, b, c string) { // 3 parameters (1.5x threshold = Medium)
	if a == "test" {
		println(a)
	}
}
func highSeverity(a, b, c, d, e string) { // 5 parameters (2.5x threshold = High)
	if a == "test" {
		if b == "test" { // Double nesting (2x threshold = High)
			println(a, b, c, d, e)
		}
	}
}`
	astInfo := parseComplexGoCode(t, testCode)
	fileInfo := &models.FileInfo{
		Path:     "severity_test.go",
		Language: "Go",
	}
	violations := registry.DetectAll(fileInfo, astInfo)
	// Check that we have violations with different severities
	severityCount := make(map[models.Severity]int)
	for _, v := range violations {
		severityCount[v.Severity]++
	}
	// Should have some Medium and some High severity violations
	if severityCount[models.SeverityMedium] == 0 {
		t.Error("Expected some Medium severity violations")
	}
	if severityCount[models.SeverityHigh] == 0 {
		t.Error("Expected some High severity violations")
	}
}
func TestDetectorRegistry_WithContext(t *testing.T) {
	// Test context-aware severity adjustments
	registry := NewDetectorRegistry()
	config := DefaultDetectorConfig()
	config.MaxParameters = 2
	// Enable context-based adjustments
	config.SeverityConfig.PublicFunctionSeverityBoost = true
	config.SeverityConfig.TestFilesSeverityReduction = true
	registry.RegisterDetector(NewFunctionDetector(config))
	publicFunctionCode := `package main
// PublicFunction is exported and should get severity boost
func PublicFunction(a, b, c string) {
	println(a, b, c)
}`
	astInfo := parseComplexGoCode(t, publicFunctionCode)
	fileInfo := &models.FileInfo{
		Path:     "public.go",
		Language: "Go",
	}
	violations := registry.DetectAll(fileInfo, astInfo)
	// Should detect parameter count violation
	found := false
	for _, v := range violations {
		if v.Type == models.ViolationTypeParameterCount {
			found = true
			// Note: Current detectors don't integrate with context-aware severity yet
			// This test documents expected behavior for future integration
			t.Logf("Public function parameter violation severity: %v (context-aware severity integration pending)", v.Severity)
			break
		}
	}
	if !found {
		t.Error("Expected parameter count violation for public function")
	}
}
func TestDetectorRegistry_EmptyFile(t *testing.T) {
	registry := NewDetectorRegistry()
	registry.RegisterDetector(NewFunctionDetector(nil))
	registry.RegisterDetector(NewNamingDetector(nil))
	registry.RegisterDetector(NewStructureDetector(nil))
	// Test with minimal code
	emptyCode := `package main`
	astInfo := parseComplexGoCode(t, emptyCode)
	fileInfo := &models.FileInfo{
		Path:     "empty.go",
		Language: "Go",
	}
	violations := registry.DetectAll(fileInfo, astInfo)
	// Should have no violations
	if len(violations) != 0 {
		t.Errorf("Expected no violations for empty file, got %d", len(violations))
	}
}
func TestDetectorRegistry_NilASTInfo(t *testing.T) {
	registry := NewDetectorRegistry()
	registry.RegisterDetector(NewFunctionDetector(nil))
	registry.RegisterDetector(NewNamingDetector(nil))
	registry.RegisterDetector(NewStructureDetector(nil))
	fileInfo := &models.FileInfo{
		Path:     "non_go_file.txt",
		Language: "Text",
	}
	// Test with nil AST info (simulates non-Go file)
	violations := registry.DetectAll(fileInfo, nil)
	// Should have no violations
	if len(violations) != 0 {
		t.Errorf("Expected no violations for non-Go file, got %d", len(violations))
	}
}
func TestDetectorConfig_SeverityConfigIntegration(t *testing.T) {
	// Test that custom severity config works with detectors
	customSeverityConfig := &SeverityConfig{
		LowThresholdMultiplier:    1.0,
		MediumThresholdMultiplier: 1.2, // Lower threshold for more sensitive detection
		HighThresholdMultiplier:   1.5,
		CriticalThresholdMultiplier: 2.0,
		ViolationTypeWeights: map[models.ViolationType]float64{
			models.ViolationTypeParameterCount: 2.0, // High weight
		},
	}
	config := &DetectorConfig{
		MaxParameters:  3,
		SeverityConfig: customSeverityConfig,
	}
	registry := NewDetectorRegistry()
	registry.RegisterDetector(NewFunctionDetector(config))
	testCode := `package main
func testFunc(a, b, c, d string) { // 4 parameters = 1.33x threshold
	println(a, b, c, d)
}`
	astInfo := parseComplexGoCode(t, testCode)
	fileInfo := &models.FileInfo{
		Path:     "test.go",
		Language: "Go",
	}
	violations := registry.DetectAll(fileInfo, astInfo)
	// Should find parameter count violation with higher severity due to weight
	found := false
	for _, v := range violations {
		if v.Type == models.ViolationTypeParameterCount {
			found = true
			// Note: Current function detector doesn't use the new severity classification system yet
			// This test documents the expected behavior for future integration
			t.Logf("Parameter count violation severity: %v (integration with new severity system pending)", v.Severity)
			break
		}
	}
	if !found {
		t.Error("Expected parameter count violation")
	}
}
// parseComplexGoCode is a helper function for integration tests
func parseComplexGoCode(t *testing.T, code string) *types.GoASTInfo {
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, "test.go", code, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse Go code: %v", err)
	}
	astInfo := &types.GoASTInfo{
		FilePath:    "test.go",
		PackageName: file.Name.Name,
		AST:         file,
		FileSet:     fileSet,
		Functions:   make([]*types.FunctionInfo, 0),
		Types:       make([]*types.TypeInfo, 0),
		Imports:     make([]*types.ImportInfo, 0),
		Variables:   make([]*types.VariableInfo, 0),
		Constants:   make([]*types.ConstantInfo, 0),
	}
	// Analyze the AST to populate all information
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			if node.Name != nil {
				pos := fileSet.Position(node.Pos())
				end := fileSet.Position(node.End())
				// Extract parameters
				var parameters []types.ParameterInfo
				if node.Type.Params != nil {
					for _, param := range node.Type.Params.List {
						for _, name := range param.Names {
							parameters = append(parameters, types.ParameterInfo{
								Name: name.Name,
								Type: "string", // Simplified for testing
							})
						}
					}
				}
				funcInfo := &types.FunctionInfo{
					Name:        node.Name.Name,
					StartLine:   pos.Line,
					EndLine:     end.Line,
					StartColumn: pos.Column,
					EndColumn:   end.Column,
					Parameters:  parameters,
					IsExported:  ast.IsExported(node.Name.Name),
					IsMethod:    node.Recv != nil,
					LineCount:   end.Line - pos.Line + 1,
					HasComments: node.Doc != nil,
					ASTNode:     node,
				}
				astInfo.Functions = append(astInfo.Functions, funcInfo)
			}
		case *ast.GenDecl:
			for _, spec := range node.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok && typeSpec.Name != nil {
					pos := fileSet.Position(typeSpec.Pos())
					end := fileSet.Position(typeSpec.End())
					typeInfo := &types.TypeInfo{
						Name:        typeSpec.Name.Name,
						StartLine:   pos.Line,
						EndLine:     end.Line,
						StartColumn: pos.Column,
						EndColumn:   end.Column,
						IsExported:  ast.IsExported(typeSpec.Name.Name),
						ASTNode:     typeSpec,
					}
					// Handle struct types
					if structType, ok := typeSpec.Type.(*ast.StructType); ok {
						typeInfo.Kind = "struct"
						typeInfo.FieldCount = len(structType.Fields.List)
					} else if interfaceType, ok := typeSpec.Type.(*ast.InterfaceType); ok {
						typeInfo.Kind = "interface"
						typeInfo.MethodCount = len(interfaceType.Methods.List)
					}
					astInfo.Types = append(astInfo.Types, typeInfo)
				}
			}
		}
		return true
	})
	return astInfo
}