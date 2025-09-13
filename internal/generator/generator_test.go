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

func TestGenerator(t *testing.T) {
	// Setup a temporary project structure
	tempDir, err := os.MkdirTemp("", "generator-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	cfg := &config.Config{
		ProjectRoot: tempDir,
		AssetsDir:   filepath.Join(tempDir, "assets"),
		PagesDir:    filepath.Join(tempDir, "pages"),
		TemplateDir: filepath.Join(tempDir, "templates"),
	}

	// Create directories
	err = os.MkdirAll(cfg.AssetsDir, 0755)
	require.NoError(t, err)
	err = os.MkdirAll(cfg.PagesDir, 0755)
	require.NoError(t, err)
	err = os.MkdirAll(cfg.TemplateDir, 0755)
	require.NoError(t, err)

	// Create a test template
	templateContent := `Template content for {{ .CurrentPath }}`
	templatePath := filepath.Join(cfg.TemplateDir, "test.template")
	err = os.WriteFile(templatePath, []byte(templateContent), 0644)
	require.NoError(t, err)

	// Create a test index.ini
	iniDir := filepath.Join(cfg.AssetsDir, "test_category")
	err = os.Mkdir(iniDir, 0755)
	require.NoError(t, err)

	iniContent := `
[header]
template = test
[properties]
key1 = value1
key2 = value2
`
	iniPath := filepath.Join(iniDir, "index.ini")
	err = os.WriteFile(iniPath, []byte(iniContent), 0644)
	require.NoError(t, err)

	// --- Test Build ---
t.Run("builds pages from ini files", func(t *testing.T) {
		gen := generator.New(cfg)
		err := gen.Build()
		require.NoError(t, err)

		// Check for the output file
		outputFile := filepath.Join(cfg.PagesDir, "test_category.md")
		assert.FileExists(t, outputFile)

		// Check the content of the output file
		content, err := os.ReadFile(outputFile)
		require.NoError(t, err)

		// Normalize newlines for comparison
		expectedContent := strings.Join([]string{
			"generated:: true",
			"key1:: value1",
			"key2:: value2",
			"",
			"Template content for test_category",
		}, "\n")

		assert.Equal(t, expectedContent, strings.TrimSpace(strings.ReplaceAll(string(content), "\r\n", "\n")))
	})

	// --- Test Clear ---
t.Run("clears generated files", func(t *testing.T) {
		// Ensure the file exists before clearing
		outputFile := filepath.Join(cfg.PagesDir, "test_category.md")
		require.FileExists(t, outputFile, "Test setup failed, file to be cleared does not exist")

		gen := generator.New(cfg)
		err := gen.Clear()
		require.NoError(t, err)

		// Check that the file was removed
		assert.NoFileExists(t, outputFile)
	})
}
