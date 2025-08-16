package violations
import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/types"
)
func TestNewStructureDetector(t *testing.T) {
	detector := NewStructureDetector(nil)
	if detector == nil {
		t.Error("Expected detector to be created")
	}
	if detector.config == nil {
		t.Error("Expected default config to be set")
	}
}
func TestStructureDetector_Name(t *testing.T) {
	detector := NewStructureDetector(nil)
	name := detector.Name()
	expected := "Code Structure Analysis"
	if name != expected {
		t.Errorf("Expected name %s, got %s", expected, name)
	}
}
func TestStructureDetector_Description(t *testing.T) {
	detector := NewStructureDetector(nil)
	desc := detector.Description()
	if desc == "" {
		t.Error("Expected non-empty description")
	}
}
func TestStructureDetector_Detect_WithNilASTInfo(t *testing.T) {
	detector := NewStructureDetector(nil)
	fileInfo := &models.FileInfo{
		Path:     "test.txt",
		Language: "Text",
	}
	violations := detector.Detect(fileInfo, nil)
	if len(violations) != 0 {
		t.Errorf("Expected 0 violations for non-Go file, got %d", len(violations))
	}
}
func TestStructureDetector_LargeStruct(t *testing.T) {
	// Create test configuration with low thresholds
	config := &DetectorConfig{
		MaxClassLines: 5, // Very low threshold for testing
	}
	detector := NewStructureDetector(config)
	// Create a large struct for testing
	code := `package main
type LargeStruct struct {
	Field1 string
	Field2 int
	Field3 bool
	Field4 float64
	Field5 []string
	Field6 map[string]int
	Field7 interface{}
	Field8 chan int
}`
	astInfo := parseGoCode(t, code)
	fileInfo := &models.FileInfo{
		Path:     "test.go",
		Language: "Go",
	}
	violations := detector.Detect(fileInfo, astInfo)
	// Should detect large struct violation
	found := false
	for _, v := range violations {
		if v.Type == models.ViolationTypeClassSize && v.Rule == "type-size" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find large struct violation")
	}
}
func TestStructureDetector_ManyStructFields(t *testing.T) {
	detector := NewStructureDetector(nil)
	// Create struct with many fields (exceeds the 15 field threshold)
	code := `package main
type ManyFieldsStruct struct {
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
	Field16 uintptr // This exceeds the threshold
}`
	astInfo := parseGoCode(t, code)
	fileInfo := &models.FileInfo{
		Path:     "test.go",
		Language: "Go",
	}
	violations := detector.Detect(fileInfo, astInfo)
	// Should detect struct field count violation
	found := false
	for _, v := range violations {
		if v.Type == models.ViolationTypeClassSize && v.Rule == "struct-field-count" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find struct with too many fields violation")
	}
}
func TestStructureDetector_LargeInterface(t *testing.T) {
	detector := NewStructureDetector(nil)
	// Create interface with many methods (exceeds the 5 method threshold)
	code := `package main
type LargeInterface interface {
	Method1() string
	Method2() int
	Method3() bool
	Method4() float64
	Method5() []string
	Method6() map[string]int // This exceeds the threshold
}`
	astInfo := parseGoCode(t, code)
	fileInfo := &models.FileInfo{
		Path:     "test.go",
		Language: "Go",
	}
	violations := detector.Detect(fileInfo, astInfo)
	// Should detect interface method count violation
	found := false
	for _, v := range violations {
		if v.Type == models.ViolationTypeClassSize && v.Rule == "interface-method-count" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find interface with too many methods violation")
	}
}
func TestStructureDetector_GodObject(t *testing.T) {
	// Create test configuration with low method threshold
	config := &DetectorConfig{
		MaxMethods: 3, // Low threshold for testing
	}
	detector := NewStructureDetector(config)
	// Create a type with many methods
	code := `package main
type MyStruct struct {
	field string
}
func (m MyStruct) Method1() {}
func (m MyStruct) Method2() {}
func (m MyStruct) Method3() {}
func (m MyStruct) Method4() {} // Exceeds threshold
func (m MyStruct) Method5() {}
`
	astInfo := parseGoCode(t, code)
	fileInfo := &models.FileInfo{
		Path:     "test.go",
		Language: "Go",
	}
	violations := detector.Detect(fileInfo, astInfo)
	// Should detect god object violation
	found := false
	for _, v := range violations {
		if v.Type == models.ViolationTypeClassSize && v.Rule == "god-object" {
			found = true
			if !strings.Contains(v.Message, "MyStruct") {
				t.Errorf("Expected violation message to contain 'MyStruct', got: %s", v.Message)
			}
			break
		}
	}
	if !found {
		t.Error("Expected to find god object violation")
	}
}
func TestStructureDetector_MagicNumbers(t *testing.T) {
	detector := NewStructureDetector(nil)
	// Create code with magic numbers
	code := `package main
func processData() {
	timeout := 5000 // Magic number
	maxRetries := 42 // Magic number
	if timeout > maxRetries {
		return
	}
}
`
	astInfo := parseGoCode(t, code)
	fileInfo := &models.FileInfo{
		Path:     "test.go",
		Language: "Go",
	}
	violations := detector.Detect(fileInfo, astInfo)
	// Should detect magic number violations
	magicNumberCount := 0
	for _, v := range violations {
		if v.Type == models.ViolationTypeMagicNumbers {
			magicNumberCount++
		}
	}
	if magicNumberCount < 2 {
		t.Errorf("Expected at least 2 magic number violations, got %d", magicNumberCount)
	}
}
func TestStructureDetector_CleanReceiverType(t *testing.T) {
	detector := NewStructureDetector(nil)
	tests := []struct {
		input    string
		expected string
	}{
		{"*MyStruct", "MyStruct"},
		{"MyStruct", "MyStruct"},
		{"pkg.MyStruct", "MyStruct"},
		{"*pkg.MyStruct", "MyStruct"},
	}
	for _, test := range tests {
		result := detector.cleanReceiverType(test.input)
		if result != test.expected {
			t.Errorf("cleanReceiverType(%s) = %s, expected %s", test.input, result, test.expected)
		}
	}
}
func TestStructureDetector_IsMagicNumber(t *testing.T) {
	detector := NewStructureDetector(nil)
	tests := []struct {
		value    string
		kind     token.Token
		expected bool
	}{
		{"0", token.INT, false},    // Common non-magic
		{"1", token.INT, false},    // Common non-magic
		{"2", token.INT, false},    // Common non-magic
		{"-1", token.INT, false},   // Common non-magic
		{"42", token.INT, true},    // Magic number
		{"100", token.INT, true},   // Magic number
		{"3.14", token.FLOAT, true}, // Magic number
		{"0.0", token.FLOAT, false}, // Common non-magic
		{"1.0", token.FLOAT, false}, // Common non-magic
		{"hello", token.STRING, false}, // Not a number
	}
	for _, test := range tests {
		lit := &ast.BasicLit{
			Kind:  test.kind,
			Value: test.value,
		}
		result := detector.isMagicNumber(lit)
		if result != test.expected {
			t.Errorf("isMagicNumber(%s, %s) = %v, expected %v", test.value, test.kind.String(), result, test.expected)
		}
	}
}
// Helper functions
func parseGoCode(t *testing.T, code string) *types.GoASTInfo {
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
	// Analyze the AST to populate function and type information
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			if node.Name != nil {
				pos := fileSet.Position(node.Pos())
				end := fileSet.Position(node.End())
				funcInfo := &types.FunctionInfo{
					Name:        node.Name.Name,
					StartLine:   pos.Line,
					EndLine:     end.Line,
					StartColumn: pos.Column,
					EndColumn:   end.Column,
					IsExported:  ast.IsExported(node.Name.Name),
					IsMethod:    node.Recv != nil,
					ASTNode:     node,
				}
				// Extract receiver type for methods
				if node.Recv != nil && len(node.Recv.List) > 0 {
					if recv := node.Recv.List[0]; recv.Type != nil {
						funcInfo.ReceiverType = extractTypeName(recv.Type)
					}
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
					// Determine type category and count fields/methods
					switch t := typeSpec.Type.(type) {
					case *ast.StructType:
						typeInfo.Kind = "struct"
						typeInfo.FieldCount = len(t.Fields.List)
					case *ast.InterfaceType:
						typeInfo.Kind = "interface"
						typeInfo.MethodCount = len(t.Methods.List)
					default:
						typeInfo.Kind = "alias"
					}
					astInfo.Types = append(astInfo.Types, typeInfo)
				}
			}
		}
		return true
	})
	return astInfo
}
func extractTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + extractTypeName(t.X)
	case *ast.SelectorExpr:
		return extractTypeName(t.X) + "." + t.Sel.Name
	default:
		return "unknown"
	}
}
