package violations

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"testing"

	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/testutils"
	"github.com/ericfisherdev/goclean/internal/types"
)

// BenchmarkViolationDetection tests detection performance with different violation types
func BenchmarkViolationDetection(b *testing.B) {
	config := DefaultDetectorConfig()
	detectors := map[string]Detector{
		"Function":      NewFunctionDetector(config),
		"Naming":        NewNamingDetector(config),
		"Structure":     NewStructureDetector(config),
		"Documentation": NewDocumentationDetector(config),
		"TodoTracker":   NewTodoTrackerDetector(config),
		"CommentedCode": NewCommentedCodeDetector(config),
		"Duplication":   NewDuplicationDetector(config),
	}

	for name, detector := range detectors {
		b.Run(name, func(b *testing.B) {
			benchmarkDetector(b, detector)
		})
	}
}

func benchmarkDetector(b *testing.B, detector Detector) {
	helper := testutils.NewBenchmarkHelper(b)
	filePath := helper.CreateLargeTestFile(b, 500)
	
	// Parse the test file
	fset := token.NewFileSet()
	src := readTestFile(b, filePath)
	file, err := parser.ParseFile(fset, filePath, src, parser.ParseComments)
	if err != nil {
		b.Fatalf("Failed to parse test file: %v", err)
	}
	
	fileInfo := &models.FileInfo{
		Path:     filePath,
		Language: "go",
	}
	
	// Create proper GoASTInfo structure
	astInfo := &types.GoASTInfo{
		FilePath:    filePath,
		PackageName: file.Name.Name,
		AST:         file,
		FileSet:     fset,
		Functions:   []*types.FunctionInfo{}, // Would be populated by real AST analyzer
		Types:       []*types.TypeInfo{},
		Imports:     []*types.ImportInfo{},
		Variables:   []*types.VariableInfo{},
		Constants:   []*types.ConstantInfo{},
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		violations := detector.Detect(fileInfo, astInfo)
		b.ReportMetric(float64(len(violations)), "violations/op")
	}
}

// BenchmarkBatchViolationDetection tests detection performance with multiple files
func BenchmarkBatchViolationDetection(b *testing.B) {
	fileCounts := []int{10, 50, 100, 500}
	
	for _, count := range fileCounts {
		b.Run(fmt.Sprintf("Files_%d", count), func(b *testing.B) {
			benchmarkBatchDetection(b, count)
		})
	}
}

func benchmarkBatchDetection(b *testing.B, fileCount int) {
	helper := testutils.NewBenchmarkHelper(b)
	files := helper.CreateBenchmarkFiles(b, fileCount, 200)
	
	detector := NewFunctionDetector(DefaultDetectorConfig())
	fset := token.NewFileSet()
	
	// Pre-parse all files
	var astInfos []*types.GoASTInfo
	var fileInfos []*models.FileInfo
	
	for _, filePath := range files {
		src := readTestFile(b, filePath)
		file, err := parser.ParseFile(fset, filePath, src, parser.ParseComments)
		if err != nil {
			b.Fatalf("Failed to parse file %s: %v", filePath, err)
		}
		
		astInfo := &types.GoASTInfo{
			FilePath:    filePath,
			PackageName: file.Name.Name,
			AST:         file,
			FileSet:     fset,
			Functions:   []*types.FunctionInfo{},
			Types:       []*types.TypeInfo{},
			Imports:     []*types.ImportInfo{},
			Variables:   []*types.VariableInfo{},
			Constants:   []*types.ConstantInfo{},
		}
		
		astInfos = append(astInfos, astInfo)
		fileInfos = append(fileInfos, &models.FileInfo{
			Path:     filePath,
			Language: "go",
		})
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		totalViolations := 0
		for j, astInfo := range astInfos {
			violations := detector.Detect(fileInfos[j], astInfo)
			totalViolations += len(violations)
		}
		b.ReportMetric(float64(totalViolations), "violations/op")
	}
}

// BenchmarkDuplicationDetection tests code duplication detection performance
func BenchmarkDuplicationDetection(b *testing.B) {
	helper := testutils.NewBenchmarkHelper(b)
	files := helper.CreateBenchmarkFiles(b, 100, 150)
	
	detector := NewDuplicationDetector(DefaultDetectorConfig())
	fset := token.NewFileSet()
	
	var astInfos []*types.GoASTInfo
	var fileInfos []*models.FileInfo
	
	for _, filePath := range files {
		src := readTestFile(b, filePath)
		file, err := parser.ParseFile(fset, filePath, src, parser.ParseComments)
		if err != nil {
			b.Fatalf("Failed to parse file %s: %v", filePath, err)
		}
		
		astInfo := &types.GoASTInfo{
			FilePath:    filePath,
			PackageName: file.Name.Name,
			AST:         file,
			FileSet:     fset,
			Functions:   []*types.FunctionInfo{},
			Types:       []*types.TypeInfo{},
			Imports:     []*types.ImportInfo{},
			Variables:   []*types.VariableInfo{},
			Constants:   []*types.ConstantInfo{},
		}
		
		astInfos = append(astInfos, astInfo)
		fileInfos = append(fileInfos, &models.FileInfo{
			Path:     filePath,
			Language: "go",
		})
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		totalViolations := 0
		for j, astInfo := range astInfos {
			violations := detector.Detect(fileInfos[j], astInfo)
			totalViolations += len(violations)
		}
		b.ReportMetric(float64(totalViolations), "duplications/op")
	}
}

// BenchmarkViolationSeverityClassification tests severity assignment performance
func BenchmarkViolationSeverityClassification(b *testing.B) {
	violations := testutils.CreateViolationBatch(1000)
	config := DefaultDetectorConfig()
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		for j := range violations {
			_ = config.ClassifyViolationSeverity(violations[j].Type, 50, 25, nil)
		}
	}
}

// Helper function to read test file content
func readTestFile(b *testing.B, filePath string) []byte {
	b.Helper()
	src, err := os.ReadFile(filePath)
	if err != nil {
		b.Fatalf("Failed to read test file: %v", err)
	}
	return src
}