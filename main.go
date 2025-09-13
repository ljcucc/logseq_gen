package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"gopkg.in/ini.v1"
)

const generatedMarker = "generated:: true"

// Config holds the configuration from the generate.ini file.
type Config struct {
	AssetsDir   string
	PagesDir    string
	TemplateDir string
}

// findProjectRoot searches recursively for generate.ini to find the project root.
func findProjectRoot(startPath string) (string, error) {
	var rootPath string
	err := filepath.WalkDir(startPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Name() == "generate.ini" {
			rootPath = filepath.Dir(path)
			return fs.SkipDir
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if rootPath == "" {
		return "", fmt.Errorf("generate.ini not found")
	}
	return rootPath, nil
}

// LoadConfig loads configuration from generate.ini.
func LoadConfig() (*Config, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}

	projectRoot, err := findProjectRoot(wd)
	if err != nil {
		log.Printf("generate.ini not found, using defaults: %v", err)
		return &Config{AssetsDir: "assets", PagesDir: "pages", TemplateDir: "templates"}, nil
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
	}, nil
}

// Generator manages the file generation process.
type Generator struct {
	Config       *Config
	templateCache map[string]*template.Template
}

// NewGenerator creates a new Generator.
func NewGenerator(cfg *Config) *Generator {
	return &Generator{
		Config:       cfg,
		templateCache: make(map[string]*template.Template),
	}
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

// Clear removes generated files from the pages directory.
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

// Build generates markdown pages from index.ini files.
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

// findIniFiles finds all index.ini files in the assets directory.
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

// processIniFile processes a single index.ini file to generate a page.
func (g *Generator) processIniFile(iniPath string) {
	fmt.Printf("Processing: %s\n", iniPath)
	cfg, err := ini.Load(iniPath)
	if err != nil {
		log.Printf("[SKIP] Could not load %s: %v", iniPath, err)
		return
	}

	var outputContent strings.Builder
	templateName := cfg.Section("header").Key("template").String()

	if templateName != "" {
		tmpl, err := g.getTemplate(templateName)
		if err != nil {
			log.Printf("[SKIP] Could not get template %s: %v", templateName, err)
			return
		}

		relPath, err := filepath.Rel(g.Config.AssetsDir, filepath.Dir(iniPath))
		if err != nil {
			log.Printf("[SKIP] Could not get relative path for %s: %v", iniPath, err)
			return
		}

		currentPath := strings.ReplaceAll(relPath, string(os.PathSeparator), "/")

		data := struct {
			CurrentPath string
			Properties  map[string]string
		}{
			CurrentPath: currentPath,
			Properties:  cfg.Section("properties").KeysHash(),
		}

		var renderedTemplate bytes.Buffer
		if err := tmpl.Execute(&renderedTemplate, data); err != nil {
			log.Printf("[SKIP] Could not execute template for %s: %v", iniPath, err)
			return
		}
		outputContent.WriteString(renderedTemplate.String())

	} else {
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

	finalContent := generatedMarker + "\n" + outputContent.String()
	if err := os.WriteFile(outputFilepath, []byte(finalContent), 0644); err != nil {
		log.Printf("[SKIP] Could not write file %s: %v", outputFilepath, err)
		return
	}
	fmt.Printf("-> Generated %s\n", outputFilepath)
}


// getTemplate retrieves a template from cache or parses it from file.
func (g *Generator) getTemplate(name string) (*template.Template, error) {
	if tmpl, ok := g.templateCache[name]; ok {
		return tmpl, nil
	}

	templateFile := filepath.Join(g.Config.TemplateDir, fmt.Sprintf("%s.template", name))
	content, err := os.ReadFile(templateFile)
	if err != nil {
		return nil, fmt.Errorf("could not read template file %s: %w", templateFile, err)
	}

	tmpl, err := template.New(name).Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("could not parse template %s: %w", name, err)
	}

	g.templateCache[name] = tmpl
	return tmpl, nil
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
