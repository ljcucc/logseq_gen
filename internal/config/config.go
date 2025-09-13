package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/ini.v1"
)

const (
	// DefaultAssetsDir is the default directory for input assets.
	DefaultAssetsDir = "assets"
	// DefaultPagesDir is the default directory for output pages.
	DefaultPagesDir = "pages"
	// DefaultTemplateDir is the default directory for templates.
	DefaultTemplateDir = "templates"
)

// Config holds the application configuration.
type Config struct {
	AssetsDir   string
	PagesDir    string
	TemplateDir string
	ProjectRoot string
}

// Load finds and loads the configuration from a generate.ini file.
// It starts searching from the current working directory and goes up.
// If not found, it returns a default configuration.
func Load() (*Config, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}

	projectRoot, err := findProjectRoot(wd)
	if err != nil {
		log.Printf("generate.ini not found, using defaults: %v", err)
		return &Config{
			AssetsDir:   DefaultAssetsDir,
			PagesDir:    DefaultPagesDir,
			TemplateDir: DefaultTemplateDir,
			ProjectRoot: wd,
		}, nil
	}

	iniPath := filepath.Join(projectRoot, "generate.ini")
	cfg, err := ini.Load(iniPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load %s: %w", iniPath, err)
	}

	inputPath := cfg.Section("input").Key("path").String()
	outputPath := cfg.Section("output").Key("path").String()
	templatePath := cfg.Section("template").Key("path").String()

	if inputPath == "" || outputPath == "" || templatePath == "" {
		return nil, fmt.Errorf("input.path, output.path, or template.path not set in %s", iniPath)
	}

	return &Config{
		AssetsDir:   filepath.Join(projectRoot, inputPath),
		PagesDir:    filepath.Join(projectRoot, outputPath),
		TemplateDir: filepath.Join(projectRoot, templatePath),
		ProjectRoot: projectRoot,
	}, nil
}

// findProjectRoot searches recursively for generate.ini to find the project root.
func findProjectRoot(startPath string) (string, error) {
	currentPath, err := filepath.Abs(startPath)
	if err != nil {
		return "", err
	}

	for {
		iniPath := filepath.Join(currentPath, "generate.ini")
		if _, err := os.Stat(iniPath); err == nil {
			return currentPath, nil
		}

		parent := filepath.Dir(currentPath)
		if parent == currentPath { // Reached the filesystem root
			break
		}
		currentPath = parent
	}

	return "", fmt.Errorf("generate.ini not found in any parent directory")
}
