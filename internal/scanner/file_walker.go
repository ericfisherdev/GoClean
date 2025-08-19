package scanner

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/ericfisherdev/goclean/internal/models"
)

// FileWalker handles directory traversal and file discovery
type FileWalker struct {
	includePaths      []string
	excludePatterns   []string
	fileTypes         []string
	verbose           bool
	skipTestFiles     bool
	aggressiveMode    bool
	testPatterns      *TestFilePatterns
	customTestPatterns []string
	maxFileSize       int64 // Maximum file size in bytes, 0 means no limit
}

// NewFileWalker creates a new FileWalker instance
func NewFileWalker(includePaths []string, excludePatterns []string, fileTypes []string, verbose bool) *FileWalker {
	return &FileWalker{
		includePaths:      includePaths,
		excludePatterns:   excludePatterns,
		fileTypes:         fileTypes,
		verbose:           verbose,
		skipTestFiles:     true, // Default to skipping test files
		aggressiveMode:    false,
		testPatterns:      DefaultTestPatterns(),
		customTestPatterns: []string{},
		maxFileSize:       0, // No limit by default
	}
}

// NewFileWalkerWithConfig creates a new FileWalker with test file configuration
func NewFileWalkerWithConfig(includePaths []string, excludePatterns []string, fileTypes []string, verbose bool, 
	skipTestFiles bool, aggressiveMode bool, customTestPatterns []string) *FileWalker {
	walker := &FileWalker{
		includePaths:       includePaths,
		excludePatterns:    excludePatterns,
		fileTypes:          fileTypes,
		verbose:            verbose,
		skipTestFiles:      skipTestFiles,
		aggressiveMode:     aggressiveMode,
		testPatterns:       DefaultTestPatterns(),
		customTestPatterns: customTestPatterns,
	}
	
	// Add custom test patterns if provided
	if len(customTestPatterns) > 0 {
		walker.testPatterns.AddCustomPatterns(customTestPatterns)
	}
	
	return walker
}

// Walk traverses the specified paths and returns a list of files to scan
func (fw *FileWalker) Walk() ([]*models.FileInfo, error) {
	var files []*models.FileInfo
	
	for _, path := range fw.includePaths {
		pathFiles, err := fw.walkPath(path)
		if err != nil {
			return nil, fmt.Errorf("error walking path %s: %w", path, err)
		}
		files = append(files, pathFiles...)
	}
	
	if fw.verbose {
		fmt.Printf("Discovered %d files for scanning\n", len(files))
	}
	
	return files, nil
}

// walkPath walks a single path and returns discovered files
func (fw *FileWalker) walkPath(rootPath string) ([]*models.FileInfo, error) {
	var files []*models.FileInfo
	
	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if fw.verbose {
				fmt.Printf("Warning: skipping %s due to error: %v\n", path, err)
			}
			return nil // Continue walking, don't fail on individual file errors
		}
		
		// Skip directories
		if d.IsDir() {
			// Check if directory should be excluded
			if fw.shouldExclude(path) {
				if fw.verbose {
					fmt.Printf("Excluding directory: %s\n", path)
				}
				return filepath.SkipDir
			}
			return nil
		}
		
		// Check if file should be excluded
		if fw.shouldExclude(path) {
			if fw.verbose {
				fmt.Printf("Excluding file: %s\n", path)
			}
			return nil
		}
		
		// Check if test file should be skipped
		if fw.shouldSkipTestFile(filepath.ToSlash(path)) {
			if fw.verbose {
				fmt.Printf("Skipping test file: %s\n", path)
			}
			return nil
		}
		
		// Check file type filter
		if !fw.isAllowedFileType(path) {
			return nil
		}
		
		// Get file info
		info, err := d.Info()
		if err != nil {
			if fw.verbose {
				fmt.Printf("Warning: cannot get info for %s: %v\n", path, err)
			}
			return nil
		}
		
		// Check file size limit
		if !fw.isFileSizeAllowed(info.Size()) {
			if fw.verbose {
				fmt.Printf("Skipping file due to size limit: %s (%d bytes)\n", path, info.Size())
			}
			return nil
		}
		
		fileInfo := &models.FileInfo{
			Path:         path,
			Name:         d.Name(),
			Extension:    filepath.Ext(path),
			Size:         info.Size(),
			ModifiedTime: info.ModTime(),
			Language:     fw.detectLanguage(path),
			Scanned:      false,
		}
		
		files = append(files, fileInfo)
		
		if fw.verbose {
			fmt.Printf("Found file: %s (%s)\n", path, fileInfo.Language)
		}
		
		return nil
	})
	
	return files, err
}

// shouldExclude checks if a path matches any exclude pattern
func (fw *FileWalker) shouldExclude(path string) bool {
	for _, pattern := range fw.excludePatterns {
		if fw.matchesPattern(path, pattern) {
			return true
		}
	}
	return false
}

// shouldSkipTestFile determines if a test file should be skipped based on configuration
func (fw *FileWalker) shouldSkipTestFile(path string) bool {
	// If in aggressive mode, don't skip any test files
	if fw.aggressiveMode {
		return false
	}
	
	// If not skipping test files, don't skip
	if !fw.skipTestFiles {
		return false
	}
	
	// Check if this is a test file
	return fw.testPatterns.IsTestFile(path)
}

// matchesPattern checks if a path matches a pattern.
// It is a simple implementation that can be enhanced with glob patterns.
func (fw *FileWalker) matchesPattern(path, pattern string) bool {
	// Normalize both path and pattern to use forward slashes for consistent matching.
	path = filepath.ToSlash(path)
	pattern = filepath.ToSlash(strings.TrimSpace(pattern))

	// Handle directory patterns (e.g., "vendor/")
	if strings.HasSuffix(pattern, "/") {
		// Match if the path contains the pattern as a directory component.
		return strings.Contains(path, "/"+pattern) || strings.HasPrefix(path, pattern)
	}

	// Handle file extension patterns (e.g., "*.go")
	if strings.HasPrefix(pattern, "*.") {
		return strings.HasSuffix(path, pattern[1:])
	}

	// Handle suffix patterns (e.g., "_test.go")
	if strings.HasSuffix(pattern, ".go") || strings.Contains(pattern, "_") {
		return strings.Contains(path, pattern)
	}

	// For other patterns, check if any path component matches the pattern exactly.
	// This is more robust than a simple substring search.
	// e.g., "tests" will match "/path/to/tests/file.go" but not "/path/to/mytests/file.go".
	for _, part := range strings.Split(path, "/") {
		if part == pattern {
			return true
		}
	}

	return false
}

// isAllowedFileType checks if the file type is allowed for scanning
func (fw *FileWalker) isAllowedFileType(path string) bool {
	// If no file types specified, allow all
	if len(fw.fileTypes) == 0 {
		return fw.isSupportedLanguage(path)
	}
	
	ext := filepath.Ext(path)
	for _, allowedType := range fw.fileTypes {
		if ext == allowedType {
			return true
		}
	}
	
	return false
}

// isSupportedLanguage checks if the file extension is supported
func (fw *FileWalker) isSupportedLanguage(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	supportedExts := []string{
		".go", ".js", ".ts", ".jsx", ".tsx", ".py", ".java", ".cs", 
		".c", ".cpp", ".cc", ".cxx", ".h", ".hpp", ".rb", ".php",
		".swift", ".kt", ".scala", ".rs",
	}
	
	for _, supported := range supportedExts {
		if ext == supported {
			return true
		}
	}
	
	return false
}

// detectLanguage determines the programming language based on file extension
func (fw *FileWalker) detectLanguage(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	
	languageMap := map[string]string{
		".go":   "Go",
		".js":   "JavaScript",
		".ts":   "TypeScript",
		".jsx":  "JavaScript",
		".tsx":  "TypeScript",
		".py":   "Python",
		".java": "Java",
		".cs":   "C#",
		".c":    "C",
		".cpp":  "C++",
		".cc":   "C++",
		".cxx":  "C++",
		".h":    "C",
		".hpp":  "C++",
		".rb":   "Ruby",
		".php":  "PHP",
		".swift": "Swift",
		".kt":   "Kotlin",
		".scala": "Scala",
		".rs":   "Rust",
	}
	
	if lang, exists := languageMap[ext]; exists {
		return lang
	}
	
	return "Unknown"
}

// parseSizeString parses size strings like "1MB", "500KB", "2GB" into bytes
func parseSizeString(sizeStr string) (int64, error) {
	if sizeStr == "" {
		return 0, nil // No limit
	}
	
	// Regular expression to match size format like "1MB", "500KB", etc.
	re := regexp.MustCompile(`^(\d+(?:\.\d+)?)\s*(B|KB|MB|GB|TB)?$`)
	matches := re.FindStringSubmatch(strings.ToUpper(strings.TrimSpace(sizeStr)))
	
	if len(matches) == 0 {
		return 0, fmt.Errorf("invalid size format: %s", sizeStr)
	}
	
	// Parse the numeric part
	size, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, fmt.Errorf("invalid numeric value in size: %s", sizeStr)
	}
	
	// Get the unit (default to bytes if not specified)
	unit := "B"
	if len(matches) > 2 && matches[2] != "" {
		unit = matches[2]
	}
	
	// Convert to bytes
	var multiplier int64
	switch unit {
	case "B":
		multiplier = 1
	case "KB":
		multiplier = 1024
	case "MB":
		multiplier = 1024 * 1024
	case "GB":
		multiplier = 1024 * 1024 * 1024
	case "TB":
		multiplier = 1024 * 1024 * 1024 * 1024
	default:
		return 0, fmt.Errorf("unsupported size unit: %s", unit)
	}
	
	return int64(size * float64(multiplier)), nil
}

// SetMaxFileSize sets the maximum file size limit
func (fw *FileWalker) SetMaxFileSize(maxSizeStr string) error {
	maxSize, err := parseSizeString(maxSizeStr)
	if err != nil {
		return err
	}
	fw.maxFileSize = maxSize
	return nil
}

// isFileSizeAllowed checks if the file size is within the allowed limit
func (fw *FileWalker) isFileSizeAllowed(fileSize int64) bool {
	// If no limit is set (maxFileSize == 0), allow all files
	if fw.maxFileSize == 0 {
		return true
	}
	
	return fileSize <= fw.maxFileSize
}