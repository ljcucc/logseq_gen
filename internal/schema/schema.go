package schema

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

// Schema represents the structure of a schema file.
type Schema struct {
	Version int             `yaml:"version"`
	Types   map[string]Type `yaml:"types"`
}

// Type represents the type definition for a property.
type Type struct {
	Required bool               `yaml:"required"`
	Type     string             `yaml:"type"`
	Default  interface{}        `yaml:"default"`
	Keys     map[string]EnumKey `yaml:"keys"`
	Schema   string             `yaml:"schema"`
}

// EnumKey represents a key in an enum.
type EnumKey struct {
	Display string `yaml:"display"`
}

// LoadSchema loads a schema from a YAML file.
func LoadSchema(path string) (*Schema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema file %s: %w", path, err)
	}

	var schema Schema
	err = yaml.Unmarshal(data, &schema)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal schema file %s: %w", path, err)
	}

	return &schema, nil
}

// ValidateAndTransform validates and transforms a record based on the schema.
func (s *Schema) ValidateAndTransform(record map[string]string) (map[string]string, error) {
	result := make(map[string]string)
	for key, value := range record {
		result[key] = value
	}

	for key, typeDef := range s.Types {
		value, exists := result[key]

		if !exists && typeDef.Default != nil {
			value = fmt.Sprintf("%v", typeDef.Default)
			result[key] = value
			exists = true
		}

		if typeDef.Required && !exists {
			return nil, fmt.Errorf("required property '%s' is missing", key)
		}

		if !exists {
			continue
		}

		switch typeDef.Type {
		case "number":
			if _, err := strconv.ParseFloat(value, 64); err != nil {
				return nil, fmt.Errorf("property '%s' with value '%s' is not a valid number", key, value)
			}
		case "boolean":
			if _, err := strconv.ParseBool(value); err != nil {
				return nil, fmt.Errorf("property '%s' with value '%s' is not a valid boolean", key, value)
			}
		case "string":
			// No validation needed for string type
		case "enum":
			if enumKey, ok := typeDef.Keys[value]; ok {
				result[key] = fmt.Sprintf("[[%s/%s]]", key, enumKey.Display)
			} else {
				return nil, fmt.Errorf("property '%s' with value '%s' is not a valid enum key", key, value)
			}
		case "link":
			// In a real-world scenario, you might want to validate the link format.
			// For now, we just check if it's a string.
		case "date":
			if _, err := time.Parse("2006-01-02", value); err != nil {
				return nil, fmt.Errorf("property '%s' with value '%s' is not a valid date in YYYY-MM-DD format", key, value)
			}
			result[key] = fmt.Sprintf("[[%s]]", value)
		default:
			return nil, fmt.Errorf("unknown type '%s' for property '%s'", typeDef.Type, key)
		}
	}

	return result, nil
}
