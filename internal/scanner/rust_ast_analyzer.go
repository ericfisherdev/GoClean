package scanner

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ericfisherdev/goclean/internal/types"
)

// RustASTAnalyzer handles Rust AST parsing and analysis
// Note: This is a simplified implementation that will be enhanced with syn crate integration
type RustASTAnalyzer struct {
	verbose bool
}

// NewRustASTAnalyzer creates a new Rust AST analyzer instance
func NewRustASTAnalyzer(verbose bool) *RustASTAnalyzer {
	return &RustASTAnalyzer{
		verbose: verbose,
	}
}

// AnalyzeRustFile performs AST-based analysis of a Rust source file
// TODO: This is a temporary implementation using regex parsing
// Will be replaced with proper syn crate integration via CGO
func (a *RustASTAnalyzer) AnalyzeRustFile(filePath string, content []byte) (*types.RustASTInfo, error) {
	if a.verbose {
		fmt.Printf("Analyzing Rust file with AST: %s\n", filePath)
	}

	source := string(content)
	lines := strings.Split(source, "\n")

	// Create AST info
	astInfo := &types.RustASTInfo{
		FilePath:  filePath,
		CrateName: a.extractCrateName(source),
		Functions: make([]*types.RustFunctionInfo, 0),
		Structs:   make([]*types.RustStructInfo, 0),
		Enums:     make([]*types.RustEnumInfo, 0),
		Traits:    make([]*types.RustTraitInfo, 0),
		Impls:     make([]*types.RustImplInfo, 0),
		Modules:   make([]*types.RustModuleInfo, 0),
		Constants: make([]*types.RustConstantInfo, 0),
		Uses:      make([]*types.RustUseInfo, 0),
		Macros:    make([]*types.RustMacroInfo, 0),
	}

	// Parse file line by line for basic information
	// TODO: Replace with proper syn crate parsing
	a.parseRustContent(lines, astInfo)

	if a.verbose {
		fmt.Printf("Rust AST analysis complete for %s: %d functions, %d structs, %d enums\n",
			filePath, len(astInfo.Functions), len(astInfo.Structs), len(astInfo.Enums))
	}

	return astInfo, nil
}

// parseRustContent parses Rust content using regex patterns
// This is a temporary implementation that will be replaced with syn crate integration
func (a *RustASTAnalyzer) parseRustContent(lines []string, astInfo *types.RustASTInfo) {
	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") {
			continue
		}

		// Parse function declarations
		if funcInfo := a.parseFunctionDeclaration(line, lineNum); funcInfo != nil {
			astInfo.Functions = append(astInfo.Functions, funcInfo)
		}

		// Parse struct declarations
		if structInfo := a.parseStructDeclaration(line, lineNum); structInfo != nil {
			astInfo.Structs = append(astInfo.Structs, structInfo)
		}

		// Parse enum declarations
		if enumInfo := a.parseEnumDeclaration(line, lineNum); enumInfo != nil {
			astInfo.Enums = append(astInfo.Enums, enumInfo)
		}

		// Parse trait declarations
		if traitInfo := a.parseTraitDeclaration(line, lineNum); traitInfo != nil {
			astInfo.Traits = append(astInfo.Traits, traitInfo)
		}

		// Parse impl blocks
		if implInfo := a.parseImplDeclaration(line, lineNum); implInfo != nil {
			astInfo.Impls = append(astInfo.Impls, implInfo)
		}

		// Parse module declarations
		if moduleInfo := a.parseModuleDeclaration(line, lineNum); moduleInfo != nil {
			astInfo.Modules = append(astInfo.Modules, moduleInfo)
		}

		// Parse constant declarations
		if constInfo := a.parseConstantDeclaration(line, lineNum); constInfo != nil {
			astInfo.Constants = append(astInfo.Constants, constInfo)
		}

		// Parse use declarations
		if useInfo := a.parseUseDeclaration(line, lineNum); useInfo != nil {
			astInfo.Uses = append(astInfo.Uses, useInfo)
		}

		// Parse macro declarations
		if macroInfo := a.parseMacroDeclaration(line, lineNum); macroInfo != nil {
			astInfo.Macros = append(astInfo.Macros, macroInfo)
		}
	}
}

// extractCrateName attempts to extract crate name from the file content
func (a *RustASTAnalyzer) extractCrateName(content string) string {
	// For now, return empty string - will be enhanced with proper parsing
	return ""
}

// parseFunctionDeclaration parses Rust function declarations
func (a *RustASTAnalyzer) parseFunctionDeclaration(line string, lineNum int) *types.RustFunctionInfo {
	// Regex pattern for Rust function declarations
	// Matches: [visibility] [async] [unsafe] [const] fn function_name
	fnRegex := regexp.MustCompile(`^\s*(pub(?:\([^)]*\))?\s+)?(async\s+)?(unsafe\s+)?(const\s+)?fn\s+(\w+)\s*\(`)
	
	matches := fnRegex.FindStringSubmatch(line)
	if matches == nil {
		return nil
	}

	visibility := strings.TrimSpace(matches[1])
	if visibility == "" {
		visibility = "private"
	}

	return &types.RustFunctionInfo{
		Name:           matches[5],
		StartLine:      lineNum,
		EndLine:        lineNum, // Will be updated when we find the closing brace
		StartColumn:    strings.Index(line, "fn") + 1,
		EndColumn:      len(line),
		Parameters:     []types.RustParameterInfo{}, // TODO: Parse parameters
		ReturnType:     "",                          // TODO: Parse return type
		IsPublic:       strings.Contains(visibility, "pub"),
		IsAsync:        matches[2] != "",
		IsUnsafe:       matches[3] != "",
		IsConst:        matches[4] != "",
		Complexity:     1, // TODO: Calculate actual complexity
		LineCount:      1, // TODO: Calculate actual line count
		HasDocComments: false, // TODO: Check for doc comments above
		Visibility:     visibility,
	}
}

// parseStructDeclaration parses Rust struct declarations
func (a *RustASTAnalyzer) parseStructDeclaration(line string, lineNum int) *types.RustStructInfo {
	structRegex := regexp.MustCompile(`^\s*(pub(?:\([^)]*\))?\s+)?struct\s+(\w+)`)
	
	matches := structRegex.FindStringSubmatch(line)
	if matches == nil {
		return nil
	}

	visibility := strings.TrimSpace(matches[1])
	if visibility == "" {
		visibility = "private"
	}

	return &types.RustStructInfo{
		Name:           matches[2],
		StartLine:      lineNum,
		EndLine:        lineNum,
		StartColumn:    strings.Index(line, "struct") + 1,
		EndColumn:      len(line),
		IsPublic:       strings.Contains(visibility, "pub"),
		FieldCount:     0, // TODO: Count fields
		Visibility:     visibility,
		HasDocComments: false, // TODO: Check for doc comments
	}
}

// parseEnumDeclaration parses Rust enum declarations
func (a *RustASTAnalyzer) parseEnumDeclaration(line string, lineNum int) *types.RustEnumInfo {
	enumRegex := regexp.MustCompile(`^\s*(pub(?:\([^)]*\))?\s+)?enum\s+(\w+)`)
	
	matches := enumRegex.FindStringSubmatch(line)
	if matches == nil {
		return nil
	}

	visibility := strings.TrimSpace(matches[1])
	if visibility == "" {
		visibility = "private"
	}

	return &types.RustEnumInfo{
		Name:           matches[2],
		StartLine:      lineNum,
		EndLine:        lineNum,
		StartColumn:    strings.Index(line, "enum") + 1,
		EndColumn:      len(line),
		IsPublic:       strings.Contains(visibility, "pub"),
		VariantCount:   0, // TODO: Count variants
		Visibility:     visibility,
		HasDocComments: false, // TODO: Check for doc comments
	}
}

// parseTraitDeclaration parses Rust trait declarations
func (a *RustASTAnalyzer) parseTraitDeclaration(line string, lineNum int) *types.RustTraitInfo {
	traitRegex := regexp.MustCompile(`^\s*(pub(?:\([^)]*\))?\s+)?trait\s+(\w+)`)
	
	matches := traitRegex.FindStringSubmatch(line)
	if matches == nil {
		return nil
	}

	visibility := strings.TrimSpace(matches[1])
	if visibility == "" {
		visibility = "private"
	}

	return &types.RustTraitInfo{
		Name:           matches[2],
		StartLine:      lineNum,
		EndLine:        lineNum,
		StartColumn:    strings.Index(line, "trait") + 1,
		EndColumn:      len(line),
		IsPublic:       strings.Contains(visibility, "pub"),
		MethodCount:    0, // TODO: Count methods
		Visibility:     visibility,
		HasDocComments: false, // TODO: Check for doc comments
	}
}

// parseImplDeclaration parses Rust impl blocks
func (a *RustASTAnalyzer) parseImplDeclaration(line string, lineNum int) *types.RustImplInfo {
	implRegex := regexp.MustCompile(`^\s*impl\s+(?:(\w+)\s+for\s+)?(\w+)`)
	
	matches := implRegex.FindStringSubmatch(line)
	if matches == nil {
		return nil
	}

	return &types.RustImplInfo{
		StartLine:   lineNum,
		EndLine:     lineNum,
		StartColumn: strings.Index(line, "impl") + 1,
		EndColumn:   len(line),
		TargetType:  matches[2],
		TraitName:   matches[1], // Empty for inherent impls
		MethodCount: 0,          // TODO: Count methods
	}
}

// parseModuleDeclaration parses Rust module declarations
func (a *RustASTAnalyzer) parseModuleDeclaration(line string, lineNum int) *types.RustModuleInfo {
	moduleRegex := regexp.MustCompile(`^\s*(pub(?:\([^)]*\))?\s+)?mod\s+(\w+)`)
	
	matches := moduleRegex.FindStringSubmatch(line)
	if matches == nil {
		return nil
	}

	visibility := strings.TrimSpace(matches[1])
	if visibility == "" {
		visibility = "private"
	}

	return &types.RustModuleInfo{
		Name:           matches[2],
		StartLine:      lineNum,
		EndLine:        lineNum,
		StartColumn:    strings.Index(line, "mod") + 1,
		EndColumn:      len(line),
		IsPublic:       strings.Contains(visibility, "pub"),
		Visibility:     visibility,
		HasDocComments: false, // TODO: Check for doc comments
	}
}

// parseConstantDeclaration parses Rust constant declarations
func (a *RustASTAnalyzer) parseConstantDeclaration(line string, lineNum int) *types.RustConstantInfo {
	constRegex := regexp.MustCompile(`^\s*(pub(?:\([^)]*\))?\s+)?const\s+(\w+)\s*:\s*([^=]+)`)
	
	matches := constRegex.FindStringSubmatch(line)
	if matches == nil {
		return nil
	}

	visibility := strings.TrimSpace(matches[1])
	if visibility == "" {
		visibility = "private"
	}

	return &types.RustConstantInfo{
		Name:           matches[2],
		Type:           strings.TrimSpace(matches[3]),
		Line:           lineNum,
		Column:         strings.Index(line, "const") + 1,
		IsPublic:       strings.Contains(visibility, "pub"),
		Visibility:     visibility,
		HasDocComments: false, // TODO: Check for doc comments
	}
}

// parseUseDeclaration parses Rust use declarations
func (a *RustASTAnalyzer) parseUseDeclaration(line string, lineNum int) *types.RustUseInfo {
	useRegex := regexp.MustCompile(`^\s*(pub(?:\([^)]*\))?\s+)?use\s+([^;]+)`)
	
	matches := useRegex.FindStringSubmatch(line)
	if matches == nil {
		return nil
	}

	visibility := strings.TrimSpace(matches[1])
	path := strings.TrimSpace(matches[2])

	// Handle aliasing (use path as alias)
	alias := ""
	if strings.Contains(path, " as ") {
		parts := strings.SplitN(path, " as ", 2)
		path = strings.TrimSpace(parts[0])
		alias = strings.TrimSpace(parts[1])
	}

	return &types.RustUseInfo{
		Path:       path,
		Alias:      alias,
		Line:       lineNum,
		Column:     strings.Index(line, "use") + 1,
		Visibility: visibility,
	}
}

// parseMacroDeclaration parses Rust macro declarations
func (a *RustASTAnalyzer) parseMacroDeclaration(line string, lineNum int) *types.RustMacroInfo {
	macroRegex := regexp.MustCompile(`^\s*macro_rules!\s+(\w+)`)
	
	matches := macroRegex.FindStringSubmatch(line)
	if matches == nil {
		return nil
	}

	return &types.RustMacroInfo{
		Name:           matches[1],
		StartLine:      lineNum,
		EndLine:        lineNum,
		StartColumn:    strings.Index(line, "macro_rules!") + 1,
		EndColumn:      len(line),
		IsPublic:       false, // TODO: Determine visibility
		MacroType:      "macro_rules!",
		HasDocComments: false, // TODO: Check for doc comments
	}
}

// calculateCyclomaticComplexity calculates cyclomatic complexity for Rust functions
// This is a simplified version that will be enhanced with proper AST analysis
func (a *RustASTAnalyzer) calculateCyclomaticComplexity(content string) int {
	complexity := 1 // Base complexity

	// Count control flow statements
	patterns := []string{
		`\bif\b`, `\belse\s+if\b`, `\bwhile\b`, `\bfor\b`, `\bloop\b`,
		`\bmatch\b`, `\bcontinue\b`, `\bbreak\b`, `\breturn\b`,
	}

	for _, pattern := range patterns {
		regex := regexp.MustCompile(pattern)
		matches := regex.FindAllString(content, -1)
		complexity += len(matches)
	}

	return complexity
}