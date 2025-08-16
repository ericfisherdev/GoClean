// Package violations provides detectors for various clean code violations in Go source code.
package violations

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
	"strings"

	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/types"
)

const (
	// Float parsing precision
	floatParsePrecision = 64
	
	// Numeric calculation constants
	decimalBase = 10
	
	// Acceptable numeric ranges
	smallNumberLimit = 10
	piLowerBound = 3.1
	piUpperBound = 3.15
	eLowerBound = 2.7
	eUpperBound = 2.72
)

// MagicNumberDetector detects hardcoded magic numbers in code
type MagicNumberDetector struct {
	config        *DetectorConfig
	codeExtractor *CodeExtractor
	whitelist     *MagicNumberWhitelist
}

// NewMagicNumberDetector creates a new magic number detector
func NewMagicNumberDetector(config *DetectorConfig) *MagicNumberDetector {
	return &MagicNumberDetector{
		config:        config,
		codeExtractor: NewCodeExtractor(),
		whitelist:     DefaultMagicNumberWhitelist(),
	}
}

// Name returns the name of this detector
func (d *MagicNumberDetector) Name() string {
	return "Magic Number Detector"
}

// Description returns a description of what this detector checks for
func (d *MagicNumberDetector) Description() string {
	return "Detects hardcoded numeric literals that should be constants"
}

// Detect analyzes the provided file information and returns violations
func (d *MagicNumberDetector) Detect(fileInfo *models.FileInfo, astInfo interface{}) []*models.Violation {
	var violations []*models.Violation
	
	if astInfo == nil {
		return violations
	}
	
	// Type assertion to get scanner.GoASTInfo
	goAstInfo, ok := astInfo.(*types.GoASTInfo)
	if !ok || goAstInfo == nil || goAstInfo.AST == nil {
		return violations
	}
	
	// Walk the AST to find magic numbers with context awareness
	// Use a custom visitor to track parent and grandparent nodes
	d.walkWithFullContext(goAstInfo.AST, nil, nil, func(n ast.Node, parent ast.Node, grandparent ast.Node) bool {
		switch x := n.(type) {
		case *ast.BasicLit:
			if x.Kind == token.INT || x.Kind == token.FLOAT {
				// Check if this is a magic number with context
				if violation := d.checkMagicNumberWithFullContext(x, parent, grandparent, goAstInfo.FileSet, fileInfo.Path); violation != nil {
					violations = append(violations, violation)
				}
			}
		}
		return true
	})
	
	return violations
}

// checkMagicNumberWithFullContext checks if a literal is a magic number considering its full context
func (d *MagicNumberDetector) checkMagicNumberWithFullContext(lit *ast.BasicLit, parent ast.Node, grandparent ast.Node, fset *token.FileSet, filePath string) *models.Violation {
	value := lit.Value
	
	// Self-reference protection: skip magic number detection in the detector's own whitelist functions
	if d.isInDetectorCode(filePath, parent, grandparent) {
		return nil
	}
	
	// Determine context for whitelist checking
	context := d.determineContext(lit, parent, grandparent, filePath)
	
	// Check whitelist first
	if whitelisted, reason := d.whitelist.IsWhitelistedMagicNumber(value, context); whitelisted {
		if d.config.Verbose {
			fmt.Printf("Whitelisted magic number %s: %s\n", value, reason)
		}
		return nil
	}
	
	// Enhanced context checking
	if d.isInAcceptableContext(lit, parent, grandparent, fset) {
		return nil
	}
	
	return d.createMagicNumberViolation(lit, fset, filePath, context)
}

// checkMagicNumber checks if a literal is a magic number
func (d *MagicNumberDetector) checkMagicNumber(lit *ast.BasicLit, fset *token.FileSet, filePath string) *models.Violation {
	// Parse the value
	value := lit.Value
	
	// Check if it's an integer
	if lit.Kind == token.INT {
		intVal, err := strconv.Atoi(value)
		if err != nil {
			return nil
		}
		
		// Ignore common acceptable values
		if d.isAcceptableInt(intVal) {
			return nil
		}
	} else if lit.Kind == token.FLOAT {
		floatVal, err := strconv.ParseFloat(value, floatParsePrecision)
		if err != nil {
			return nil
		}
		
		// Ignore common acceptable float values
		if d.isAcceptableFloat(floatVal) {
			return nil
		}
	}
	
	// Get position information
	pos := fset.Position(lit.Pos())
	
	codeSnippet := d.extractCodeSnippet(filePath, pos.Line, pos.Line)
	
	return &models.Violation{
		Type:        models.ViolationTypeMagicNumber,
		Severity:    models.SeverityLow,
		File:        filePath,
		Line:        pos.Line,
		Column:      pos.Column,
		Message:     fmt.Sprintf("Magic number '%s' detected", value),
		Suggestion:  "Consider extracting this value to a named constant for better readability and maintainability",
		CodeSnippet: codeSnippet,
	}
}

// isAcceptableInt checks if an integer value is commonly acceptable
func (d *MagicNumberDetector) isAcceptableInt(value int) bool {
	// Common acceptable values that don't need to be constants
	acceptable := []int{
		-1, 0, 1, 2, // Very common values
		10, 100, 1000, 10000, 100000, 1000000, // Powers of 10
		60, 24, 7, 365, // Time-related
		1024, 2048, 4096, 8192, // Powers of 2
	}
	
	for _, v := range acceptable {
		if v == value {
			return true
		}
	}
	
	// Check if it's a power of 10
	if d.isPowerOfTen(value) {
		return true
	}
	
	// Check if it's a small loop counter or array index
	if value >= 0 && value <= smallNumberLimit {
		return true
	}
	
	return false
}

// isPowerOfTen checks if a number is a power of 10
func (d *MagicNumberDetector) isPowerOfTen(value int) bool {
	if value <= 0 {
		return false
	}
	
	temp := value
	for temp > 1 {
		if temp%decimalBase != 0 {
			return false
		}
		temp = temp / decimalBase
	}
	return temp == 1
}

// isAcceptableFloat checks if a float value is commonly acceptable
func (d *MagicNumberDetector) isAcceptableFloat(value float64) bool {
	// Common acceptable float values
	acceptable := []float64{
		0.0, 1.0, 2.0, // Common values
		0.5, 0.25, 0.75, // Common fractions
		3.14, 3.14159, 3.141592653589793, // Pi approximations
		2.71828, 2.718281828459045, // e approximations
	}
	
	for _, v := range acceptable {
		if v == value {
			return true
		}
	}
	
	// Check for close approximations of Pi
	if value > piLowerBound && value < piUpperBound {
		return true
	}
	
	// Check for close approximations of e
	if value > eLowerBound && value < eUpperBound {
		return true
	}
	
	return false
}

// walkWithFullContext walks the AST while tracking parent and grandparent nodes
func (d *MagicNumberDetector) walkWithFullContext(node ast.Node, parent ast.Node, grandparent ast.Node, fn func(ast.Node, ast.Node, ast.Node) bool) {
	if node == nil {
		return
	}
	
	if !fn(node, parent, grandparent) {
		return
	}
	
	// Walk children with current node as parent and previous parent as grandparent
	switch n := node.(type) {
	case *ast.File:
		for _, decl := range n.Decls {
			d.walkWithFullContext(decl, node, parent, fn)
		}
	case *ast.GenDecl:
		for _, spec := range n.Specs {
			d.walkWithFullContext(spec, node, parent, fn)
		}
	case *ast.ValueSpec:
		for _, value := range n.Values {
			d.walkWithFullContext(value, node, parent, fn)
		}
	case *ast.FuncDecl:
		if n.Body != nil {
			d.walkWithFullContext(n.Body, node, parent, fn)
		}
	case *ast.BlockStmt:
		for _, stmt := range n.List {
			d.walkWithFullContext(stmt, node, parent, fn)
		}
	case *ast.AssignStmt:
		for _, expr := range n.Rhs {
			d.walkWithFullContext(expr, node, parent, fn)
		}
	case *ast.ExprStmt:
		d.walkWithFullContext(n.X, node, parent, fn)
	case *ast.CallExpr:
		for _, arg := range n.Args {
			d.walkWithFullContext(arg, node, parent, fn)
		}
	case *ast.BinaryExpr:
		d.walkWithFullContext(n.X, node, parent, fn)
		d.walkWithFullContext(n.Y, node, parent, fn)
	case *ast.UnaryExpr:
		d.walkWithFullContext(n.X, node, parent, fn)
	case *ast.ParenExpr:
		d.walkWithFullContext(n.X, node, parent, fn)
	}
}

// isInAcceptableFullContext checks if a literal is in a context where it shouldn't be flagged
func (d *MagicNumberDetector) isInAcceptableFullContext(lit *ast.BasicLit, parent ast.Node, grandparent ast.Node, fset *token.FileSet) bool {
	if parent == nil {
		return false
	}
	
	switch p := parent.(type) {
	case *ast.ValueSpec:
		// Check if this is a const declaration by looking at grandparent
		return d.isConstDeclaration(p, grandparent)
	case *ast.AssignStmt:
		// Check if the variable name suggests it's not a magic number
		return d.hasDescriptiveVariableName(p, lit)
	}
	
	return false
}

// isConstDeclaration checks if a ValueSpec is part of a const declaration
func (d *MagicNumberDetector) isConstDeclaration(spec *ast.ValueSpec, grandparent ast.Node) bool {
	// Check if the grandparent is a GenDecl with token.CONST
	if genDecl, ok := grandparent.(*ast.GenDecl); ok {
		return genDecl.Tok == token.CONST
	}
	return false
}

// hasDescriptiveVariableName checks if an assignment uses descriptive variable names
func (d *MagicNumberDetector) hasDescriptiveVariableName(assign *ast.AssignStmt, lit *ast.BasicLit) bool {
	// Check if any of the left-hand side variables have descriptive names
	for _, lhs := range assign.Lhs {
		if ident, ok := lhs.(*ast.Ident); ok {
			name := ident.Name
			// Check for descriptive patterns
			if len(name) > 3 && (
				containsAny(name, []string{"timeout", "delay", "count", "size", "limit", "max", "min", "weight", "rate", "factor", "ratio", "percent"}) ||
				isDescriptiveName(name)) {
				return true
			}
		}
	}
	return false
}

// containsAny checks if a string contains any of the given substrings (case insensitive)
func containsAny(s string, substrings []string) bool {
	s = strings.ToLower(s)
	for _, substr := range substrings {
		if strings.Contains(s, strings.ToLower(substr)) {
			return true
		}
	}
	return false
}

// isDescriptiveName checks if a variable name is descriptive enough
func isDescriptiveName(name string) bool {
	// Variables with meaningful suffixes or prefixes
	descriptiveSuffixes := []string{"Weight", "Factor", "Rate", "Count", "Size", "Limit", "Max", "Min", "Timeout", "Delay"}
	descriptivePrefixes := []string{"max", "min", "default", "initial", "final"}
	
	for _, suffix := range descriptiveSuffixes {
		if strings.HasSuffix(name, suffix) {
			return true
		}
	}
	
	for _, prefix := range descriptivePrefixes {
		if strings.HasPrefix(strings.ToLower(name), prefix) {
			return true
		}
	}
	
	return false
}

// extractCodeSnippet extracts a code snippet for the violation with context
func (d *MagicNumberDetector) extractCodeSnippet(filePath string, startLine, endLine int) string {
	if d.codeExtractor == nil {
		return fmt.Sprintf("Line %d: <code snippet unavailable>", startLine)
	}
	
	snippet, err := d.codeExtractor.ExtractSnippet(filePath, startLine, endLine)
	if err != nil {
		return fmt.Sprintf("Line %d: <code snippet unavailable>", startLine)
	}
	
	return snippet
}

// determineContext analyzes the context around a magic number for better whitelist checking
func (d *MagicNumberDetector) determineContext(lit *ast.BasicLit, parent ast.Node, grandparent ast.Node, filePath string) string {
	contexts := []string{}
	
	// File context
	if strings.HasSuffix(filePath, "_test.go") {
		contexts = append(contexts, "testing")
	}
	
	// Function context
	if funcName := d.findContainingFunction(parent, grandparent); funcName != "" {
		funcNameLower := strings.ToLower(funcName)
		if strings.Contains(funcNameLower, "test") {
			contexts = append(contexts, "testing")
		}
		if strings.Contains(funcNameLower, "http") {
			contexts = append(contexts, "http")
		}
		if strings.Contains(funcNameLower, "permission") || 
		   strings.Contains(funcNameLower, "chmod") {
			contexts = append(contexts, "permissions")
		}
		// Mathematical context detection
		if d.isMathematicalContext(funcNameLower) {
			contexts = append(contexts, "mathematical")
		}
		// Benchmark context detection
		if strings.Contains(funcNameLower, "benchmark") {
			contexts = append(contexts, "benchmark")
		}
	}
	
	// Package context
	if pkg := d.getPackageName(filePath); pkg != "" {
		contexts = append(contexts, pkg)
	}
	
	return strings.Join(contexts, ",")
}

// isMathematicalContext checks if a function name suggests mathematical context
func (d *MagicNumberDetector) isMathematicalContext(funcName string) bool {
	mathKeywords := []string{
		"calculate", "compute", "math", "tolerance", "precision", "factor", "ratio",
		"percentage", "percent", "multiply", "divide", "sum", "average", "mean",
		"median", "variance", "deviation", "sqrt", "pow", "sin", "cos", "tan",
		"log", "exp", "abs", "round", "floor", "ceil", "min", "max",
	}
	
	for _, keyword := range mathKeywords {
		if strings.Contains(funcName, keyword) {
			return true
		}
	}
	return false
}

// findContainingFunction finds the name of the function containing the magic number
func (d *MagicNumberDetector) findContainingFunction(parent ast.Node, grandparent ast.Node) string {
	// Walk up the AST to find a function declaration
	current := parent
	for current != nil {
		if funcDecl, ok := current.(*ast.FuncDecl); ok {
			if funcDecl.Name != nil {
				return funcDecl.Name.Name
			}
		}
		// Move up one level (this is a simplified approach)
		current = grandparent
		grandparent = nil
	}
	return ""
}

// getPackageName extracts package name from file path
func (d *MagicNumberDetector) getPackageName(filePath string) string {
	// Extract package name from path (simplified approach)
	parts := strings.Split(filePath, "/")
	if len(parts) > 1 {
		return parts[len(parts)-2] // parent directory name
	}
	return ""
}

// isInAcceptableContext checks if a literal is in a context where it shouldn't be flagged
func (d *MagicNumberDetector) isInAcceptableContext(lit *ast.BasicLit, parent ast.Node, grandparent ast.Node, fset *token.FileSet) bool {
	if parent == nil {
		return false
	}
	
	switch p := parent.(type) {
	case *ast.ValueSpec:
		// Check if this is a const declaration by looking at grandparent
		return d.isConstDeclaration(p, grandparent)
	case *ast.AssignStmt:
		// Check if the variable name suggests it's not a magic number
		if d.hasDescriptiveVariableName(p, lit) {
			return true
		}
		
		// Check for port assignments and other descriptive contexts
		for _, lhs := range p.Lhs {
			if ident, ok := lhs.(*ast.Ident); ok {
				name := strings.ToLower(ident.Name)
				if strings.Contains(name, "port") || 
				   strings.Contains(name, "timeout") ||
				   strings.Contains(name, "interval") {
					return true
				}
			}
		}
		
	case *ast.CallExpr:
		// Check for os.OpenFile, os.Chmod, os.Mkdir calls
		if sel, ok := p.Fun.(*ast.SelectorExpr); ok {
			if ident, ok := sel.X.(*ast.Ident); ok {
				funcName := ident.Name + "." + sel.Sel.Name
				permissionFuncs := []string{
					"os.OpenFile", "os.Chmod", "os.Mkdir", "os.MkdirAll",
					"os.Create", "os.CreateTemp",
				}
				for _, pFunc := range permissionFuncs {
					if funcName == pFunc {
						return true // File permission context
					}
				}
			}
		}
		
		// Check for HTTP status codes in http package calls
		if sel, ok := p.Fun.(*ast.SelectorExpr); ok {
			if ident, ok := sel.X.(*ast.Ident); ok {
				if ident.Name == "http" || strings.Contains(sel.Sel.Name, "Status") {
					return true // HTTP status context
				}
				
				// Check for standard library functions that use bit sizes as parameters
				funcName := ident.Name + "." + sel.Sel.Name
				if d.isStandardLibraryBitSizeContext(p, lit, funcName) {
					return true // Standard library bit size context
				}
			}
		}
	}
	
	// Use the original acceptable context check as fallback
	return d.isInAcceptableFullContext(lit, parent, grandparent, fset)
}

// isInDetectorCode checks if the magic number is within the detector's own code
func (d *MagicNumberDetector) isInDetectorCode(filePath string, parent ast.Node, grandparent ast.Node) bool {
	// Skip magic number detection in the detector's own implementation file
	if strings.Contains(filePath, "magic_numbers.go") {
		return true // Skip all magic number detection in the magic number detector itself
	}
	return false
}

// isStandardLibraryBitSizeContext checks if a number is used as a bit size parameter in standard library functions
func (d *MagicNumberDetector) isStandardLibraryBitSizeContext(callExpr *ast.CallExpr, lit *ast.BasicLit, funcName string) bool {
	// Map of functions and their bit size parameter positions (0-indexed)
	bitSizeFunctions := map[string][]int{
		"strconv.ParseFloat":  {1}, // strconv.ParseFloat(s, bitSize)
		"strconv.ParseInt":    {2}, // strconv.ParseInt(s, base, bitSize)
		"strconv.ParseUint":   {2}, // strconv.ParseUint(s, base, bitSize)
		"strconv.FormatFloat": {2}, // strconv.FormatFloat(f, fmt, prec, bitSize)
		"strconv.FormatInt":   {1}, // strconv.FormatInt(i, base)
		"strconv.FormatUint":  {1}, // strconv.FormatUint(i, base)
		"math.Float32bits":    {},  // No bit size param, but handles 32-bit
		"math.Float64bits":    {},  // No bit size param, but handles 64-bit
	}

	positions, exists := bitSizeFunctions[funcName]
	if !exists {
		return false
	}

	// For functions without bit size parameters, check if the literal is a valid bit size
	if len(positions) == 0 {
		value := lit.Value
		return value == "32" || value == "64"
	}

	// Check if the literal is at one of the bit size parameter positions
	for _, pos := range positions {
		if pos < len(callExpr.Args) && callExpr.Args[pos] == lit {
			// Verify it's a valid bit size (8, 16, 32, 64)
			value := lit.Value
			validBitSizes := []string{"8", "16", "32", "64"}
			for _, validSize := range validBitSizes {
				if value == validSize {
					return true
				}
			}
		}
	}

	return false
}

// createMagicNumberViolation creates a violation for a magic number with enhanced context
func (d *MagicNumberDetector) createMagicNumberViolation(lit *ast.BasicLit, fset *token.FileSet, filePath string, context string) *models.Violation {
	// Get position information
	pos := fset.Position(lit.Pos())
	
	codeSnippet := d.extractCodeSnippet(filePath, pos.Line, pos.Line)
	
	// Enhanced message with context
	contextMsg := ""
	if context != "" {
		contextMsg = fmt.Sprintf(" (context: %s)", context)
	}
	
	return &models.Violation{
		Type:        models.ViolationTypeMagicNumber,
		Severity:    models.SeverityLow,
		File:        filePath,
		Line:        pos.Line,
		Column:      pos.Column,
		Message:     fmt.Sprintf("Magic number '%s' detected%s", lit.Value, contextMsg),
		Suggestion:  "Consider extracting this value to a named constant for better readability and maintainability",
		CodeSnippet: codeSnippet,
	}
}