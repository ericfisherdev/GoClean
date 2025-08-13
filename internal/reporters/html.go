package reporters

import (
	"embed"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ericfisherdev/goclean/internal/models"
)

//go:embed templates/*
var templatesFS embed.FS

// HTMLReporter generates HTML reports
type HTMLReporter struct {
	template *template.Template
	config   *HTMLConfig
}

// HTMLConfig contains HTML-specific configuration
type HTMLConfig struct {
	OutputPath      string
	AutoRefresh     bool
	RefreshInterval int
	Theme           string
}

// NewHTMLReporter creates a new HTML reporter
func NewHTMLReporter(config *HTMLConfig) (*HTMLReporter, error) {
	// Load templates
	tmpl, err := template.New("").Funcs(getTemplateFunctions()).ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to load HTML templates: %w", err)
	}

	return &HTMLReporter{
		template: tmpl,
		config:   config,
	}, nil
}

// Generate creates an HTML report from the provided data
func (h *HTMLReporter) Generate(report *models.Report) error {
	// Ensure output directory exists
	outputDir := filepath.Dir(h.config.OutputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create output file
	file, err := os.Create(h.config.OutputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Prepare template data
	templateData := struct {
		*models.Report
		Config       *HTMLConfig
		GeneratedAt  string
		FileTree     *models.FileTreeNode
		RefreshMeta  template.HTML
	}{
		Report:      report,
		Config:      h.config,
		GeneratedAt: report.GeneratedAt.Format("2006-01-02 15:04:05"),
		FileTree:    report.BuildFileTree(),
		RefreshMeta: h.getRefreshMeta(),
	}

	// Execute template
	if err := h.template.ExecuteTemplate(file, "report.html", templateData); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

// getRefreshMeta returns meta tag for auto-refresh if enabled
func (h *HTMLReporter) getRefreshMeta() template.HTML {
	if h.config.AutoRefresh && h.config.RefreshInterval > 0 {
		return template.HTML(fmt.Sprintf(`<meta http-equiv="refresh" content="%d">`, h.config.RefreshInterval))
	}
	return ""
}

// getTemplateFunctions returns custom template functions
func getTemplateFunctions() template.FuncMap {
	return template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"percentage": func(part, total int) float64 {
			if total == 0 {
				return 0
			}
			return float64(part) / float64(total) * 100
		},
		"formatDuration": func(d time.Duration) string {
			if d.Hours() >= 1 {
				return fmt.Sprintf("%.1fh", d.Hours())
			} else if d.Minutes() >= 1 {
				return fmt.Sprintf("%.1fm", d.Minutes())
			}
			return fmt.Sprintf("%.1fs", d.Seconds())
		},
		"severityColor": func(s models.Severity) string {
			return s.GetColor()
		},
		"severityIcon": func(s models.Severity) string {
			switch s {
			case models.SeverityLow:
				return "info-circle"
			case models.SeverityMedium:
				return "exclamation-triangle"
			case models.SeverityHigh:
				return "exclamation-triangle"
			case models.SeverityCritical:
				return "exclamation-circle"
			default:
				return "question-circle"
			}
		},
		"violationTypeDisplay": func(vt models.ViolationType) string {
			return vt.GetDisplayName()
		},
		"truncate": func(s string, length int) string {
			if len(s) <= length {
				return s
			}
			return s[:length] + "..."
		},
		"basename": func(path string) string {
			return filepath.Base(path)
		},
		"dirname": func(path string) string {
			return filepath.Dir(path)
		},
		"replace": func(s, old, new string) string {
			return strings.ReplaceAll(s, old, new)
		},
		"hasViolations": func(violations []*models.Violation) bool {
			return len(violations) > 0
		},
		"first": func(slice []*models.Violation, n int) []*models.Violation {
			if len(slice) <= n {
				return slice
			}
			return slice[:n]
		},
		"formatCode": func(code string) template.HTML {
			// Basic syntax highlighting - in a real implementation, 
			// you might want to use a proper syntax highlighter
			code = template.HTMLEscapeString(code)
			
			// Add some basic highlighting for Go keywords
			keywords := []string{"func", "var", "const", "type", "package", "import", "if", "else", "for", "range", "return", "struct", "interface"}
			for _, keyword := range keywords {
				code = strings.ReplaceAll(code, keyword+" ", fmt.Sprintf(`<span class="keyword">%s</span> `, keyword))
			}
			
			return template.HTML(code)
		},
		"themeClass": func(theme string) string {
			if theme == "dark" {
				return "dark-theme"
			}
			return ""
		},
		"severityBadge": func(s models.Severity) string {
			switch s {
			case models.SeverityLow:
				return "badge bg-success"
			case models.SeverityMedium:
				return "badge bg-warning"
			case models.SeverityHigh:
				return "badge bg-danger"
			case models.SeverityCritical:
				return "badge bg-danger"
			default:
				return "badge bg-secondary"
			}
		},
	}
}

// GenerateWithProgress generates an HTML report with progress updates
func (h *HTMLReporter) GenerateWithProgress(report *models.Report, progressFn func(string)) error {
	if progressFn != nil {
		progressFn("Preparing HTML report...")
	}

	// Generate the report
	if err := h.Generate(report); err != nil {
		return err
	}

	if progressFn != nil {
		progressFn(fmt.Sprintf("HTML report generated: %s", h.config.OutputPath))
	}

	return nil
}