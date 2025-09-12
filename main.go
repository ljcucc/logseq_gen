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

const (
	generatedMarker = "generated:: true"
	assetsDir       = "assets"
	pagesDir        = "pages"
)

func isGeneratedFile(path string) (bool, error) {
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

func findIniFiles() ([]string, error) {
	var iniFiles []string
	err := filepath.Walk(assetsDir, func(path string, info fs.FileInfo, err error) error {
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

func clearBuild() {
	if _, err := os.Stat(pagesDir); os.IsNotExist(err) {
		fmt.Println("Pages directory does not exist. Nothing to clear.")
		return
	}

	fmt.Printf("Clearing generated files from %s...\n", pagesDir)
	files, err := filepath.Glob(filepath.Join(pagesDir, "*.md"))
	if err != nil {
		log.Printf("Error finding markdown files: %v", err)
		return
	}

	for _, file := range files {
		generated, err := isGeneratedFile(file)
		if err != nil {
			log.Printf("Error checking if file is generated: %v", err)
			continue
		}
		if generated {
			err := os.Remove(file)
			if err != nil {
				log.Printf("Error removing file %s: %v", file, err)
			} else {
				fmt.Printf("Removed %s\n", filepath.Base(file))
			}
		}
	}
	fmt.Println("Clear build finished.")
}

func processIniFile(iniPath string) {
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
		contentFilename := strings.Trim(headerSection.Key("content").String(), "")
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

	relPath, err := filepath.Rel(assetsDir, filepath.Dir(iniPath))
	if err != nil {
		log.Printf("[SKIP] Could not determine relative path for %s: %v", iniPath, err)
		return
	}

	outputFilenameBase := "index"
	if relPath != "." {
		outputFilenameBase = strings.ReplaceAll(relPath, string(os.PathSeparator), "___")
	}
	outputFilepath := filepath.Join(pagesDir, fmt.Sprintf("%s.md", outputFilenameBase))

	err = os.WriteFile(outputFilepath, []byte(outputContent.String()), 0644)
	if err != nil {
		log.Printf("[SKIP] Could not write file %s: %v", outputFilepath, err)
		return
	}
	fmt.Printf("-> Generated %s\n", outputFilepath)
}

func buildMarkdown() {
	clearBuild()
	if err := os.MkdirAll(pagesDir, 0755); err != nil {
		log.Fatalf("Could not create pages directory: %v", err)
	}

	fmt.Printf("\nStarting build process from %s...\n", assetsDir)
	iniFiles, err := findIniFiles()
	if err != nil {
		log.Fatalf("Error finding ini files: %v", err)
	}

	for _, iniPath := range iniFiles {
		processIniFile(iniPath)
	}
	fmt.Println("\nBuild process finished.")
}

func main() {
	command := "build"
	if len(os.Args) > 1 {
		command = strings.ToLower(os.Args[1])
	}

	switch command {
	case "build":
		buildMarkdown()
	case "clear":
		clearBuild()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		fmt.Fprintf(os.Stderr, "Usage: %s [build|clear]\n", os.Args[0])
		os.Exit(1)
	}
}
