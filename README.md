# Logseq Generator

A universal generation tool for Logseq, creating Markdown files from structured assets and templates.

## Description

This project provides a universal generation tool designed to create Markdown files in the `pages/` directory based on `.ini` files. Generation can be done in one of two ways: either from a template or directly from properties and a content file.

**How it works:**

The tool scans the `assets` directory for `index.ini` files. For each `index.ini` found, it checks for a `template` key in the `[header]` section.

*   **If a `template` is specified**, the tool will use the corresponding template file from the `templates` directory to generate the page.
*   **If no `template` is specified**, the tool will generate the page directly by listing the key-value pairs from the `[properties]` section and appending the content of a specified markdown file.

## Generation Methods

### 1. Template-Based Generation

This is the recommended method for complex or customized page structures.

**Example:**

`generate.ini`:
```ini
[input]
path=./assets

[output]
path=./pages

[template]
path=./templates
```

`assets/xxx/yyy/aaa/index.ini`:
```ini
[header]
template=example_template

[properties]
property_a=10
```

`templates/example_template.template`:
```
{{ "{{-" }} embed [[{{ .CurrentPath }}]] {{ "-}}" }}
```

This will generate `pages/xxx___yyy___aaa.md` with the following content:

```markdown
generated:: true
{{ embed [[xxx/yyy/aaa]] }}
```

### 2. Direct Generation (No Template)

This method is for simple pages that only require a list of properties and content from a markdown file.

**Example:**

`assets/xxx/yyy/zzz/index.ini`:
```ini
[header]
content="index.md"

[properties]
property_a=10
property_b=20
```

`assets/xxx/yyy/zzz/index.md`:
```markdown
- Some content
```

This will generate `pages/xxx___yyy___zzz.md` with the following content:

```markdown
generated:: true
property_a:: 10
property_b:: 20

- Some content
```

## Features

*   Automated Markdown generation from structured asset files.
*   Dual generation modes: template-based or direct.
*   Template-based generation uses Go's `text/template` engine.
*   Configuration via `generate.ini` to define input, output, and template directories.
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

The tool is configured via a `generate.ini` file. The script automatically searches for this file to determine the project root.

### `generate.ini`

This file defines the main paths for the generator.

```ini
[input]
path=./assets

[output]
path=./pages

[template]
path=./templates
```

*   `input.path`: The directory containing your asset structure.
*   `output.path`: The directory where the Markdown pages will be generated.
*   `template.path`: The directory containing your `.template` files.

### `index.ini`

These files define the generation method and provide the necessary data.

```ini
[header]
; template=your_template_name  ; For template-based generation
; content=your_content_file.md ; For direct generation

[properties]
key=value
```

*   `header.template`: **(Optional)** The name of the template file to use. If present, enables template-based generation.
*   `header.content`: **(Optional)** The name of the markdown file to include for direct generation. Used when `template` is not specified.
*   `properties`: A section for key-value pairs. In template mode, they are available under the `.Properties` map. In direct mode, they are listed at the top of the file.

### Templates

Templates are written using Go's `text/template` syntax. The following data is available to the templates:

*   `.CurrentPath`: The relative path of the `index.ini` file from the assets directory (e.g., `xxx/yyy/aaa`).
*   `.Properties`: A map of the key-value pairs from the `[properties]` section of the `index.ini` file.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
