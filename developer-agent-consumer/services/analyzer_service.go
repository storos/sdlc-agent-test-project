package services

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/storos/sdlc-agent/developer-agent-consumer/models"
)

type AnalyzerService struct {
	logger *logrus.Logger
}

func NewAnalyzerService(logger *logrus.Logger) *AnalyzerService {
	return &AnalyzerService{
		logger: logger,
	}
}

var (
	// Entry point files
	entryPointFiles = []string{
		"main.go", "index.js", "index.ts", "app.js", "app.ts",
		"server.js", "server.ts", "index.html", "App.tsx", "App.jsx",
	}

	// Key directories to look for
	keyDirectories = []string{
		"handlers", "controllers", "routes", "api",
		"services", "business", "logic",
		"models", "entities", "schemas",
		"repositories", "data", "db",
		"utils", "helpers", "lib",
		"middleware", "middlewares",
		"config", "configuration",
		"tests", "test", "__tests__",
	}

	// Configuration files
	configFiles = []string{
		"go.mod", "package.json", "requirements.txt", "Pipfile",
		"pom.xml", "build.gradle", "Cargo.toml", "composer.json",
		".env.example", "config.yaml", "config.yml", "config.json",
		"Dockerfile", "docker-compose.yml",
	}

	// Language indicators
	languageExtensions = map[string]string{
		".go":   "Go",
		".js":   "JavaScript",
		".ts":   "TypeScript",
		".py":   "Python",
		".java": "Java",
		".rb":   "Ruby",
		".php":  "PHP",
		".cs":   "C#",
		".rs":   "Rust",
		".cpp":  "C++",
		".c":    "C",
	}
)

func (s *AnalyzerService) AnalyzeRepository(repoPath string) (*models.RepositoryAnalysis, error) {
	s.logger.WithFields(logrus.Fields{
		"path": repoPath,
	}).Info("Analyzing repository structure")

	analysis := &models.RepositoryAnalysis{
		EntryPoints:        []string{},
		KeyDirectories:     []string{},
		ConfigFiles:        []string{},
		Languages:          []string{},
		Patterns:           make(map[string]string),
		DependencyManagers: []string{},
	}

	// Scan directory
	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden files and directories
		if strings.HasPrefix(info.Name(), ".") && info.Name() != ".env.example" {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip common ignore directories
		if info.IsDir() {
			dirName := info.Name()
			if dirName == "node_modules" || dirName == "vendor" || dirName == "venv" || dirName == "__pycache__" {
				return filepath.SkipDir
			}
		}

		relPath, _ := filepath.Rel(repoPath, path)

		// Check for entry points
		if !info.IsDir() {
			for _, entryPoint := range entryPointFiles {
				if info.Name() == entryPoint {
					analysis.EntryPoints = append(analysis.EntryPoints, relPath)
				}
			}

			// Check for config files
			for _, configFile := range configFiles {
				if info.Name() == configFile {
					analysis.ConfigFiles = append(analysis.ConfigFiles, relPath)
				}
			}

			// Detect languages
			ext := filepath.Ext(info.Name())
			if lang, ok := languageExtensions[ext]; ok {
				if !contains(analysis.Languages, lang) {
					analysis.Languages = append(analysis.Languages, lang)
				}
			}
		}

		// Check for key directories
		if info.IsDir() {
			for _, keyDir := range keyDirectories {
				if strings.EqualFold(info.Name(), keyDir) {
					analysis.KeyDirectories = append(analysis.KeyDirectories, relPath)
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Determine project type
	analysis.ProjectType = s.detectProjectType(analysis)

	// Detect dependency managers
	analysis.DependencyManagers = s.detectDependencyManagers(analysis)

	// Detect common patterns
	s.detectPatterns(analysis)

	s.logger.WithFields(logrus.Fields{
		"entry_points":   len(analysis.EntryPoints),
		"key_dirs":       len(analysis.KeyDirectories),
		"config_files":   len(analysis.ConfigFiles),
		"languages":      analysis.Languages,
		"project_type":   analysis.ProjectType,
	}).Info("Repository analysis complete")

	return analysis, nil
}

func (s *AnalyzerService) detectProjectType(analysis *models.RepositoryAnalysis) string {
	// Check for specific config files
	for _, configFile := range analysis.ConfigFiles {
		if strings.Contains(configFile, "go.mod") {
			return "Go Application"
		}
		if strings.Contains(configFile, "package.json") {
			return "Node.js Application"
		}
		if strings.Contains(configFile, "requirements.txt") || strings.Contains(configFile, "Pipfile") {
			return "Python Application"
		}
		if strings.Contains(configFile, "pom.xml") || strings.Contains(configFile, "build.gradle") {
			return "Java Application"
		}
		if strings.Contains(configFile, "Cargo.toml") {
			return "Rust Application"
		}
	}

	// Default to primary language
	if len(analysis.Languages) > 0 {
		return analysis.Languages[0] + " Application"
	}

	return "Unknown"
}

func (s *AnalyzerService) detectDependencyManagers(analysis *models.RepositoryAnalysis) []string {
	managers := []string{}

	for _, configFile := range analysis.ConfigFiles {
		if strings.Contains(configFile, "go.mod") {
			managers = append(managers, "Go Modules")
		}
		if strings.Contains(configFile, "package.json") {
			managers = append(managers, "npm/yarn")
		}
		if strings.Contains(configFile, "requirements.txt") {
			managers = append(managers, "pip")
		}
		if strings.Contains(configFile, "Pipfile") {
			managers = append(managers, "pipenv")
		}
		if strings.Contains(configFile, "pom.xml") {
			managers = append(managers, "Maven")
		}
		if strings.Contains(configFile, "build.gradle") {
			managers = append(managers, "Gradle")
		}
		if strings.Contains(configFile, "Cargo.toml") {
			managers = append(managers, "Cargo")
		}
	}

	return managers
}

func (s *AnalyzerService) detectPatterns(analysis *models.RepositoryAnalysis) {
	// Detect architectural patterns based on directory structure
	hasHandlers := containsAny(analysis.KeyDirectories, []string{"handlers", "controllers", "routes"})
	hasServices := containsAny(analysis.KeyDirectories, []string{"services", "business"})
	hasModels := containsAny(analysis.KeyDirectories, []string{"models", "entities"})
	hasRepositories := containsAny(analysis.KeyDirectories, []string{"repositories", "data"})

	if hasHandlers && hasServices && hasModels && hasRepositories {
		analysis.Patterns["architecture"] = "Clean Architecture (Handlers -> Services -> Repositories -> Models)"
	} else if hasHandlers && hasServices {
		analysis.Patterns["architecture"] = "Layered Architecture"
	} else if hasHandlers {
		analysis.Patterns["architecture"] = "MVC-like"
	}

	// Detect API style
	if containsAny(analysis.KeyDirectories, []string{"api", "routes", "handlers"}) {
		analysis.Patterns["api_style"] = "RESTful API"
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func containsAny(slice []string, items []string) bool {
	for _, dir := range slice {
		for _, item := range items {
			if strings.Contains(strings.ToLower(dir), strings.ToLower(item)) {
				return true
			}
		}
	}
	return false
}
