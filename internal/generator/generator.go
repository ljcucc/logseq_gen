package generator

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

	"logseq_gen/internal/config"
	"logseq_gen/internal/schema"
)

const generatedMarker = "generated:: true"

// Generator manages the file generation process.
type Generator struct {
	config        *config.Config
	templateCache map[string]*template.Template
	schemaCache   map[string]*schema.Schema
}

// New creates a new Generator.
func New(cfg *config.Config) *Generator {
	return &Generator{
		config:        cfg,
		templateCache: make(map[string]*template.Template),
		schemaCache:   make(map[string]*schema.Schema),
	}
}

// Build generates markdown pages from index.ini files.
func (g *Generator) Build() error {
	if err := g.Clear(); err != nil {
		return err
	}
	if err := os.MkdirAll(g.config.PagesDir, 0755); err != nil {
		return fmt.Errorf("could not create pages directory: %w", err)
	}

	fmt.Printf("\nStarting build process from %s...\n", g.config.AssetsDir)
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

// Clear removes generated files from the pages directory.
func (g *Generator) Clear() error {
	if _, err := os.Stat(g.config.PagesDir); os.IsNotExist(err) {
		fmt.Println("Pages directory does not exist. Nothing to clear.")
		return nil
	}

	fmt.Printf("Clearing generated files from %s...\n", g.config.PagesDir)
	files, err := filepath.Glob(filepath.Join(g.config.PagesDir, "*.md"))
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

// findIniFiles finds all index.ini files in the assets directory.
func (g *Generator) findIniFiles() ([]string, error) {
	var iniFiles []string
	err := filepath.Walk(g.config.AssetsDir, func(path string, info fs.FileInfo, err error) error {
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

	if shouldSkip := g.processFile(iniPath, cfg, &outputContent); shouldSkip {
		return
	}

	relPath, err := filepath.Rel(g.config.AssetsDir, filepath.Dir(iniPath))
	if err != nil {
		log.Printf("[SKIP] Could not determine relative path for %s: %v", iniPath, err)
		return
	}


	outputFilenameBase := strings.ReplaceAll(relPath, string(os.PathSeparator), "___")
	if outputFilenameBase == "." {
		outputFilenameBase = "index"
	}
	outputFilepath := filepath.Join(g.config.PagesDir, fmt.Sprintf("%s.md", outputFilenameBase))

	finalContent := generatedMarker + "\n" + outputContent.String()
	if err := os.WriteFile(outputFilepath, []byte(finalContent), 0644); err != nil {
		log.Printf("[SKIP] Could not write file %s: %v", outputFilepath, err)
		return
	}
	fmt.Printf("-> Generated %s\n", outputFilepath)
}

func (g *Generator) processWithTemplate(iniPath string, cfg *ini.File, templateName string, props map[string]string, outputContent *strings.Builder) {
	// Then, process the template
	tmpl, err := g.getTemplate(templateName)
	if err != nil {
		log.Printf("[SKIP] Could not get template %s: %v", templateName, err)
		return
	}

	relPath, err := filepath.Rel(g.config.AssetsDir, filepath.Dir(iniPath))
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
		Properties:  props,
	}

	var renderedTemplate bytes.Buffer
	if err := tmpl.Execute(&renderedTemplate, data); err != nil {
		log.Printf("[SKIP] Could not execute template for %s: %v", iniPath, err)
		return
	}
	outputContent.WriteString(renderedTemplate.String())
}

func (g *Generator) processFile(iniPath string, cfg *ini.File, outputContent *strings.Builder) (shouldSkip bool) {
	propertiesSection := cfg.Section("properties")
	orderedKeys := propertiesSection.KeyStrings()
	props := make(map[string]string)
	for _, key := range orderedKeys {
		props[key] = propertiesSection.Key(key).String()
	}

	headerSection := cfg.Section("header")

	if headerSection.HasKey("schema") {
		schemaName := headerSection.Key("schema").String()
		s, err := g.getSchema(schemaName)
		if err != nil {
			log.Printf("[SKIP] Schema '%s' not found or invalid: %v", schemaName, err)
			return true
		}

		transformedProps, err := s.ValidateAndTransform(props)
		if err != nil {
			log.Printf("[SKIP] Validation failed for %s: %v", iniPath, err)
			return true
		}
		props = transformedProps
	}

	for _, key := range orderedKeys {
		if value, ok := props[key]; ok {
			outputContent.WriteString(fmt.Sprintf("%s:: %s\n", key, value))
			delete(props, key) // Remove the key to handle remaining new properties
		}
	}

	// Append any new properties added by the schema (e.g., defaults)
	for key, value := range props {
		outputContent.WriteString(fmt.Sprintf("%s:: %s\n", key, value))
	}
	outputContent.WriteString("\n")

	if headerSection.HasKey("template") {
		templateName := headerSection.Key("template").String()
		g.processWithTemplate(iniPath, cfg, templateName, props, outputContent)
	} else if headerSection.HasKey("content") {
		contentFilename := strings.Trim(headerSection.Key("content").String(), "\"")
		contentFilepath := filepath.Join(filepath.Dir(iniPath), contentFilename)
		if _, err := os.Stat(contentFilepath); os.IsNotExist(err) {
			log.Printf("[SKIP] Content file '%s' not found.", contentFilepath)
			return true
		}
		content, err := os.ReadFile(contentFilepath)
		if err != nil {
			log.Printf("[SKIP] Could not read content file %s: %v", contentFilepath, err)
			return true
		}
		outputContent.Write(content)
	}
	return false
}

// getTemplate retrieves a template from cache or parses it from file.
func (g *Generator) getTemplate(name string) (*template.Template, error) {
	if tmpl, ok := g.templateCache[name]; ok {
		return tmpl, nil
	}

	templateFile := filepath.Join(g.config.TemplateDir, fmt.Sprintf("%s.template", name))
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

// getSchema retrieves a schema from cache or loads it from file.
func (g *Generator) getSchema(name string) (*schema.Schema, error) {
	if s, ok := g.schemaCache[name]; ok {
		return s, nil
	}

	schemaFile := filepath.Join(g.config.SchemaDir, fmt.Sprintf("%s.yaml", name))
	if _, err := os.Stat(schemaFile); os.IsNotExist(err) {
		schemaFile = filepath.Join(g.config.SchemaDir, fmt.Sprintf("%s.json", name))
	}

	s, err := schema.LoadSchema(schemaFile)
	if err != nil {
		return nil, err
	}

	g.schemaCache[name] = s
	return s, nil
}
