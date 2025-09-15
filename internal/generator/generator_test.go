package generator_test

import (
	"logseq_gen/internal/config"
	"logseq_gen/internal/generator"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerator_Build(t *testing.T) {
	// Setup a temporary project structure
	tempDir, err := os.MkdirTemp("", "generator-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	cfg := &config.Config{
		ProjectRoot: tempDir,
		AssetsDir:   filepath.Join(tempDir, "assets"),
		PagesDir:    filepath.Join(tempDir, "pages"),
		TemplateDir: filepath.Join(tempDir, "templates"),
		SchemaDir:   filepath.Join(tempDir, "schemas"),
	}

	// Create directories
	require.NoError(t, os.MkdirAll(cfg.AssetsDir, 0755))
	require.NoError(t, os.MkdirAll(cfg.PagesDir, 0755))
	require.NoError(t, os.MkdirAll(cfg.TemplateDir, 0755))
	require.NoError(t, os.MkdirAll(cfg.SchemaDir, 0755))

	// Create a test template
	templateContent := `Template content for {{ .CurrentPath }}`
	templatePath := filepath.Join(cfg.TemplateDir, "test.template")
	require.NoError(t, os.WriteFile(templatePath, []byte(templateContent), 0644))

	// Create a test schema
	schemaContent := `
version: 1
types:
  property_a:
    required: true
    type: string
  property_b:
    type: enum
    keys:
      key1:
        display: Value 1
`
	schemaPath := filepath.Join(cfg.SchemaDir, "test_schema.yaml")
	require.NoError(t, os.WriteFile(schemaPath, []byte(schemaContent), 0644))

	// Create a valid test index.ini
	validIniDir := filepath.Join(cfg.AssetsDir, "valid_test")
	require.NoError(t, os.MkdirAll(validIniDir, 0755))
	validIniContent := `
[header]
schema = test_schema
template = test
[properties]
property_a = hello
property_b = key1
`
	validIniPath := filepath.Join(validIniDir, "index.ini")
	require.NoError(t, os.WriteFile(validIniPath, []byte(validIniContent), 0644))

	// Create an invalid test index.ini (missing required property_a)
	invalidIniDir := filepath.Join(cfg.AssetsDir, "invalid_test")
	require.NoError(t, os.MkdirAll(invalidIniDir, 0755))
	invalidIniContent := `
[header]
schema = test_schema
template = test
[properties]
property_b = key1
`
	invalidIniPath := filepath.Join(invalidIniDir, "index.ini")
	require.NoError(t, os.WriteFile(invalidIniPath, []byte(invalidIniContent), 0644))

	// Run the generator
	gen := generator.New(cfg)
	err = gen.Build()
	require.NoError(t, err)

	// ---
	// Assertions
	// ---

	// 1. Check that the valid file was created with correct content
	outputFileValid := filepath.Join(cfg.PagesDir, "valid_test.md")
	assert.FileExists(t, outputFileValid)

	content, err := os.ReadFile(outputFileValid)
	require.NoError(t, err)

	expectedContent := strings.Join([]string{
		"generated:: true",
		"property_a:: hello",
		"property_b:: [[property_b/Value 1]]",
		"",
		"Template content for valid_test",
	}, "\n")
	assert.Equal(t, expectedContent, strings.TrimSpace(strings.ReplaceAll(string(content), "\r\n", "\n")))

	// 2. Check that the invalid file was NOT created
	outputFileInvalid := filepath.Join(cfg.PagesDir, "invalid_test.md")
	assert.NoFileExists(t, outputFileInvalid)
}