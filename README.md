# Logseq Generator

A universal generation tool for Logseq, creating Markdown files from structured assets, templates, and schemas.

## Description

This project provides a powerful generation tool that creates Markdown files in the `pages/` directory based on `.ini` files. It validates and transforms data using user-defined schemas, and then generates pages either from a template or by combining properties with a content file.

**How it works:**

The tool scans the `assets/` directory for `index.ini` files. For each file, it performs the following steps:

1.  **Schema Validation:** If the `index.ini` references a schema (e.g., `schema = my_schema`), the tool loads the corresponding schema from the `schemas/` directory.
2.  **Data Transformation:** It validates the data in the `[properties]` section against the schema. If the data is valid, it transforms the values based on the schema rules (e.g., formatting dates, replacing enum keys).
3.  **Generation:** If validation succeeds, it generates a Markdown file using either a template or direct content inclusion, similar to the basic functionality.

If validation fails at any step, the file is skipped, and an error is logged.

## Features

*   Automated Markdown generation from structured asset files.
*   **Schema-based validation and transformation** of data.
*   Support for data types: `string`, `number`, `boolean`, `enum`, `link`, and `date`.
*   Configurable default values and required fields.
*   Dual generation modes: template-based or direct.
*   Clear build functionality to remove previously generated files.

## Installation

```bash
git clone https://github.com/ljcucc/logseq_gen.git
cd logseq_gen
```

## Usage

First, ensure you have Go installed on your system.

To build the Markdown files:
```bash
go run main.go build
```

To clear any previously generated files:
```bash
go run main.go clear
```

## Configuration

The tool is configured via a `generate.ini` file at the project root.

### `generate.ini`

This file defines the main paths for the generator.

```ini
[input]
path=./assets

[output]
path=./pages

[template]
path=./templates

[schema]
path=./schemas
```

*   `input.path`: The directory containing your asset structure.
*   `output.path`: The directory where the Markdown pages will be generated.
*   `template.path`: The directory containing your `.template` files.
*   `schema.path`: The directory containing your schema definition files (`.yaml` or `.json`).

---

## Schemas

Schemas are the core of the validation and transformation system. They are defined in YAML or JSON files and placed in the directory specified by `schema.path`.

### `index.ini` Schema Reference

To apply a schema to an `index.ini` file, add a `schema` key to its `[header]` section. The value should be the name of the schema file (without the extension).

`assets/xxx/yyy/aaa/index.ini`:
```ini
[header]
schema = bbb
template = example_template

[properties]
property_a = 123
property_e = num_1
property_f = 2025-09-15
```

### Schema Definition

A schema file defines the types, rules, and transformations for properties.

`schemas/bbb.yaml`:
```yaml
version: 1
types:
  # Required, no default
  property_a:
    required: true
    type: number

  # Not required
  property_b:
    required: false
    type: number

  # Has a default value, so `required` is ignored
  property_c:
    type: string
    default: 3c

  # Enum type with key-value replacement
  property_e:
    required: true
    type: enum
    keys:
      num_1:
        display: Number 1
      num_2:
        display: Number 2

  # Date type, expects YYYY-MM-DD
  property_f:
    type: date
```

### Property Types and Rules

Each entry under `types` defines a property key.

*   `required` (`bool`): If `true`, the property must exist in the `index.ini` file. This is ignored if a `default` value is provided.
*   `default` (`any`): A fallback value to use if the property is not present.
*   `type` (`string`): The data type of the property. This determines the validation and transformation rules.

#### Supported Types

| Type      | Validation                               | Transformation Output Example                               |
| :-------- | :--------------------------------------- | :---------------------------------------------------------- |
| `string`  | None.                                    | `some string`                                               |
| `number`  | Must be a valid number.                  | `123.45`                                                    |
| `boolean` | Must be a valid boolean (`true`, `false`). | `true`                                                      |
| `enum`    | Value must be a key defined in `keys`.   | `[[property_name/Display Value]]` (e.g., `[[property_e/Number 1]]`) |
| `date`    | Must be in `YYYY-MM-DD` format.          | `[[YYYY-MM-DD]]` (e.g., `[[2025-09-15]]`)                    |

#### Enum Keys

For the `enum` type, you must provide a `keys` map. Each entry in the map represents a valid input value and its corresponding `display` value for transformation.

```yaml
property_e:
  type: enum
  keys:
    num_1: # Input value
      display: Number 1 # Output display value
```

## Generation Methods

Generation proceeds only after successful schema validation.

### 1. Template-Based Generation

The transformed properties are available in the `.Properties` map within the template.

`templates/example_template.template`:
```
- Property A: {{ .Properties.property_a }}
- Property E: {{ .Properties.property_e }}
- Property F: {{ .Properties.property_f }}
```

### 2. Direct Generation

The transformed key-value pairs are listed at the top of the generated file.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
