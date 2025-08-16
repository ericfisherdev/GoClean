// Package violations provides detectors for various clean code violations in Go source code.
package violations

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/types"
)

// DocumentationDetector checks for missing or poor quality documentation
type DocumentationDetector struct {
	config        *DetectorConfig
	codeExtractor *CodeExtractor
}

// NewDocumentationDetector creates a new documentation detector
func NewDocumentationDetector(config *DetectorConfig) *DocumentationDetector {
	return &DocumentationDetector{
		config:        config,
		codeExtractor: NewCodeExtractor(),
	}
}

// Name returns the name of this detector
func (d *DocumentationDetector) Name() string {
	return "Documentation Quality Detector"
}

// Description returns a description of what this detector checks for
func (d *DocumentationDetector) Description() string {
	return "Checks for missing or poor quality documentation in code"
}

// Detect analyzes the provided file information and returns violations
func (d *DocumentationDetector) Detect(fileInfo *models.FileInfo, astInfo interface{}) []*models.Violation {
	var violations []*models.Violation
	
	if astInfo == nil {
		return violations
	}
	
	// Type assertion to get scanner.GoASTInfo
	goAstInfo, ok := astInfo.(*types.GoASTInfo)
	if !ok || goAstInfo == nil || goAstInfo.AST == nil {
		return violations
	}
	
	// Check package documentation
	if violation := d.checkPackageDoc(goAstInfo.AST, goAstInfo.FileSet, fileInfo.Path); violation != nil {
		violations = append(violations, violation)
	}
	
	// Check exported functions
	if goAstInfo.Functions != nil {
		for _, fn := range goAstInfo.Functions {
			if fn != nil && ast.IsExported(fn.Name) {
				if violation := d.checkFunctionDoc(fn, goAstInfo.AST, goAstInfo.FileSet, fileInfo.Path); violation != nil {
					violations = append(violations, violation)
				}
			}
		}
	}
	
	// Check exported types
	ast.Inspect(goAstInfo.AST, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.GenDecl:
			if x.Tok == token.TYPE {
				for _, spec := range x.Specs {
					if typeSpec, ok := spec.(*ast.TypeSpec); ok {
						if ast.IsExported(typeSpec.Name.Name) {
							if violation := d.checkTypeDoc(x, typeSpec, goAstInfo.FileSet, fileInfo.Path); violation != nil {
								violations = append(violations, violation)
							}
						}
					}
				}
			}
		case *ast.FuncDecl:
			// Check exported methods
			if x.Recv != nil && ast.IsExported(x.Name.Name) {
				if violation := d.checkMethodDoc(x, goAstInfo.FileSet, fileInfo.Path); violation != nil {
					violations = append(violations, violation)
				}
			}
		}
		return true
	})
	
	return violations
}

// checkPackageDoc checks if the package has proper documentation
func (d *DocumentationDetector) checkPackageDoc(file *ast.File, fset *token.FileSet, filePath string) *models.Violation {
	if !d.config.RequireCommentsForPublic {
		return nil
	}
	
	// Check if package has documentation
	if file.Doc == nil || len(file.Doc.List) == 0 {
		pos := fset.Position(file.Package)
		
		return &models.Violation{
			Type:        models.ViolationTypeDocumentation,
			Severity:    models.SeverityLow,
			File:        filePath,
			Line:        pos.Line,
			Column:      pos.Column,
			Message:     fmt.Sprintf("Package '%s' is missing documentation", file.Name.Name),
			Suggestion:  fmt.Sprintf("Add a package comment starting with 'Package %s ...'", file.Name.Name),
			CodeSnippet: fmt.Sprintf("package %s", file.Name.Name),
		}
	}
	
	// Check documentation quality
	doc := file.Doc.Text()
	expectedPrefix := fmt.Sprintf("Package %s", file.Name.Name)
	if !strings.HasPrefix(doc, expectedPrefix) {
		pos := fset.Position(file.Doc.Pos())
		
		return &models.Violation{
			Type:        models.ViolationTypeDocumentation,
			Severity:    models.SeverityInfo,
			File:        filePath,
			Line:        pos.Line,
			Column:      pos.Column,
			Message:     "Package documentation should start with 'Package [name]'",
			Suggestion:  fmt.Sprintf("Update documentation to start with '%s'", expectedPrefix),
			CodeSnippet: truncateString(doc, 50),
		}
	}
	
	return nil
}

// checkFunctionDoc checks if an exported function has proper documentation
func (d *DocumentationDetector) checkFunctionDoc(fn *types.FunctionInfo, file *ast.File, fset *token.FileSet, filePath string) *models.Violation {
	if !d.config.RequireCommentsForPublic {
		return nil
	}
	
	// Find the function declaration in the AST
	var funcDecl *ast.FuncDecl
	ast.Inspect(file, func(n ast.Node) bool {
		if fd, ok := n.(*ast.FuncDecl); ok && fd.Name.Name == fn.Name {
			funcDecl = fd
			return false
		}
		return true
	})
	
	if funcDecl == nil {
		return nil
	}
	
	// Check if function has documentation
	if funcDecl.Doc == nil || len(funcDecl.Doc.List) == 0 {
		return &models.Violation{
			Type:        models.ViolationTypeDocumentation,
			Severity:    models.SeverityMedium,
			File:        filePath,
			Line:        fn.StartLine,
			Column:      0,
			Message:     fmt.Sprintf("Exported function '%s' is missing documentation", fn.Name),
			Suggestion:  fmt.Sprintf("Add a comment starting with '%s ...' before the function", fn.Name),
			CodeSnippet: fmt.Sprintf("func %s(...)", fn.Name),
		}
	}
	
	// Check documentation quality
	doc := funcDecl.Doc.Text()
	if !strings.HasPrefix(doc, fn.Name) {
		return &models.Violation{
			Type:        models.ViolationTypeDocumentation,
			Severity:    models.SeverityInfo,
			File:        filePath,
			Line:        fn.StartLine - 1,
			Column:      0,
			Message:     fmt.Sprintf("Function documentation should start with the function name '%s'", fn.Name),
			Suggestion:  "Follow Go documentation conventions: comments should start with the name being documented",
			CodeSnippet: truncateString(doc, 50),
		}
	}
	
	return nil
}

// checkTypeDoc checks if an exported type has proper documentation
func (d *DocumentationDetector) checkTypeDoc(genDecl *ast.GenDecl, typeSpec *ast.TypeSpec, fset *token.FileSet, filePath string) *models.Violation {
	if !d.config.RequireCommentsForPublic {
		return nil
	}
	
	typeName := typeSpec.Name.Name
	pos := fset.Position(typeSpec.Pos())
	
	// Check if type has documentation
	if genDecl.Doc == nil || len(genDecl.Doc.List) == 0 {
		return &models.Violation{
			Type:        models.ViolationTypeDocumentation,
			Severity:    models.SeverityMedium,
			File:        filePath,
			Line:        pos.Line,
			Column:      pos.Column,
			Message:     fmt.Sprintf("Exported type '%s' is missing documentation", typeName),
			Suggestion:  fmt.Sprintf("Add a comment starting with '%s ...' before the type definition", typeName),
			CodeSnippet: fmt.Sprintf("type %s ...", typeName),
		}
	}
	
	// Check documentation quality
	doc := genDecl.Doc.Text()
	if !strings.HasPrefix(doc, typeName) {
		docPos := fset.Position(genDecl.Doc.Pos())
		
		return &models.Violation{
			Type:        models.ViolationTypeDocumentation,
			Severity:    models.SeverityInfo,
			File:        filePath,
			Line:        docPos.Line,
			Column:      docPos.Column,
			Message:     fmt.Sprintf("Type documentation should start with the type name '%s'", typeName),
			Suggestion:  "Follow Go documentation conventions: comments should start with the name being documented",
			CodeSnippet: truncateString(doc, 50),
		}
	}
	
	return nil
}

// checkMethodDoc checks if an exported method has proper documentation
func (d *DocumentationDetector) checkMethodDoc(funcDecl *ast.FuncDecl, fset *token.FileSet, filePath string) *models.Violation {
	if !d.config.RequireCommentsForPublic {
		return nil
	}
	
	methodName := funcDecl.Name.Name
	pos := fset.Position(funcDecl.Pos())
	
	// Check if method has documentation
	if funcDecl.Doc == nil || len(funcDecl.Doc.List) == 0 {
		// Get receiver type name for better error message
		receiverType := "receiver"
		if funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 {
			if ident, ok := getReceiverType(funcDecl.Recv.List[0].Type); ok {
				receiverType = ident
			}
		}
		
		return &models.Violation{
			Type:        models.ViolationTypeDocumentation,
			Severity:    models.SeverityMedium,
			File:        filePath,
			Line:        pos.Line,
			Column:      pos.Column,
			Message:     fmt.Sprintf("Exported method '%s' on '%s' is missing documentation", methodName, receiverType),
			Suggestion:  fmt.Sprintf("Add a comment starting with '%s ...' before the method", methodName),
			CodeSnippet: fmt.Sprintf("func (...) %s(...)", methodName),
		}
	}
	
	// Check documentation quality
	doc := funcDecl.Doc.Text()
	if !strings.HasPrefix(doc, methodName) {
		docPos := fset.Position(funcDecl.Doc.Pos())
		
		return &models.Violation{
			Type:        models.ViolationTypeDocumentation,
			Severity:    models.SeverityInfo,
			File:        filePath,
			Line:        docPos.Line,
			Column:      docPos.Column,
			Message:     fmt.Sprintf("Method documentation should start with the method name '%s'", methodName),
			Suggestion:  "Follow Go documentation conventions: comments should start with the name being documented",
			CodeSnippet: truncateString(doc, 50),
		}
	}
	
	return nil
}

// getReceiverType extracts the receiver type name from an expression
func getReceiverType(expr ast.Expr) (string, bool) {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name, true
	case *ast.StarExpr:
		return getReceiverType(t.X)
	}
	return "", false
}

// truncateString truncates a string to a maximum length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}