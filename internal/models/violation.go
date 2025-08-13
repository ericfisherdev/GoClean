package models

// Severity represents the severity level of a violation
type Severity int

const (
	SeverityLow Severity = iota
	SeverityMedium
	SeverityHigh
	SeverityCritical
)

func (s Severity) String() string {
	switch s {
	case SeverityLow:
		return "Low"
	case SeverityMedium:
		return "Medium"
	case SeverityHigh:
		return "High"
	case SeverityCritical:
		return "Critical"
	default:
		return "Unknown"
	}
}

// ViolationType represents the category of violation
type ViolationType string

const (
	ViolationTypeFunctionLength    ViolationType = "function_length"
	ViolationTypeFunctionComplexity ViolationType = "function_complexity"
	ViolationTypeFunctionParameters ViolationType = "function_parameters"
	ViolationTypeNesting           ViolationType = "nesting_depth"
	ViolationTypeNaming            ViolationType = "naming_convention"
	ViolationTypeClassSize         ViolationType = "class_size"
	ViolationTypeDocumentation     ViolationType = "documentation"
	ViolationTypeMagicNumbers      ViolationType = "magic_numbers"
	ViolationTypeDuplication       ViolationType = "code_duplication"
)

// Violation represents a clean code violation found during scanning
type Violation struct {
	ID          string        `json:"id"`
	Type        ViolationType `json:"type"`
	Severity    Severity      `json:"severity"`
	Message     string        `json:"message"`
	Description string        `json:"description"`
	File        string        `json:"file"`
	Line        int           `json:"line"`
	Column      int           `json:"column"`
	EndLine     int           `json:"end_line,omitempty"`
	EndColumn   int           `json:"end_column,omitempty"`
	Context     string        `json:"context,omitempty"`
	Rule        string        `json:"rule"`
	Suggestion  string        `json:"suggestion,omitempty"`
}

// Location represents a position in source code
type Location struct {
	File      string `json:"file"`
	Line      int    `json:"line"`
	Column    int    `json:"column"`
	EndLine   int    `json:"end_line,omitempty"`
	EndColumn int    `json:"end_column,omitempty"`
}