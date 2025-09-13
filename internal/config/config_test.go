package config_test

import (
	"logseq_gen/internal/config"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	t.Run("finds and loads generate.ini", func(t *testing.T) {
		// Setup: Create a temporary directory structure
		tempDir, err := os.MkdirTemp("", "config-test-")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		subDir := filepath.Join(tempDir, "subdir")
		err = os.Mkdir(subDir, 0755)
		require.NoError(t, err)

		iniContent := `
[input]
path = my_assets
[output]
path = my_pages
[template]
path = my_templates
`
		iniPath := filepath.Join(tempDir, "generate.ini")
		err = os.WriteFile(iniPath, []byte(iniContent), 0644)
		require.NoError(t, err)

		// Change to the subdirectory to test the upward search
		originalWD, err := os.Getwd()
		require.NoError(t, err)
		err = os.Chdir(subDir)
		require.NoError(t, err)
		defer os.Chdir(originalWD)

		// Execute
		cfg, err := config.Load()

		// Assert
		require.NoError(t, err)

		expectedRoot, err := filepath.EvalSymlinks(tempDir)
		require.NoError(t, err)
		actualRoot, err := filepath.EvalSymlinks(cfg.ProjectRoot)
		require.NoError(t, err)

		assert.Equal(t, expectedRoot, actualRoot)
		assert.Equal(t, filepath.Join(expectedRoot, "my_assets"), cfg.AssetsDir)
		assert.Equal(t, filepath.Join(expectedRoot, "my_pages"), cfg.PagesDir)
		assert.Equal(t, filepath.Join(expectedRoot, "my_templates"), cfg.TemplateDir)
	})

	t.Run("returns defaults when generate.ini is not found", func(t *testing.T) {
		// Setup: Create a temporary directory with no generate.ini
		tempDir, err := os.MkdirTemp("", "config-test-no-ini-")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		originalWD, err := os.Getwd()
		require.NoError(t, err)
		err = os.Chdir(tempDir)
		require.NoError(t, err)
		defer os.Chdir(originalWD)

		// Execute
		cfg, err := config.Load()

		// Assert
		require.NoError(t, err)

		expectedRoot, err := filepath.EvalSymlinks(tempDir)
		require.NoError(t, err)
		actualRoot, err := filepath.EvalSymlinks(cfg.ProjectRoot)
		require.NoError(t, err)

		assert.Equal(t, expectedRoot, actualRoot)
		assert.Equal(t, config.DefaultAssetsDir, cfg.AssetsDir)
		assert.Equal(t, config.DefaultPagesDir, cfg.PagesDir)
		assert.Equal(t, config.DefaultTemplateDir, cfg.TemplateDir)
	})
}
