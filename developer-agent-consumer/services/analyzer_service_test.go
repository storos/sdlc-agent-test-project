package services

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestAnalyzeRepository_GoProject(t *testing.T) {
	// Create temporary test directory
	tempDir := createTestGoProject(t)
	defer os.RemoveAll(tempDir)

	logger := logrus.New()
	logger.SetOutput(os.Stdout)

	service := NewAnalyzerService(logger)
	analysis, err := service.AnalyzeRepository(tempDir)
	if err != nil {
		t.Fatalf("Failed to analyze repository: %v", err)
	}

	// Validate entry points
	if len(analysis.EntryPoints) == 0 {
		t.Error("Expected to find entry points")
	}

	hasMainGo := false
	for _, ep := range analysis.EntryPoints {
		if filepath.Base(ep) == "main.go" {
			hasMainGo = true
			break
		}
	}
	if !hasMainGo {
		t.Error("Expected to find main.go in entry points")
	}

	// Validate languages
	if !contains(analysis.Languages, "Go") {
		t.Error("Expected to detect Go language")
	}

	// Validate project type
	if analysis.ProjectType != "Go Application" {
		t.Errorf("Expected project type 'Go Application', got '%s'", analysis.ProjectType)
	}

	// Validate config files
	hasGoMod := false
	for _, cf := range analysis.ConfigFiles {
		if filepath.Base(cf) == "go.mod" {
			hasGoMod = true
			break
		}
	}
	if !hasGoMod {
		t.Error("Expected to find go.mod in config files")
	}

	// Validate dependency managers
	if !contains(analysis.DependencyManagers, "Go Modules") {
		t.Error("Expected to detect Go Modules")
	}
}

func TestAnalyzeRepository_NodeProject(t *testing.T) {
	// Create temporary test directory
	tempDir := createTestNodeProject(t)
	defer os.RemoveAll(tempDir)

	logger := logrus.New()
	logger.SetOutput(os.Stdout)

	service := NewAnalyzerService(logger)
	analysis, err := service.AnalyzeRepository(tempDir)
	if err != nil {
		t.Fatalf("Failed to analyze repository: %v", err)
	}

	// Validate languages
	if !contains(analysis.Languages, "JavaScript") && !contains(analysis.Languages, "TypeScript") {
		t.Error("Expected to detect JavaScript or TypeScript language")
	}

	// Validate config files
	hasPackageJson := false
	for _, cf := range analysis.ConfigFiles {
		if filepath.Base(cf) == "package.json" {
			hasPackageJson = true
			break
		}
	}
	if !hasPackageJson {
		t.Error("Expected to find package.json in config files")
	}

	// Validate dependency managers
	if !contains(analysis.DependencyManagers, "npm/yarn") {
		t.Error("Expected to detect npm/yarn")
	}
}

func TestAnalyzeRepository_CleanArchitecture(t *testing.T) {
	// Create temporary test directory with clean architecture
	tempDir := createTestCleanArchitectureProject(t)
	defer os.RemoveAll(tempDir)

	logger := logrus.New()
	logger.SetOutput(os.Stdout)

	service := NewAnalyzerService(logger)
	analysis, err := service.AnalyzeRepository(tempDir)
	if err != nil {
		t.Fatalf("Failed to analyze repository: %v", err)
	}

	// Validate key directories
	expectedDirs := []string{"handlers", "services", "models", "repositories"}
	for _, dir := range expectedDirs {
		found := false
		for _, keyDir := range analysis.KeyDirectories {
			if filepath.Base(keyDir) == dir {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find directory: %s", dir)
		}
	}

	// Validate architecture pattern
	if arch, ok := analysis.Patterns["architecture"]; ok {
		if arch != "Clean Architecture (Handlers -> Services -> Repositories -> Models)" {
			t.Errorf("Expected Clean Architecture pattern, got: %s", arch)
		}
	} else {
		t.Error("Expected architecture pattern to be detected")
	}
}

func createTestGoProject(t *testing.T) string {
	tempDir, err := os.MkdirTemp("", "test-go-project-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create main.go
	mainGo := filepath.Join(tempDir, "main.go")
	os.WriteFile(mainGo, []byte("package main\n\nfunc main() {}\n"), 0644)

	// Create go.mod
	goMod := filepath.Join(tempDir, "go.mod")
	os.WriteFile(goMod, []byte("module test\n\ngo 1.20\n"), 0644)

	// Create some Go files
	os.WriteFile(filepath.Join(tempDir, "handler.go"), []byte("package main\n"), 0644)
	os.WriteFile(filepath.Join(tempDir, "service.go"), []byte("package main\n"), 0644)

	return tempDir
}

func createTestNodeProject(t *testing.T) string {
	tempDir, err := os.MkdirTemp("", "test-node-project-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create index.js
	indexJs := filepath.Join(tempDir, "index.js")
	os.WriteFile(indexJs, []byte("console.log('Hello');\n"), 0644)

	// Create package.json
	packageJson := filepath.Join(tempDir, "package.json")
	os.WriteFile(packageJson, []byte("{\"name\": \"test\"}\n"), 0644)

	// Create some JS/TS files
	os.WriteFile(filepath.Join(tempDir, "app.js"), []byte("// App\n"), 0644)
	os.WriteFile(filepath.Join(tempDir, "utils.ts"), []byte("// Utils\n"), 0644)

	return tempDir
}

func createTestCleanArchitectureProject(t *testing.T) string {
	tempDir, err := os.MkdirTemp("", "test-clean-arch-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create directory structure
	dirs := []string{"handlers", "services", "models", "repositories"}
	for _, dir := range dirs {
		os.MkdirAll(filepath.Join(tempDir, dir), 0755)
		// Add a dummy file in each directory
		os.WriteFile(filepath.Join(tempDir, dir, "dummy.go"), []byte("package "+dir+"\n"), 0644)
	}

	// Create main.go
	os.WriteFile(filepath.Join(tempDir, "main.go"), []byte("package main\n"), 0644)

	// Create go.mod
	os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte("module test\n\ngo 1.20\n"), 0644)

	return tempDir
}
