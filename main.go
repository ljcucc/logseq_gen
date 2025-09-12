package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/ini.v1"
)

const generatedMarker = "generated:: true"

// Config holds the configuration from the generate.ini file.
// It specifies the input directory for assets and the output directory for pages.
type Config struct {
	AssetsDir string
	PagesDir  string
}

// findProjectRoot searches recursively from the given path for a directory containing
// "generate.ini" and returns that directory's path as the project root.
func findProjectRoot(startPath string) (string, error) {
	var rootPath string
	err := filepath.WalkDir(startPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Name() == "generate.ini" {
			rootPath = filepath.Dir(path)
			return fs.SkipDir // Stop searching once we find the first one
		}
		return nil
	})

	if err != nil {
		return "", err
	}
	if rootPath == "" {
		return "", fmt.Errorf("generate.ini not found in any subdirectory")
	}
	return rootPath, nil
}

// LoadConfig finds and loads the configuration from the generate.ini file.
// It determines the project root and resolves the asset and page paths.
func LoadConfig() (*Config, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}

	projectRoot, err := findProjectRoot(wd)
	if err != nil {
		// Fallback for backward compatibility if generate.ini is not found
		log.Println("generate.ini not found, falling back to default paths ('assets', 'pages')")
		return &Config{AssetsDir: "assets", PagesDir: "pages"}, nil
	}

	iniPath := filepath.Join(projectRoot, "generate.ini")
	cfg, err := ini.Load(iniPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load %s: %w", iniPath, err)
	}

	inputPath := cfg.Section("input").Key("path").String()
	outputPath := cfg.Section("output").Key("path").String()

	if inputPath == "" || outputPath == "" {
		return nil, fmt.Errorf("input.path or output.path not set in %s", iniPath)
	}

	return &Config{
		AssetsDir: filepath.Join(projectRoot, inputPath),
		PagesDir:  filepath.Join(projectRoot, outputPath),
	}, nil
}
type Generator struct {
	Config *Config
}

// NewGenerator creates a new Generator instance.
func NewGenerator(cfg *Config) *Generator {
	return &Generator{Config: cfg}
}

// isGeneratedFile checks if a file is marked as generated.
func (g *Generator) isGeneratedFile(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text()) == generatedMarker, nil
	}

	return false, scanner.Err()
}

// Clear removes all generated markdown files from the pages directory.
func (g *Generator) Clear() error {
	if _, err := os.Stat(g.Config.PagesDir); os.IsNotExist(err) {
		fmt.Println("Pages directory does not exist. Nothing to clear.")
		return nil
	}

	fmt.Printf("Clearing generated files from %s...\n", g.Config.PagesDir)
	files, err := filepath.Glob(filepath.Join(g.Config.PagesDir, "*.md"))
	if err != nil {
		return fmt.Errorf("error finding markdown files: %w", err)
	}

	for _, file := range files {
		generated, err := g.isGeneratedFile(file)
		if err != nil {
			log.Printf("Error checking if file %s is generated: %v", file, err)
			continue
		}
		if generated {
			if err := os.Remove(file); err != nil {
				log.Printf("Error removing file %s: %v", file, err)
			} else {
				fmt.Printf("Removed %s\n", filepath.Base(file))
			}
		}
	}
	fmt.Println("Clear build finished.")
	return nil
}

// Build finds all index.ini files and generates corresponding markdown pages.
func (g *Generator) Build() error {
	if err := g.Clear(); err != nil {
		return err
	}
	if err := os.MkdirAll(g.Config.PagesDir, 0755); err != nil {
		return fmt.Errorf("could not create pages directory: %w", err)
	}

	fmt.Printf("\nStarting build process from %s...\n", g.Config.AssetsDir)
	iniFiles, err := g.findIniFiles()
	if err != nil {
		return fmt.Errorf("error finding ini files: %w", err)
	}

	for _, iniPath := range iniFiles {
		g.processIniFile(iniPath)
	}
	fmt.Println("\nBuild process finished.")
	return nil
}

func (g *Generator) findIniFiles() ([]string, error) {
	var iniFiles []string
	err := filepath.Walk(g.Config.AssetsDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && info.Name() == "index.ini" {
			iniFiles = append(iniFiles, path)
		}
		return nil
	})
	return iniFiles, err
}

func (g *Generator) processIniFile(iniPath string) {
	fmt.Printf("Processing: %s\n", iniPath)
	cfg, err := ini.Load(iniPath)
	if err != nil {
		log.Printf("[SKIP] Could not process %s: %v", iniPath, err)
		return
	}

	var outputContent strings.Builder
	outputContent.WriteString(generatedMarker + "\n")

	propsSection := cfg.Section("properties")
	for _, key := range propsSection.KeyStrings() {
		value := propsSection.Key(key).String()
		outputContent.WriteString(fmt.Sprintf("%s:: %s\n", key, value))
	}
	outputContent.WriteString("\n")

	headerSection := cfg.Section("header")
	if headerSection.HasKey("content") {
		contentFilename := strings.Trim(headerSection.Key("content").String(), "\"")
		contentFilepath := filepath.Join(filepath.Dir(iniPath), contentFilename)
		if _, err := os.Stat(contentFilepath); os.IsNotExist(err) {
			log.Printf("[SKIP] Content file '%s' not found.", contentFilepath)
			return
		}
		content, err := os.ReadFile(contentFilepath)
		if err != nil {
			log.Printf("[SKIP] Could not read content file %s: %v", contentFilepath, err)
			return
		}
		outputContent.Write(content)
	}

	relPath, err := filepath.Rel(g.Config.AssetsDir, filepath.Dir(iniPath))
	if err != nil {
		log.Printf("[SKIP] Could not determine relative path for %s: %v", iniPath, err)
		return
	}

	outputFilenameBase := "index"
	if relPath != "." {
		outputFilenameBase = strings.ReplaceAll(relPath, string(os.PathSeparator), "___")
	}
	outputFilepath := filepath.Join(g.Config.PagesDir, fmt.Sprintf("%s.md", outputFilenameBase))

	if err := os.WriteFile(outputFilepath, []byte(outputContent.String()), 0644); err != nil {
		log.Printf("[SKIP] Could not write file %s: %v", outputFilepath, err)
		return
	}
	fmt.Printf("-> Generated %s\n", outputFilepath)
}

// Run executes the command-line interface.
func (g *Generator) Run(args []string) error {
	command := "build"
	if len(args) > 1 {
		command = strings.ToLower(args[1])
	}

	switch command {
	case "build":
		return g.Build()
	case "clear":
		return g.Clear()
	default:
		return fmt.Errorf("unknown command: %s\nUsage: %s [build|clear]", command, args[0])
	}
}

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	generator := NewGenerator(cfg)
	if err := generator.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}