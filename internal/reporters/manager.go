package reporters

import (
	"fmt"
	"sync"

	"github.com/ericfisherdev/goclean/internal/config"
	"github.com/ericfisherdev/goclean/internal/models"
)

// Reporter interface defines the contract for all reporters
type Reporter interface {
	Generate(report *models.Report) error
}

// Manager coordinates multiple reporters
type Manager struct {
	reporters []Reporter
	config    *config.Config
}

// NewManager creates a new reporter manager
func NewManager(cfg *config.Config) (*Manager, error) {
	manager := &Manager{
		reporters: make([]Reporter, 0),
		config:    cfg,
	}

	// Initialize HTML reporter if configured
	if cfg.Output.HTML.Path != "" {
		htmlConfig := &HTMLConfig{
			OutputPath:      cfg.Output.HTML.Path,
			AutoRefresh:     cfg.Output.HTML.AutoRefresh,
			RefreshInterval: cfg.Output.HTML.RefreshInterval,
			Theme:           cfg.Output.HTML.Theme,
		}

		htmlReporter, err := NewHTMLReporter(htmlConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create HTML reporter: %w", err)
		}

		manager.reporters = append(manager.reporters, htmlReporter)
	}

	// Initialize Markdown reporter if configured
	if cfg.Output.Markdown.Enabled && cfg.Output.Markdown.Path != "" {
		markdownConfig := &MarkdownConfig{
			OutputPath:      cfg.Output.Markdown.Path,
			IncludeExamples: cfg.Output.Markdown.IncludeExamples,
		}

		markdownReporter := NewMarkdownReporter(markdownConfig)
		manager.reporters = append(manager.reporters, markdownReporter)
	}

	return manager, nil
}

// GenerateReports creates all configured reports
func (m *Manager) GenerateReports(summary *models.ScanSummary, files []*models.ScanResult) error {
	// Create report config from current config
	reportConfig := &models.ReportConfig{
		Paths:     m.config.Scan.Paths,
		FileTypes: m.config.Scan.FileTypes,
		Thresholds: &models.Thresholds{
			FunctionLines:        m.config.Thresholds.FunctionLines,
			CyclomaticComplexity: m.config.Thresholds.CyclomaticComplexity,
			Parameters:           m.config.Thresholds.Parameters,
			NestingDepth:         m.config.Thresholds.NestingDepth,
			ClassLines:           m.config.Thresholds.ClassLines,
		},
		HTMLSettings: &models.HTMLOptions{
			AutoRefresh:     m.config.Output.HTML.AutoRefresh,
			RefreshInterval: m.config.Output.HTML.RefreshInterval,
			Theme:           m.config.Output.HTML.Theme,
		},
	}

	// Create the report
	report := models.NewReport(summary, files, reportConfig)

	// Generate all reports concurrently
	var wg sync.WaitGroup
	errorChan := make(chan error, len(m.reporters))

	for _, reporter := range m.reporters {
		wg.Add(1)
		go func(r Reporter) {
			defer wg.Done()
			if err := r.Generate(report); err != nil {
				errorChan <- err
			}
		}(reporter)
	}

	wg.Wait()
	close(errorChan)

	// Check for errors
	for err := range errorChan {
		return err
	}

	return nil
}

// GenerateConsoleReport generates a console report (doesn't require file output)
func (m *Manager) GenerateConsoleReport(summary *models.ScanSummary, files []*models.ScanResult, verbose, colors bool) error {
	// Create report config
	reportConfig := &models.ReportConfig{
		Paths:     m.config.Scan.Paths,
		FileTypes: m.config.Scan.FileTypes,
		Thresholds: &models.Thresholds{
			FunctionLines:        m.config.Thresholds.FunctionLines,
			CyclomaticComplexity: m.config.Thresholds.CyclomaticComplexity,
			Parameters:           m.config.Thresholds.Parameters,
			NestingDepth:         m.config.Thresholds.NestingDepth,
			ClassLines:           m.config.Thresholds.ClassLines,
		},
	}

	// Create the report
	report := models.NewReport(summary, files, reportConfig)

	// Create and use console reporter
	consoleReporter := NewConsoleReporter(verbose, colors)
	return consoleReporter.Generate(report)
}

// GenerateHTMLReportWithProgress generates an HTML report with progress updates
func (m *Manager) GenerateHTMLReportWithProgress(summary *models.ScanSummary, files []*models.ScanResult, progressFn func(string)) error {
	if m.config.Output.HTML.Path == "" {
		return fmt.Errorf("HTML output path not configured")
	}

	// Create report config
	reportConfig := &models.ReportConfig{
		Paths:     m.config.Scan.Paths,
		FileTypes: m.config.Scan.FileTypes,
		Thresholds: &models.Thresholds{
			FunctionLines:        m.config.Thresholds.FunctionLines,
			CyclomaticComplexity: m.config.Thresholds.CyclomaticComplexity,
			Parameters:           m.config.Thresholds.Parameters,
			NestingDepth:         m.config.Thresholds.NestingDepth,
			ClassLines:           m.config.Thresholds.ClassLines,
		},
		HTMLSettings: &models.HTMLOptions{
			AutoRefresh:     m.config.Output.HTML.AutoRefresh,
			RefreshInterval: m.config.Output.HTML.RefreshInterval,
			Theme:           m.config.Output.HTML.Theme,
		},
	}

	// Create the report
	report := models.NewReport(summary, files, reportConfig)

	// Create HTML reporter
	htmlConfig := &HTMLConfig{
		OutputPath:      m.config.Output.HTML.Path,
		AutoRefresh:     m.config.Output.HTML.AutoRefresh,
		RefreshInterval: m.config.Output.HTML.RefreshInterval,
		Theme:           m.config.Output.HTML.Theme,
	}

	htmlReporter, err := NewHTMLReporter(htmlConfig)
	if err != nil {
		return fmt.Errorf("failed to create HTML reporter: %w", err)
	}

	// Generate with progress
	return htmlReporter.GenerateWithProgress(report, progressFn)
}

// GetConfiguredReporters returns the list of configured reporter types
func (m *Manager) GetConfiguredReporters() []string {
	var types []string

	if m.config.Output.HTML.Path != "" {
		types = append(types, "HTML")
	}
	if m.config.Output.Markdown.Enabled && m.config.Output.Markdown.Path != "" {
		types = append(types, "Markdown")
	}

	return types
}

// GetHTMLOutputPath returns the configured HTML output path
func (m *Manager) GetHTMLOutputPath() string {
	return m.config.Output.HTML.Path
}

// GetMarkdownOutputPath returns the configured Markdown output path
func (m *Manager) GetMarkdownOutputPath() string {
	if m.config.Output.Markdown.Enabled {
		return m.config.Output.Markdown.Path
	}
	return ""
}