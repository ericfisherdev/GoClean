package violations

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/types"
)

// Helper function to create proper AST info from Go source code
func createASTInfo(t *testing.T, source string) *types.GoASTInfo {
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, "test.go", source, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	astInfo := &types.GoASTInfo{
		AST:       astFile,
		FileSet:   fset,
		Functions: make([]*types.FunctionInfo, 0),
		Variables: make([]*types.VariableInfo, 0),
		Constants: make([]*types.ConstantInfo, 0),
		Types:     make([]*types.TypeInfo, 0),
	}

	// Walk the AST and extract information manually (simplified version)
	ast.Inspect(astFile, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			if node.Name != nil {
				pos := fset.Position(node.Pos())
				end := fset.Position(node.End())
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
				astInfo.Functions = append(astInfo.Functions, funcInfo)
			}
		case *ast.GenDecl:
			for _, spec := range node.Specs {
				switch s := spec.(type) {
				case *ast.ValueSpec:
					for _, name := range s.Names {
						pos := fset.Position(name.Pos())
						if node.Tok == token.VAR {
							varType := "unknown"
							if s.Type != nil {
								if ident, ok := s.Type.(*ast.Ident); ok {
									varType = ident.Name
								}
							}
							varInfo := &types.VariableInfo{
								Name:       name.Name,
								Type:       varType,
								Line:       pos.Line,
								Column:     pos.Column,
								IsExported: ast.IsExported(name.Name),
								ASTNode:    s,
							}
							astInfo.Variables = append(astInfo.Variables, varInfo)
						} else if node.Tok == token.CONST {
							constType := "unknown"
							if s.Type != nil {
								if ident, ok := s.Type.(*ast.Ident); ok {
									constType = ident.Name
								}
							}
							constInfo := &types.ConstantInfo{
								Name:       name.Name,
								Type:       constType,
								Line:       pos.Line,
								Column:     pos.Column,
								IsExported: ast.IsExported(name.Name),
								ASTNode:    s,
							}
							astInfo.Constants = append(astInfo.Constants, constInfo)
						}
					}
				case *ast.TypeSpec:
					pos := fset.Position(s.Pos())
					end := fset.Position(s.End())
					typeInfo := &types.TypeInfo{
						Name:        s.Name.Name,
						StartLine:   pos.Line,
						EndLine:     end.Line,
						StartColumn: pos.Column,
						EndColumn:   end.Column,
						IsExported:  ast.IsExported(s.Name.Name),
						ASTNode:     s,
					}
					// Determine type kind
					switch s.Type.(type) {
					case *ast.StructType:
						typeInfo.Kind = "struct"
					case *ast.InterfaceType:
						typeInfo.Kind = "interface"
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

func TestGoStandardNamingDetector_Name(t *testing.T) {
	detector := NewGoStandardNamingDetector(nil)
	expected := "Go Standard Naming Conventions"
	if detector.Name() != expected {
		t.Errorf("Expected name '%s', got '%s'", expected, detector.Name())
	}
}

func TestGoStandardNamingDetector_Description(t *testing.T) {
	detector := NewGoStandardNamingDetector(nil)
	description := detector.Description()
	if len(description) == 0 {
		t.Error("Description should not be empty")
	}
}

func TestGoStandardNamingDetector_NonGoFiles(t *testing.T) {
	detector := NewGoStandardNamingDetector(nil)
	
	// Test with non-Go file
	fileInfo := &models.FileInfo{
		Name: "test.js",
		Path: "/path/to/test.js",
	}
	
	violations := detector.Detect(fileInfo, nil)
	if len(violations) != 0 {
		t.Errorf("Expected no violations for non-Go file, got %d", len(violations))
	}
}

func TestGoStandardNamingDetector_FunctionNaming(t *testing.T) {
	detector := NewGoStandardNamingDetector(nil)
	
	testCases := []struct {
		name          string
		source        string
		expectedCount int
		expectedRules []string
	}{
		{
			name: "function with underscores",
			source: `
package main
func bad_function_name() {
	// implementation
}`,
			expectedCount: 2, // underscore + mixedcaps violations
		},
		{
			name: "exported function with get prefix",
			source: `
package main
func GetData() string {
	return "data"
}`,
			expectedCount: 1, // inappropriate get prefix
		},
		{
			name: "correct camelCase function",
			source: `
package main
func goodFunction() {
	// implementation
}`,
			expectedCount: 0,
		},
		{
			name: "correct PascalCase exported function",
			source: `
package main
func GoodFunction() {
	// implementation
}`,
			expectedCount: 0,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			astInfo := createASTInfo(t, tc.source)
			
			fileInfo := &models.FileInfo{
				Name: "test.go",
				Path: "/test/test.go",
			}
			
			violations := detector.Detect(fileInfo, astInfo)
			
			if len(violations) != tc.expectedCount {
				t.Errorf("Expected %d violations, got %d for test '%s'", 
					tc.expectedCount, len(violations), tc.name)
				for i, v := range violations {
					t.Logf("Violation %d: Line %d, Message: %s", i+1, v.Line, v.Message)
				}
			}
			
			for _, violation := range violations {
				if violation.Rule != goStandardNamingRule {
					t.Errorf("Expected rule '%s', got '%s'", goStandardNamingRule, violation.Rule)
				}
			}
		})
	}
}

func TestGoStandardNamingDetector_VariableNaming(t *testing.T) {
	detector := NewGoStandardNamingDetector(nil)
	
	testCases := []struct {
		name          string
		source        string
		expectedCount int
	}{
		{
			name: "variable with underscores",
			source: `
package main
var bad_variable string = "test"`,
			expectedCount: 2, // underscore + mixedcaps violations
		},
		{
			name: "boolean without prefix",
			source: `
package main
var finished bool = true`,
			expectedCount: 1, // boolean naming suggestion
		},
		{
			name: "good boolean with prefix",
			source: `
package main
var isReady bool = false`,
			expectedCount: 0,
		},
		{
			name: "error without proper suffix",
			source: `
package main
import "errors"
var customError error = errors.New("test")`,
			expectedCount: 1, // error naming suggestion
		},
		{
			name: "good error naming",
			source: `
package main
import "errors"
var parseErr error = errors.New("parse error")`,
			expectedCount: 0,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			astInfo := createASTInfo(t, tc.source)
			
			fileInfo := &models.FileInfo{
				Name: "test.go",
				Path: "/test/test.go",
			}
			
			violations := detector.Detect(fileInfo, astInfo)
			
			if len(violations) != tc.expectedCount {
				t.Errorf("Expected %d violations, got %d for test '%s'", 
					tc.expectedCount, len(violations), tc.name)
				for i, v := range violations {
					t.Logf("Violation %d: Line %d, Message: %s", i+1, v.Line, v.Message)
				}
			}
		})
	}
}

func TestGoStandardNamingDetector_ConstantNaming(t *testing.T) {
	detector := NewGoStandardNamingDetector(nil)
	
	testCases := []struct {
		name          string
		source        string
		expectedCount int
	}{
		{
			name: "exported constant not in SCREAMING_SNAKE_CASE",
			source: `
package main
const BadConstant = "wrong"`,
			expectedCount: 1,
		},
		{
			name: "exported constant in correct format",
			source: `
package main
const GOOD_CONSTANT = "correct"`,
			expectedCount: 0,
		},
		{
			name: "unexported constant with underscores",
			source: `
package main
const bad_constant = "wrong"`,
			expectedCount: 1,
		},
		{
			name: "unexported constant in camelCase",
			source: `
package main
const goodConstant = "correct"`,
			expectedCount: 0,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			astInfo := createASTInfo(t, tc.source)
			
			fileInfo := &models.FileInfo{
				Name: "test.go",
				Path: "/test/test.go",
			}
			
			violations := detector.Detect(fileInfo, astInfo)
			
			if len(violations) != tc.expectedCount {
				t.Errorf("Expected %d violations, got %d for test '%s'", 
					tc.expectedCount, len(violations), tc.name)
				for i, v := range violations {
					t.Logf("Violation %d: Line %d, Message: %s", i+1, v.Line, v.Message)
				}
			}
		})
	}
}

func TestGoStandardNamingDetector_MixedCapsValidation(t *testing.T) {
	detector := NewGoStandardNamingDetector(nil)
	
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{"camelCase", "camelCase", true},
		{"PascalCase", "PascalCase", true},
		{"snake_case", "snake_case", false},
		{"kebab-case", "kebab-case", false},
		{"SCREAMING_SNAKE", "SCREAMING_SNAKE", false},
		{"singleword", "singleword", true},
		{"Singleword", "Singleword", true},
		{"number123", "number123", true},
		{"_underscore", "_underscore", false},
		{"123invalid", "123invalid", false},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := detector.isMixedCaps(tc.input)
			if result != tc.expected {
				t.Errorf("isMixedCaps('%s') = %v, expected %v", tc.input, result, tc.expected)
			}
		})
	}
}

func TestGoStandardNamingDetector_ScreamingSnakeCaseValidation(t *testing.T) {
	detector := NewGoStandardNamingDetector(nil)
	
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{"SCREAMING_SNAKE_CASE", "SCREAMING_SNAKE_CASE", true},
		{"SINGLE_WORD", "SINGLE_WORD", true},
		{"CONSTANTVALUE", "CONSTANTVALUE", true},
		{"camelCase", "camelCase", false},
		{"snake_case", "snake_case", false},
		{"Mixed_Case", "Mixed_Case", false},
		{"CONSTANT123", "CONSTANT123", true},
		{"123INVALID", "123INVALID", false},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := detector.isScreamingSnakeCase(tc.input)
			if result != tc.expected {
				t.Errorf("isScreamingSnakeCase('%s') = %v, expected %v", tc.input, result, tc.expected)
			}
		})
	}
}

func TestGoStandardNamingDetector_BooleanNaming(t *testing.T) {
	detector := NewGoStandardNamingDetector(nil)
	
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{"isReady", "isReady", true},
		{"hasPermission", "hasPermission", true},
		{"canEdit", "canEdit", true},
		{"allowAccess", "allowAccess", true},
		{"shouldUpdate", "shouldUpdate", true},
		{"ok", "ok", true},
		{"found", "found", true},
		{"ready", "ready", true}, // "ready" is in the common booleans list
		{"active", "active", true},
		{"someVariable", "someVariable", false},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := detector.hasGoodBooleanName(tc.input)
			if result != tc.expected {
				t.Errorf("hasGoodBooleanName('%s') = %v, expected %v", tc.input, result, tc.expected)
			}
		})
	}
}

func TestGoStandardNamingDetector_GetPrefixDetection(t *testing.T) {
	detector := NewGoStandardNamingDetector(nil)
	
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{"GetData", "GetData", true},
		{"getData", "getData", true},
		{"GetValue", "GetValue", true},
		{"get", "get", false}, // too short
		{"GetterMethod", "GetterMethod", true},
		{"Data", "Data", false},
		{"SetValue", "SetValue", false},
		{"CreateGet", "CreateGet", false},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := detector.hasInappropriateGetPrefix(tc.input)
			if result != tc.expected {
				t.Errorf("hasInappropriateGetPrefix('%s') = %v, expected %v", tc.input, result, tc.expected)
			}
		})
	}
}