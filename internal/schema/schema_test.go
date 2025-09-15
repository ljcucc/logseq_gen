package schema

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSchema(t *testing.T) {
	// Create a temporary schema file for testing
	schemaContent := `
version: 1
types:
  property_a:
    required: true
    type: number
  property_b:
    required: false
    type: number
  property_c:
    required: false
    type: string
    default: 3c
  property_d:
    required: true
    type: boolean
    default: false
  property_e:
    required: true
    type: enum
    keys:
      num_1:
        display: Number 1
      num_2:
        display: Number 2
  property_f:
    type: date
`
	schemaDir := t.TempDir()
	schemaPath := filepath.Join(schemaDir, "test_schema.yaml")
	err := os.WriteFile(schemaPath, []byte(schemaContent), 0644)
	assert.NoError(t, err)

	// Load the schema
	schema, err := LoadSchema(schemaPath)
	assert.NoError(t, err)
	assert.NotNil(t, schema)

	t.Run("Valid record", func(t *testing.T) {
		record := map[string]string{
			"property_a": "123",
			"property_b": "456.7",
			"property_e": "num_1",
			"property_f": "2025-09-15",
		}

		transformed, err := schema.ValidateAndTransform(record)
		assert.NoError(t, err)
		assert.Equal(t, "123", transformed["property_a"])
		assert.Equal(t, "456.7", transformed["property_b"])
		assert.Equal(t, "3c", transformed["property_c"]) // from default
		assert.Equal(t, "false", transformed["property_d"]) // from default
		assert.Equal(t, "[[property_e/Number 1]]", transformed["property_e"]) // from enum
		assert.Equal(t, "[[2025-09-15]]", transformed["property_f"])      // from date
	})

	t.Run("Missing required property", func(t *testing.T) {
		record := map[string]string{
			"property_b": "456.7",
			"property_e": "num_1",
		}

		_, err := schema.ValidateAndTransform(record)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "required property 'property_a' is missing")
	})

	t.Run("Invalid number", func(t *testing.T) {
		record := map[string]string{
			"property_a": "abc",
			"property_e": "num_1",
		}

		_, err := schema.ValidateAndTransform(record)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "property 'property_a' with value 'abc' is not a valid number")
	})

	t.Run("Invalid boolean", func(t *testing.T) {
		record := map[string]string{
			"property_a": "123",
			"property_d": "not-a-bool",
			"property_e": "num_1",
		}

		_, err := schema.ValidateAndTransform(record)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "property 'property_d' with value 'not-a-bool' is not a valid boolean")
	})

	t.Run("Invalid enum key", func(t *testing.T) {
		record := map[string]string{
			"property_a": "123",
			"property_e": "num_999", // invalid key
		}

		_, err := schema.ValidateAndTransform(record)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "property 'property_e' with value 'num_999' is not a valid enum key")
	})

	t.Run("Invalid date", func(t *testing.T) {
		record := map[string]string{
			"property_a": "123",
			"property_e": "num_1",
			"property_f": "not-a-date",
		}

		_, err := schema.ValidateAndTransform(record)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "is not a valid date in YYYY-MM-DD format")
	})
}
