package violations

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"

	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/types"
)

// MagicNumberDetector detects hardcoded magic numbers in code
type MagicNumberDetector struct {
	config *DetectorConfig
}

// NewMagicNumberDetector creates a new magic number detector
func NewMagicNumberDetector(config *DetectorConfig) *MagicNumberDetector {
	return &MagicNumberDetector{
		config: config,
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
	
	// Walk the AST to find magic numbers
	ast.Inspect(goAstInfo.AST, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.BasicLit:
			if x.Kind == token.INT || x.Kind == token.FLOAT {
				// Check if this is a magic number
				if violation := d.checkMagicNumber(x, goAstInfo.FileSet, fileInfo.Path); violation != nil {
					violations = append(violations, violation)
				}
			}
		}
		return true
	})
	
	return violations
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
		floatVal, err := strconv.ParseFloat(value, 64)
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
	
	return &models.Violation{
		Type:        models.ViolationTypeMagicNumber,
		Severity:    models.SeverityLow,
		File:        filePath,
		Line:        pos.Line,
		Column:      pos.Column,
		Message:     fmt.Sprintf("Magic number '%s' detected", value),
		Suggestion:  "Consider extracting this value to a named constant for better readability and maintainability",
		CodeSnippet: value,
	}
}

// isAcceptableInt checks if an integer value is commonly acceptable
func (d *MagicNumberDetector) isAcceptableInt(value int) bool {
	// Common acceptable values that don't need to be constants
	acceptable := []int{
		-1, 0, 1, 2, // Very common values
		10, 100, 1000, // Powers of 10
		60, 24, 7, 365, // Time-related
		1024, 2048, 4096, // Powers of 2
	}
	
	for _, v := range acceptable {
		if v == value {
			return true
		}
	}
	
	// Check if it's a small loop counter or array index
	if value >= 0 && value <= 10 {
		return true
	}
	
	return false
}

// isAcceptableFloat checks if a float value is commonly acceptable
func (d *MagicNumberDetector) isAcceptableFloat(value float64) bool {
	// Common acceptable float values
	acceptable := []float64{
		0.0, 1.0, 2.0, // Common values
		0.5, 0.25, 0.75, // Common fractions
		3.14, 3.14159, // Pi approximations
		2.71828, // e approximation
	}
	
	for _, v := range acceptable {
		if v == value {
			return true
		}
	}
	
	return false
}