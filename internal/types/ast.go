// Package types defines Go AST-related types and interfaces for code analysis.
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

// RustASTInfo contains comprehensive AST information for a Rust file
type RustASTInfo struct {
	FilePath    string
	CrateName   string
	Functions   []*RustFunctionInfo
	Structs     []*RustStructInfo
	Enums       []*RustEnumInfo
	Traits      []*RustTraitInfo
	Impls       []*RustImplInfo
	Modules     []*RustModuleInfo
	Constants   []*RustConstantInfo
	Uses        []*RustUseInfo
	Macros      []*RustMacroInfo
}

// RustFunctionInfo contains detailed information about a Rust function
type RustFunctionInfo struct {
	Name         string
	StartLine    int
	EndLine      int
	StartColumn  int
	EndColumn    int
	Parameters   []RustParameterInfo
	ReturnType   string
	IsPublic     bool
	IsAsync      bool
	IsUnsafe     bool
	IsConst      bool
	Complexity   int
	LineCount    int
	HasDocComments bool
	Visibility   string // "pub", "pub(crate)", "pub(super)", "private"
}

// RustParameterInfo contains information about Rust function parameters
type RustParameterInfo struct {
	Name     string
	Type     string
	IsMutable bool
	IsRef    bool
}

// RustStructInfo contains information about Rust struct declarations
type RustStructInfo struct {
	Name        string
	StartLine   int
	EndLine     int
	StartColumn int
	EndColumn   int
	IsPublic    bool
	FieldCount  int
	Visibility  string
	HasDocComments bool
}

// RustEnumInfo contains information about Rust enum declarations
type RustEnumInfo struct {
	Name        string
	StartLine   int
	EndLine     int
	StartColumn int
	EndColumn   int
	IsPublic    bool
	VariantCount int
	Visibility  string
	HasDocComments bool
}

// RustTraitInfo contains information about Rust trait declarations
type RustTraitInfo struct {
	Name        string
	StartLine   int
	EndLine     int
	StartColumn int
	EndColumn   int
	IsPublic    bool
	MethodCount int
	Visibility  string
	HasDocComments bool
}

// RustImplInfo contains information about Rust impl blocks
type RustImplInfo struct {
	StartLine   int
	EndLine     int
	StartColumn int
	EndColumn   int
	TargetType  string
	TraitName   string // Empty if inherent impl
	MethodCount int
}

// RustModuleInfo contains information about Rust module declarations
type RustModuleInfo struct {
	Name        string
	StartLine   int
	EndLine     int
	StartColumn int
	EndColumn   int
	IsPublic    bool
	Visibility  string
	HasDocComments bool
}

// RustConstantInfo contains information about Rust constant declarations
type RustConstantInfo struct {
	Name        string
	Type        string
	Line        int
	Column      int
	IsPublic    bool
	Visibility  string
	HasDocComments bool
}

// RustUseInfo contains information about Rust use declarations
type RustUseInfo struct {
	Path       string
	Alias      string
	Line       int
	Column     int
	Visibility string // For re-exports
}

// RustMacroInfo contains information about Rust macro declarations
type RustMacroInfo struct {
	Name        string
	StartLine   int
	EndLine     int
	StartColumn int
	EndColumn   int
	IsPublic    bool
	MacroType   string // "macro_rules!", "proc_macro", "derive", "attribute"
	HasDocComments bool
}