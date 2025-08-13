package scanner

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
)

// ASTAnalyzer handles Go AST parsing and analysis
type ASTAnalyzer struct {
	fileSet *token.FileSet
	verbose bool
}

// NewASTAnalyzer creates a new AST analyzer instance
func NewASTAnalyzer(verbose bool) *ASTAnalyzer {
	return &ASTAnalyzer{
		fileSet: token.NewFileSet(),
		verbose: verbose,
	}
}

// AnalyzeGoFile performs AST-based analysis of a Go source file
func (a *ASTAnalyzer) AnalyzeGoFile(filePath string) (*GoASTInfo, error) {
	if a.verbose {
		fmt.Printf("Analyzing Go file with AST: %s\n", filePath)
	}

	// Parse the Go file
	src, err := parser.ParseFile(a.fileSet, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Go file %s: %w", filePath, err)
	}

	// Create AST info
	astInfo := &GoASTInfo{
		FilePath:    filePath,
		PackageName: src.Name.Name,
		AST:         src,
		FileSet:     a.fileSet,
		Functions:   make([]*FunctionInfo, 0),
		Types:       make([]*TypeInfo, 0),
		Imports:     make([]*ImportInfo, 0),
		Variables:   make([]*VariableInfo, 0),
		Constants:   make([]*ConstantInfo, 0),
	}

	// Analyze imports
	a.analyzeImports(src, astInfo)

	// Walk the AST and collect information
	ast.Inspect(src, func(n ast.Node) bool {
		a.inspectNode(n, astInfo)
		return true
	})

	if a.verbose {
		fmt.Printf("AST analysis complete for %s: %d functions, %d types, %d imports\n",
			filepath.Base(filePath), len(astInfo.Functions), len(astInfo.Types), len(astInfo.Imports))
	}

	return astInfo, nil
}

// inspectNode analyzes individual AST nodes
func (a *ASTAnalyzer) inspectNode(n ast.Node, astInfo *GoASTInfo) {
	switch node := n.(type) {
	case *ast.FuncDecl:
		a.analyzeFunctionDecl(node, astInfo)
	case *ast.GenDecl:
		a.analyzeGenDecl(node, astInfo)
	}
}

// analyzeFunctionDecl analyzes function declarations
func (a *ASTAnalyzer) analyzeFunctionDecl(fn *ast.FuncDecl, astInfo *GoASTInfo) {
	if fn.Name == nil {
		return
	}

	pos := a.fileSet.Position(fn.Pos())
	end := a.fileSet.Position(fn.End())

	funcInfo := &FunctionInfo{
		Name:         fn.Name.Name,
		StartLine:    pos.Line,
		EndLine:      end.Line,
		StartColumn:  pos.Column,
		EndColumn:    end.Column,
		Parameters:   a.extractParameters(fn.Type.Params),
		Results:      a.extractResults(fn.Type.Results),
		IsExported:   ast.IsExported(fn.Name.Name),
		IsMethod:     fn.Recv != nil,
		Complexity:   a.calculateCyclomaticComplexity(fn),
		LineCount:    end.Line - pos.Line + 1,
		HasComments:  fn.Doc != nil && len(fn.Doc.List) > 0,
		ASTNode:      fn,
	}

	// Extract receiver information for methods
	if fn.Recv != nil && len(fn.Recv.List) > 0 {
		if recv := fn.Recv.List[0]; recv.Type != nil {
			funcInfo.ReceiverType = a.extractTypeName(recv.Type)
		}
	}

	astInfo.Functions = append(astInfo.Functions, funcInfo)
}

// analyzeGenDecl analyzes general declarations (types, vars, consts)
func (a *ASTAnalyzer) analyzeGenDecl(decl *ast.GenDecl, astInfo *GoASTInfo) {
	for _, spec := range decl.Specs {
		switch s := spec.(type) {
		case *ast.TypeSpec:
			a.analyzeTypeSpec(s, astInfo)
		case *ast.ValueSpec:
			if decl.Tok == token.VAR {
				a.analyzeVarSpec(s, astInfo)
			} else if decl.Tok == token.CONST {
				a.analyzeConstSpec(s, astInfo)
			}
		}
	}
}

// analyzeTypeSpec analyzes type specifications
func (a *ASTAnalyzer) analyzeTypeSpec(spec *ast.TypeSpec, astInfo *GoASTInfo) {
	pos := a.fileSet.Position(spec.Pos())
	end := a.fileSet.Position(spec.End())

	typeInfo := &TypeInfo{
		Name:        spec.Name.Name,
		StartLine:   pos.Line,
		EndLine:     end.Line,
		StartColumn: pos.Column,
		EndColumn:   end.Column,
		IsExported:  ast.IsExported(spec.Name.Name),
		ASTNode:     spec,
	}

	// Determine type category
	switch t := spec.Type.(type) {
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

// analyzeVarSpec analyzes variable specifications
func (a *ASTAnalyzer) analyzeVarSpec(spec *ast.ValueSpec, astInfo *GoASTInfo) {
	pos := a.fileSet.Position(spec.Pos())

	for _, name := range spec.Names {
		varInfo := &VariableInfo{
			Name:       name.Name,
			Line:       pos.Line,
			Column:     pos.Column,
			IsExported: ast.IsExported(name.Name),
			ASTNode:    spec,
		}

		if spec.Type != nil {
			varInfo.Type = a.extractTypeName(spec.Type)
		}

		astInfo.Variables = append(astInfo.Variables, varInfo)
	}
}

// analyzeConstSpec analyzes constant specifications
func (a *ASTAnalyzer) analyzeConstSpec(spec *ast.ValueSpec, astInfo *GoASTInfo) {
	pos := a.fileSet.Position(spec.Pos())

	for _, name := range spec.Names {
		constInfo := &ConstantInfo{
			Name:       name.Name,
			Line:       pos.Line,
			Column:     pos.Column,
			IsExported: ast.IsExported(name.Name),
			ASTNode:    spec,
		}

		if spec.Type != nil {
			constInfo.Type = a.extractTypeName(spec.Type)
		}

		astInfo.Constants = append(astInfo.Constants, constInfo)
	}
}

// analyzeImports extracts import information
func (a *ASTAnalyzer) analyzeImports(file *ast.File, astInfo *GoASTInfo) {
	for _, imp := range file.Imports {
		pos := a.fileSet.Position(imp.Pos())
		
		importInfo := &ImportInfo{
			Path:   strings.Trim(imp.Path.Value, `"`),
			Line:   pos.Line,
			Column: pos.Column,
		}

		if imp.Name != nil {
			importInfo.Alias = imp.Name.Name
		}

		astInfo.Imports = append(astInfo.Imports, importInfo)
	}
}

// extractParameters extracts parameter information from function type
func (a *ASTAnalyzer) extractParameters(params *ast.FieldList) []ParameterInfo {
	if params == nil {
		return nil
	}

	var parameters []ParameterInfo
	for _, param := range params.List {
		typeName := a.extractTypeName(param.Type)
		
		if len(param.Names) > 0 {
			// Named parameters
			for _, name := range param.Names {
				parameters = append(parameters, ParameterInfo{
					Name: name.Name,
					Type: typeName,
				})
			}
		} else {
			// Unnamed parameter
			parameters = append(parameters, ParameterInfo{
				Type: typeName,
			})
		}
	}

	return parameters
}

// extractResults extracts return type information
func (a *ASTAnalyzer) extractResults(results *ast.FieldList) []string {
	if results == nil {
		return nil
	}

	var returnTypes []string
	for _, result := range results.List {
		typeName := a.extractTypeName(result.Type)
		returnTypes = append(returnTypes, typeName)
	}

	return returnTypes
}

// extractTypeName extracts type name from AST expression
func (a *ASTAnalyzer) extractTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + a.extractTypeName(t.X)
	case *ast.SelectorExpr:
		return a.extractTypeName(t.X) + "." + t.Sel.Name
	case *ast.ArrayType:
		return "[]" + a.extractTypeName(t.Elt)
	case *ast.MapType:
		return "map[" + a.extractTypeName(t.Key) + "]" + a.extractTypeName(t.Value)
	case *ast.ChanType:
		prefix := "chan"
		if t.Dir == ast.RECV {
			prefix = "<-chan"
		} else if t.Dir == ast.SEND {
			prefix = "chan<-"
		}
		return prefix + " " + a.extractTypeName(t.Value)
	case *ast.FuncType:
		return "func"
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.StructType:
		return "struct{}"
	default:
		return "unknown"
	}
}

// calculateCyclomaticComplexity calculates the cyclomatic complexity of a function
func (a *ASTAnalyzer) calculateCyclomaticComplexity(fn *ast.FuncDecl) int {
	if fn.Body == nil {
		return 1 // Interface method or external function
	}

	complexity := 1 // Base complexity

	ast.Inspect(fn.Body, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt, *ast.SwitchStmt, *ast.TypeSwitchStmt:
			complexity++
		case *ast.CaseClause:
			// Each case adds to complexity (except default)
			if clause, ok := n.(*ast.CaseClause); ok && clause.List != nil {
				complexity++
			}
		}
		return true
	})

	return complexity
}

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