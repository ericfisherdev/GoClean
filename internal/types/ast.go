package types

import (
	"go/ast"
	"go/token"
)

// GoASTInfo contains comprehensive AST information for a Go file
type GoASTInfo struct {
	FilePath    string
	PackageName string
	AST         *ast.File
	FileSet     *token.FileSet
	Functions   []*FunctionInfo
	Types       []*TypeInfo
	Imports     []*ImportInfo
	Variables   []*VariableInfo
	Constants   []*ConstantInfo
}

// FunctionInfo contains detailed information about a function
type FunctionInfo struct {
	Name         string
	StartLine    int
	EndLine      int
	StartColumn  int
	EndColumn    int
	Parameters   []ParameterInfo
	Results      []string
	IsExported   bool
	IsMethod     bool
	ReceiverType string
	Complexity   int
	LineCount    int
	HasComments  bool
	ASTNode      *ast.FuncDecl
}

// ParameterInfo contains information about function parameters
type ParameterInfo struct {
	Name string
	Type string
}

// TypeInfo contains information about type declarations
type TypeInfo struct {
	Name        string
	Kind        string // "struct", "interface", "alias"
	StartLine   int
	EndLine     int
	StartColumn int
	EndColumn   int
	IsExported  bool
	FieldCount  int // For structs
	MethodCount int // For interfaces
	ASTNode     *ast.TypeSpec
}

// ImportInfo contains information about imports
type ImportInfo struct {
	Path   string
	Alias  string
	Line   int
	Column int
}

// VariableInfo contains information about variable declarations
type VariableInfo struct {
	Name       string
	Type       string
	Line       int
	Column     int
	IsExported bool
	ASTNode    *ast.ValueSpec
}

// ConstantInfo contains information about constant declarations
type ConstantInfo struct {
	Name       string
	Type       string
	Line       int
	Column     int
	IsExported bool
	ASTNode    *ast.ValueSpec
}